package api

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
)

// newAwsLambdaHandlerManager is creating a new awsLambdaHandler struct with the echoadapter
// to proxy all lambda event to echo.
func newAwsLambdaHandlerManager(echoInstance *echo.Echo, basePath string) awsLambdaHandler {
	adapterAPIGtwV2 := echoadapter.NewV2(echoInstance)
	adapterALB := echoadapter.NewALB(echoInstance)
	adapterAPIGtwV1 := echoadapter.New(echoInstance)

	// Configure base path stripping if a base path is provided
	if basePath != "" {
		adapterAPIGtwV2.StripBasePath(basePath)
		adapterALB.StripBasePath(basePath)
		adapterAPIGtwV1.StripBasePath(basePath)
	}

	return awsLambdaHandler{
		adapterAPIGtwV2: adapterAPIGtwV2,
		adapterALB:      adapterALB,
		adapterAPIGtwV1: adapterAPIGtwV1,
	}
}

type awsLambdaHandler struct {
	adapterAPIGtwV2 *echoadapter.EchoLambdaV2
	adapterAPIGtwV1 *echoadapter.EchoLambda
	adapterALB      *echoadapter.EchoLambdaALB
}

// SelectAdapter returns the appropriate adapter based on the mode.
func (h *awsLambdaHandler) SelectAdapter(mode string) interface{} {
	switch strings.ToUpper(mode) {
	case strings.ToUpper(config.LambdaAdapterAPIGatewayV1):
		return h.HandlerAPIGatewayV1
	case strings.ToUpper(config.LambdaAdapterALB):
		return h.HandlerALB
	default:
		return h.HandlerAPIGatewayV2
	}
}

// HandlerAPIGatewayV2 is the function that proxy the lambda events to echo calls for API Gateway V2.
func (h *awsLambdaHandler) HandlerAPIGatewayV2(
	ctx context.Context,
	req events.APIGatewayV2HTTPRequest,
) (
	events.APIGatewayV2HTTPResponse, error) {
	return h.adapterAPIGtwV2.ProxyWithContext(ctx, req)
}

// HandlerAPIGatewayV1 is the function that proxy the lambda events to echo calls for API Gateway V1.
func (h *awsLambdaHandler) HandlerAPIGatewayV1(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (
	events.APIGatewayProxyResponse, error) {
	return h.adapterAPIGtwV1.ProxyWithContext(ctx, req)
}

// HandlerALB is the function that proxy the lambda events to echo calls for API Gateway V1.
func (h *awsLambdaHandler) HandlerALB(ctx context.Context, req events.ALBTargetGroupRequest) (
	events.ALBTargetGroupResponse, error) {
	return h.adapterALB.ProxyWithContext(ctx, req)
}
