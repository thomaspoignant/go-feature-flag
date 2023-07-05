package api

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
)

// newAwsLambdaHandler is creating a new awsLambdaHandler struct with the echoadapter
// to proxy all lambda event to echo.
func newAwsLambdaHandler(echoInstance *echo.Echo) awsLambdaHandler {
	return awsLambdaHandler{
		adapter: echoadapter.NewV2(echoInstance),
	}
}

type awsLambdaHandler struct {
	adapter *echoadapter.EchoLambdaV2
}

func (h *awsLambdaHandler) Start() {
	lambda.Start(h.Handler)
}

// Handler is the function that proxy the lambda events to echo calls.
func (h *awsLambdaHandler) Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (
	events.APIGatewayV2HTTPResponse, error) {
	return h.adapter.ProxyWithContext(ctx, req)
}
