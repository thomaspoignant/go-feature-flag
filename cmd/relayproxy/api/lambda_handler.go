package api

import (
	"context"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
)

// newAwsLambdaHandlerManager is creating a new awsLambdaHandler struct with the echoadapter
// to proxy all lambda event to echo.
func newAwsLambdaHandlerManager(echoInstance *echo.Echo) awsLambdaHandler {
	return awsLambdaHandler{
		adapterApiGtwV2: echoadapter.NewV2(echoInstance),
		adapterALB:      echoadapter.NewALB(echoInstance),
		adapterApiGtwV1: echoadapter.New(echoInstance),
	}
}

type awsLambdaHandler struct {
	adapterApiGtwV2 *echoadapter.EchoLambdaV2
	adapterApiGtwV1 *echoadapter.EchoLambda
	adapterALB      *echoadapter.EchoLambdaALB
}

func (h *awsLambdaHandler) GetAdapter(mode string) interface{} {
	switch strings.ToUpper(mode) {
	case "APIGATEWAYV1":
		return h.HandlerApiGatewayV1
	case "ALB":
		return h.HandlerALB
	default:
		return h.HandlerApiGatewayV2
	}
}

// HandlerApiGatewayV2 is the function that proxy the lambda events to echo calls for API Gateway V2.
func (h *awsLambdaHandler) HandlerApiGatewayV2(ctx context.Context, req events.APIGatewayV2HTTPRequest) (
	events.APIGatewayV2HTTPResponse, error) {
	return h.adapterApiGtwV2.ProxyWithContext(ctx, req)
}

// HandlerApiGatewayV1 is the function that proxy the lambda events to echo calls for API Gateway V1.
func (h *awsLambdaHandler) HandlerApiGatewayV1(ctx context.Context, req events.APIGatewayProxyRequest) (
	events.APIGatewayProxyResponse, error) {
	return h.adapterApiGtwV1.ProxyWithContext(ctx, req)
}

// HandlerALB is the function that proxy the lambda events to echo calls for API Gateway V1.
func (h *awsLambdaHandler) HandlerALB(ctx context.Context, req events.ALBTargetGroupRequest) (
	events.ALBTargetGroupResponse, error) {
	return h.adapterALB.ProxyWithContext(ctx, req)
}
