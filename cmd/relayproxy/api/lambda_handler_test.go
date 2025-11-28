package api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/retrieverconf"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"go.uber.org/zap"
)

func TestAwsLambdaHandler_GetAdapter(t *testing.T) {
	type test struct {
		name    string
		mode    string
		request interface{}
	}

	tests := []test{
		{
			name: "APIGatewayV2 event handler",
			mode: "APIGatewayV2",
			request: events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "GET",
						Path:   "/health",
					},
				},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "",
			},
		},
		{
			name: "APIGatewayV1 event handler",
			mode: "APIGatewayV1",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/health",
				RequestContext: events.APIGatewayProxyRequestContext{
					Path:       "/health",
					HTTPMethod: "GET",
				},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "",
			},
		},
		{
			name: "ALB event handler",
			mode: "ALB",
			request: events.ALBTargetGroupRequest{
				HTTPMethod: "GET",
				Path:       "/health",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z, err := zap.NewProduction()
			require.NoError(t, err)
			c := &config.Config{
				StartAsAwsLambda: true,
				AwsLambdaAdapter: tt.mode,
				CommonFlagSet: config.CommonFlagSet{
					Retrievers: &[]retrieverconf.RetrieverConf{
						{
							Kind: "file",
							Path: "../../../testdata/flag-config.yaml",
						},
					},
				},
			}
			flagsetManager, err := service.NewFlagsetManager(c, z, []notifier.Notifier{})
			require.NoError(t, err)
			apiServer := New(c, service.Services{
				MonitoringService: service.NewMonitoring(flagsetManager),
				WebsocketService:  service.NewWebsocketService(),
				FlagsetManager:    flagsetManager,
				Metrics:           metric.Metrics{},
			}, z)

			reqJSON, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create a Lambda handler
			handler := lambda.NewHandler(apiServer.lambdaHandler())

			// Invoke the handler with the mock event
			response, err := handler.Invoke(context.Background(), reqJSON)
			require.NoError(t, err)

			switch strings.ToLower(tt.mode) {
			case "apigatewayv2":
				var res events.APIGatewayV2HTTPResponse
				err = json.Unmarshal(response, &res)
				require.NoError(t, err)
				require.Equal(t, 200, res.StatusCode)
			case "apigatewayv1":
				var res events.APIGatewayProxyResponse
				err = json.Unmarshal(response, &res)
				require.NoError(t, err)
				require.Equal(t, 200, res.StatusCode)
			case "alb":
				var res events.ALBTargetGroupResponse
				err = json.Unmarshal(response, &res)
				require.NoError(t, err)
				require.Equal(t, 200, res.StatusCode)
			default:
				require.Fail(t, "not implemented")
			}
		})
	}
}

func TestAwsLambdaHandler_BasePathSupport(t *testing.T) {
	tests := []struct {
		name         string
		basePath     string
		requestPath  string
		expectedCode int
	}{
		{
			name:         "Health endpoint with /api base path",
			basePath:     "/api",
			requestPath:  "/api/health",
			expectedCode: 200,
		},
		{
			name:         "Health endpoint with /api/feature-flags base path",
			basePath:     "/api/feature-flags",
			requestPath:  "/api/feature-flags/health",
			expectedCode: 200,
		},
		{
			name:         "Info endpoint with base path",
			basePath:     "/dev",
			requestPath:  "/dev/info",
			expectedCode: 200,
		},
		{
			name:         "No base path configured - direct path",
			basePath:     "",
			requestPath:  "/health",
			expectedCode: 200,
		},
		{
			name:         "No base path configured - direct path",
			basePath:     "",
			requestPath:  "/health",
			expectedCode: 200,
		},
	}

	z, err := zap.NewProduction()
	require.NoError(t, err)

	flagsetManager, err := service.NewFlagsetManager(&config.Config{
		CommonFlagSet: config.CommonFlagSet{
			Retrievers: &[]retrieverconf.RetrieverConf{
				{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			},
		},
	}, z, nil)
	require.NoError(t, err)

	commonServices := service.Services{
		MonitoringService: service.NewMonitoring(flagsetManager),
		WebsocketService:  service.NewWebsocketService(),
		FlagsetManager:    flagsetManager,
		Metrics:           metric.Metrics{},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config.Config{
				StartAsAwsLambda:      true,
				AwsLambdaAdapter:      "APIGatewayV2",
				AwsApiGatewayBasePath: tt.basePath,
			}
			apiServer := New(c, commonServices, z)

			request := events.APIGatewayV2HTTPRequest{
				RequestContext: events.APIGatewayV2HTTPRequestContext{
					HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
						Method: "GET",
						Path:   tt.requestPath,
					},
				},
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: "",
			}

			reqJSON, err := json.Marshal(request)
			require.NoError(t, err)

			// Create a Lambda handler
			handler := lambda.NewHandler(apiServer.lambdaHandler())

			// Invoke the handler with the mock event
			response, err := handler.Invoke(context.Background(), reqJSON)
			require.NoError(t, err)

			var res events.APIGatewayV2HTTPResponse
			err = json.Unmarshal(response, &res)
			require.NoError(t, err)
			require.Equal(t, tt.expectedCode, res.StatusCode)
		})
	}
}
