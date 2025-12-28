package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	middleware2 "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

func TestWebsocketAuthorizer(t *testing.T) {
	type args struct {
		confAPIKey string
		urlAPIKey  string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid apiKey",
			args: args{
				confAPIKey: "valid-api-key",
				urlAPIKey:  "valid-api-key",
			},
			want:    http.StatusOK,
			wantErr: assert.NoError,
		},
		{
			name: "invalid apiKey",
			args: args{
				confAPIKey: "valid-api-key",
				urlAPIKey:  "invalid-api-key",
			},
			want:    http.StatusUnauthorized,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(
				http.MethodGet,
				fmt.Sprintf("/websocket?apiKey=%s", tt.args.urlAPIKey),
				nil,
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			conf := &config.Config{
				AuthorizedKeys: config.APIKeys{
					Evaluation: []string{tt.args.confAPIKey},
				},
			}
			conf.ForceReloadAPIKeys()
			middleware := middleware2.WebsocketAuthorizer(conf)
			handler := middleware(func(c echo.Context) error {
				return c.String(http.StatusOK, "Authorized")
			})

			err := handler(c)
			tt.wantErr(t, err)
		})
	}
}
