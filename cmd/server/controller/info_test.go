package controller_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/service"
)

func Test_info_Handler(t *testing.T) {
	type want struct {
		httpCode   int
		bodyFile   string
		handlerErr bool
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "valid info",
			want: want{
				httpCode: http.StatusOK,
				bodyFile: "../testdata/controller/info/valid.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init GO feature flag
			goFF, _ := ffclient.New(ffclient.Config{Offline: true})
			srv := service.NewMonitoring(goFF)
			infoCtrl := controller.NewInfo(srv)

			e := echo.New()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.GET, "/info", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			res := infoCtrl.Handler(c)

			if tt.want.handlerErr {
				assert.Error(t, res, "handler should return an error")
				return
			}
			body, err := ioutil.ReadFile(tt.want.bodyFile)
			assert.NoError(t, err, "Impossible the expected body file %s", tt.want.bodyFile)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, string(body), rec.Body.String(), "Invalid response body")
		})
	}
}
