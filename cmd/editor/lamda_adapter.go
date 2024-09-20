package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
)

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
