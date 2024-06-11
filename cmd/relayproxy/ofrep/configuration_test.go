package ofrep_test

import (
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func Test_Configuration(t *testing.T) {
	type want struct {
		httpCode int
		response string
	}

	tests := []struct {
		name        string
		goffPolling time.Duration
		want        want
	}{
		{
			name:        "configuration 10 sec",
			goffPolling: 10 * time.Second,
			want: want{
				httpCode: http.StatusOK,
				response: "{\"name\":\"GO Feature Flag\",\"capabilities\":{\"cacheInvalidation\":{\"polling\":{\"enabled\":true,\"minPollingInterval\":10000}},\"flagEvaluation\":{\"unsupportedTypes\":[]}}}",
			},
		},
		{
			name:        "configuration 10 minute",
			goffPolling: 10 * time.Minute,
			want: want{
				httpCode: http.StatusOK,
				response: "{\"name\":\"GO Feature Flag\",\"capabilities\":{\"cacheInvalidation\":{\"polling\":{\"enabled\":true,\"minPollingInterval\":600000}},\"flagEvaluation\":{\"unsupportedTypes\":[]}}}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// init go-feature-flag
			goFF, err := ffclient.New(ffclient.Config{
				PollingInterval: tt.goffPolling,
				Context:         context.Background(),
				Retriever: &fileretriever.Retriever{
					Path: configFlagsLocation,
				},
			})
			defer goFF.Close()
			assert.NoError(t, err)

			ctrl := ofrep.NewOFREPEvaluate(goFF, metric.Metrics{})
			e := echo.New()
			e.GET("/ofrep/v1/configuration", ctrl.Configuration)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(echo.GET, "/ofrep/v1/configuration", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")
			assert.JSONEq(t, tt.want.response, rec.Body.String(), "Invalid response wantBody")
		})
	}
}
