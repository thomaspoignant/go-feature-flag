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
				Retriever: &config.RetrieverConf{
					Kind: "file",
					Path: "../../../testdata/flag-config.yaml",
				},
			}
			goff, err := service.NewGoFeatureFlagClient(c, z, []notifier.Notifier{})
			require.NoError(t, err)
			apiServer := New(c, service.Services{
				MonitoringService:    service.NewMonitoring(goff),
				WebsocketService:     service.NewWebsocketService(),
				GOFeatureFlagService: goff,
				Metrics:              metric.Metrics{},
			}, z)

			reqJSON, err := json.Marshal(tt.request)
			require.NoError(t, err)

			// Create a Lambda handler
			handler := lambda.NewHandler(apiServer.getLambdaHandler())

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
