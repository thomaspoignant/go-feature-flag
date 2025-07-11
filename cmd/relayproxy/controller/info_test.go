package controller_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/controller"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func Test_info_Handler(t *testing.T) {
	type want struct {
		httpCode   int
		handlerErr bool
	}

	tests := []struct {
		name   string
		want   want
		config config.Config
	}{
		{
			name: "valid info default mode",
			want: want{
				httpCode: http.StatusOK,
			},
			config: config.Config{
				CommonFlagSet: config.CommonFlagSet{
					Retriever: &config.RetrieverConf{
						Kind: config.FileRetriever,
						Path: "../testdata/controller/config_flags.yaml",
					},
				},
			},
		},
		{
			name: "valid info flagset mode",
			want: want{
				httpCode: http.StatusOK,
			},
			config: config.Config{
				FlagSets: []config.FlagSet{
					{
						Name:    "teamA",
						ApiKeys: []string{"teamA-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &config.RetrieverConf{
								Kind: config.FileRetriever,
								Path: "../testdata/controller/config_flags.yaml",
							},
						},
					},
					{
						Name:    "teamB",
						ApiKeys: []string{"teamA-api-key"},
						CommonFlagSet: config.CommonFlagSet{
							Retriever: &config.RetrieverConf{
								Kind: config.FileRetriever,
								Path: "../testdata/controller/config_flags.yaml",
							},
						},
					},
				},
			},
		},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			flagsetManager, err := service.NewFlagsetManager(&tt.config, zap.NewNop(), []notifier.Notifier{})
			assert.NoError(t, err, "impossible to create flagset manager")

			srv := service.NewMonitoring(flagsetManager)
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

			assert.Equal(t, tt.want.httpCode, rec.Code, "Invalid HTTP Code")

			// Parse the response JSON to check the structure
			var response map[string]interface{}
			err = json.Unmarshal(rec.Body.Bytes(), &response)
			assert.NoError(t, err, "Response should be valid JSON")

			// Check that cacheRefresh field exists and is a valid timestamp
			cacheRefreshStr, exists := response["cacheRefresh"]
			assert.True(t, exists, "Response should contain cacheRefresh field")

			// Parse the timestamp string
			cacheRefresh, err := time.Parse(time.RFC3339, cacheRefreshStr.(string))
			assert.NoError(t, err, "cacheRefresh should be a valid RFC3339 timestamp")
			assert.False(t, cacheRefresh.IsZero(), "cacheRefresh should not be zero time")

			if tt.config.FlagSets != nil {
				flagsets, exists := response["flagsets"]
				assert.True(t, exists, "Response should contain flagsets field")
				assert.Equal(t, len(tt.config.FlagSets), len(flagsets.(map[string]interface{})), "Number of flagsets should match")
				for _, flagset := range tt.config.FlagSets {
					flagsetName := flagset.Name
					flagsetRefreshDateStr, exists := flagsets.(map[string]interface{})[flagsetName]
					assert.True(t, exists, "Response should contain flagset %s field", flagsetName)

					flagsetRefreshDate, err := time.Parse(time.RFC3339, flagsetRefreshDateStr.(string))
					assert.NoError(t, err, "flagsetRefreshDate should be a valid RFC3339 timestamp")
					assert.False(t, flagsetRefreshDate.IsZero(), "flagsetRefreshDate should not be zero time")
				}
			}
		})
	}
}
