package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"net/http"
	"testing"
)

func Test_assertRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *model.AllFlagRequest
		wantErr *echo.HTTPError
		want    ffcontext.Context
	}{
		{
			name: "no request",
			req:  nil,
			wantErr: echo.NewHTTPError(
				http.StatusBadRequest,
				"assertRequest: impossible to find user in request"),
		},
		{
			name: "request without evaluation context and user",
			req:  &model.AllFlagRequest{User: nil, EvaluationContext: nil},
			wantErr: echo.NewHTTPError(
				http.StatusBadRequest,
				"assertRequest: impossible to find user in request"),
		},
		{
			name: "user without key",
			req:  &model.AllFlagRequest{User: nil},
			wantErr: echo.NewHTTPError(
				http.StatusBadRequest,
				"assertRequest: impossible to find user in request"),
		},
		{
			name: "user with User and EvaluationContext, empty key for evaluation context",
			req: &model.AllFlagRequest{
				User:              &model.UserRequest{Key: "my-key"},
				EvaluationContext: &model.EvaluationContextRequest{Key: ""},
			},
			wantErr: echo.NewHTTPError(
				http.StatusBadRequest,
				"empty key for evaluation context, impossible to retrieve flags"),
		},
		{
			name: "invalid user but valid evaluation context should pass",
			req: &model.AllFlagRequest{
				User:              &model.UserRequest{Key: ""},
				EvaluationContext: &model.EvaluationContextRequest{Key: "my-key"},
			},
			wantErr: nil,
		},
		{
			name: "valid evaluation context and no user",
			req: &model.AllFlagRequest{
				EvaluationContext: &model.EvaluationContextRequest{Key: "my-key"},
			},
			wantErr: nil,
		},
		{
			name: "valid user and no evluation context",
			req: &model.AllFlagRequest{
				User: &model.UserRequest{Key: "my-key"},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := assertRequest(tt.req)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
func Test_evaluationContextFromRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *model.AllFlagRequest
		wantErr error
		want    ffcontext.Context
	}{
		{
			name: "no request",
			req:  nil,
			wantErr: echo.NewHTTPError(
				http.StatusBadRequest,
				"evaluationContextFromRequest: impossible to convert the request, req nil"),
		},
		{
			name:    "user without key",
			req:     &model.AllFlagRequest{User: nil},
			wantErr: fmt.Errorf("userRequestToUser: impossible to convert user, userRequest nil"),
		},
		{
			name: "valid use-case with EvaluationContext",
			req: &model.AllFlagRequest{
				User: nil,
				EvaluationContext: &model.EvaluationContextRequest{
					Key: "key-1",
					Custom: map[string]interface{}{
						"anonymous":    false,
						"custom-field": true,
					},
				},
			},
			want: ffcontext.
				NewEvaluationContextBuilder("key-1").
				AddCustom("anonymous", false).
				AddCustom("custom-field", true).
				Build(),
		},
		{
			name: "valid use-case with User",
			req: &model.AllFlagRequest{
				User: &model.UserRequest{
					Key:       "key-1",
					Anonymous: false,
					Custom: map[string]interface{}{
						"custom-field": true,
					},
				},
			},
			want: ffcontext.NewEvaluationContextBuilder("key-1").
				AddCustom("anonymous", false).
				AddCustom("custom-field", true).
				Build(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluationContextFromRequest(tt.req)
			if err != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
