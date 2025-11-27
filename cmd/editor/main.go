package main

import (
	"net/http"
	"os"

	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	custommiddleware "github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/api/middleware"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/log"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/utils"
)

// This service is an API used to evaluate a flag with an evaluation context
// This API is made for the Flag Editor to be able to evaluate a flag remotely and see if the configuration
// of the flag is working as expected.

func main() {
	logger := log.InitLogger()
	defer func() { _ = logger.ZapLogger.Sync() }()

	e := echo.New()
	e.Use(custommiddleware.ZapLogger(logger.ZapLogger, nil))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://gofeatureflag.org",
			"https://gofeatureflag.org",
			"http://www.gofeatureflag.org",
			"https://www.gofeatureflag.org",
		},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	g := e.Group("/v1")
	g.POST("/feature/evaluate", EvaluateHandler)
	e.GET("/health", HealthHandler)

	if _, exists := os.LookupEnv("RUN_AS_LAMBDA"); exists {
		adapter := awsLambdaHandler{adapter: echoadapter.NewV2(e)}
		adapter.Start()
	} else {
		e.Logger.Fatal(e.Start(":1323"))
	}
}

// EvaluateHandler is the function called when calling the endpoint /v1/feature/evaluate.
// It will perform a flag evaluation and return the resolutionDetails and the value.
func EvaluateHandler(c echo.Context) error {
	u := new(editorEvaluateRequest)
	if err := c.Bind(u); err != nil {
		return err
	}
	f := u.Flag.Convert()
	value, resolutionDetails := f.Value(
		u.FlagName,
		utils.ConvertEvaluationCtxFromRequest(u.Context.Key, u.Context.Custom),
		flag.Context{
			DefaultSdkValue: nil,
		},
	)
	resp := model.VariationResult[interface{}]{
		Value:         value,
		VariationType: resolutionDetails.Variant,
		Reason:        resolutionDetails.Reason,
		ErrorCode:     resolutionDetails.ErrorCode,
		Failed:        resolutionDetails.ErrorCode != "",
		Cacheable:     resolutionDetails.Cacheable,
		Metadata:      f.GetMetadata(),
	}
	return c.JSON(http.StatusOK, resp)
}

// HealthHandler endpoint to validate that the service is up and running.
func HealthHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// editorEvaluateRequest is the format expected to receive from the editor to test the flag.
type editorEvaluateRequest struct {
	Context  ContextWrapper `json:"context,omitempty"`
	Flag     dto.DTO        `json:"flag,omitempty"`
	FlagName string         `json:"flagName,omitempty"`
}

// ContextWrapper is a struct to migrate the API request to an actual evaluation context.
type ContextWrapper struct {
	Key    string `json:"key,omitempty"`
	Custom map[string]interface{}
}
