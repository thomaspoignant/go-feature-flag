package controller_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

const mockConfigFlagsLocation = "../testdata/controller/configuration/"

func TestFlagConfigurationAPICtrl_Handler(t *testing.T) {
	defaultGoff, err := ffclient.New(ffclient.Config{
		Retrievers: []retriever.Retriever{
			&fileretriever.Retriever{Path: "../testdata/controller/configuration_flags.yaml"},
		},
	})
	assert.NoError(t, err)
	type want struct {
		bodyLocation string
		statusCode   int
	}
	test := []struct {
		name        string
		goff        *ffclient.GoFeatureFlag
		requestBody string
		want        struct {
			bodyLocation string
			statusCode   int
		}
	}{
		{
			name:        "Test with empty body",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			goff:        defaultGoff,
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty.json",
			},
		},
		{
			name:        "Test with empty flags ",
			requestBody: mockConfigFlagsLocation + "requests/empty-flag-array.json",
			goff:        defaultGoff,
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/empty-flag-array.json",
			},
		},
		{
			name:        "Filter flags",
			requestBody: mockConfigFlagsLocation + "requests/filter-flags.json",
			goff:        defaultGoff,
			want: want{
				statusCode:   http.StatusOK,
				bodyLocation: mockConfigFlagsLocation + "responses/filter-flags.json",
			},
		},
		{
			name:        "Invalid JSON",
			requestBody: mockConfigFlagsLocation + "requests/invalid-json.json",
			goff:        defaultGoff,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:        "Offline mode",
			requestBody: mockConfigFlagsLocation + "requests/empty.json",
			goff: func() *ffclient.GoFeatureFlag {
				goff, err := ffclient.New(ffclient.Config{
					Retrievers: []retriever.Retriever{
						&fileretriever.Retriever{Path: "../testdata/controller/configuration_flags.yaml"},
					},
					Offline: true,
				})
				assert.NoError(t, err)
				return goff
			}(),
			want: want{
				statusCode: http.StatusInternalServerError,
			},
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := controller.NewAPIFlagConfiguration(tt.goff, metric.Metrics{})
			e := echo.New()
			rec := httptest.NewRecorder()

			// read the request body from the file
			requestBody, err := os.ReadFile(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(echo.POST, "/v1/flag/configuration", strings.NewReader(string(requestBody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			c := e.NewContext(req, rec)
			c.SetPath("/v1/flag/configuration")

			// Call the handler
			assert.NoError(t, ctrl.Handler(c))

			assert.Equal(t, tt.want.statusCode, rec.Code)

			if tt.want.bodyLocation != "" {
				wantBody, err := os.ReadFile(tt.want.bodyLocation)
				assert.NoError(t, err)
				assert.JSONEq(t, string(wantBody), rec.Body.String())
			}
		})
	}
}
