package controller

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
	"github.com/thomaspoignant/go-feature-flag/ffuser"
	"net/http"
	"testing"
)

func Test_assertRequest_nil_request(t *testing.T) {
	got := assertRequest(nil)
	want := echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	assert.Equal(t, got, want)
}

func Test_assertRequest_request_without_user(t *testing.T) {
	got := assertRequest(&model.AllFlagRequest{User: nil})
	want := echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	assert.Equal(t, want, got)
}
func Test_assertRequest_request_without_context(t *testing.T) {
	got := assertRequest(&model.AllFlagRequest{EvaluationContext: nil})
	want := echo.NewHTTPError(http.StatusBadRequest, "assertRequest: impossible to find user in request")
	assert.Equal(t, want, got)
}

func Test_userRequestToUser_nil_user(t *testing.T) {
	_, got := userRequestToUser(nil)
	want := fmt.Errorf("userRequestToUser: impossible to convert user, userRequest nil")
	assert.Equal(t, want, got)
}

func Test_evaluationContextFromRequest_valid_use_case(t *testing.T) {
	req := &model.AllFlagRequest{
		User: nil,
		EvaluationContext: &model.EvaluationContextRequest{
			Key: "key-1",
			Custom: map[string]interface{}{
				"anonymous": false,
			},
		},
	}
	got, err := evaluationContextFromRequest(req)
	want := ffcontext.NewEvaluationContextBuilder("key-1").AddCustom("anonymous", false).Build()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_evaluationContextFromRequest_user_entry(t *testing.T) {
	req := &model.AllFlagRequest{
		User: &model.UserRequest{
			Key:       "key-1",
			Anonymous: false,
		},
	}
	got, err := evaluationContextFromRequest(req)
	want := ffuser.NewUserBuilder("key-1").Anonymous(false).Build()
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}
