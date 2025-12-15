package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type FlagConfigurationAPICtrl struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

func NewAPIFlagConfiguration(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &FlagConfigurationAPICtrl{
		flagsetManager: flagsetManager,
		metrics:        metrics,
	}
}

type FlagConfigurationRequest struct {
	Flags []string `json:"flags"`
}

type FlagConfigurationError = string

const (
	FlagConfigErrorInvalidRequest  FlagConfigurationError = "INVALID_REQUEST"
	FlagConfigErrorRetrievingFlags FlagConfigurationError = "RETRIEVING_FLAGS_ERROR"
)

type FlagConfigurationResponse struct {
	Flags                       map[string]flag.Flag `json:"flags,omitempty"`
	EvaluationContextEnrichment map[string]any       `json:"evaluationContextEnrichment,omitempty"`
	ErrorCode                   string               `json:"errorCode,omitempty"`
	ErrorDetails                string               `json:"errorDetails,omitempty"`
}

// Handler is the endpoint to poll if you want to get the configuration of the flags.
// @Summary    Endpoint to poll if you want to get the configuration of the flags.
// @Tags GO Feature Flag Evaluation API
// @Description Making a **POST** request to the URL `/v1/flag/configuration` will give you the list of
// @Description the flags to use them for local evaluation in your provider.
// @Security    ApiKeyAuth
// @Security    XApiKeyAuth
// @Produce     json
// @Accept      json
// @Param 		data body FlagConfigurationRequest false "List of flags to get the configuration from."
// @Param       If-None-Match header string false "The request will be processed only if ETag doesn't match."
// @Success     200  {object} FlagConfigurationResponse "Success"
// @Success     304 {string} string "Etag: \"117-0193435c612c50d93b798619d9464856263dbf9f\""
// @Failure     500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router      /v1/flag/configuration [post]
func (h *FlagConfigurationAPICtrl) Handler(c echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "flagConfiguration")
	defer span.End()

	reqBody := new(FlagConfigurationRequest)
	if err := c.Bind(reqBody); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			FlagConfigurationResponse{
				ErrorCode:    FlagConfigErrorInvalidRequest,
				ErrorDetails: fmt.Sprintf("impossible to read request body: %s", err),
			},
		)
	}

	flagset, httpErr := helper.FlagSet(h.flagsetManager, helper.APIKey(c))
	if httpErr != nil {
		return httpErr
	}

	flags, err := flagset.GetFlagsFromCache()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, FlagConfigurationResponse{
			ErrorCode:    FlagConfigErrorRetrievingFlags,
			ErrorDetails: fmt.Sprintf("impossible to retrieve flag configuration: %s", err),
		})
	}

	// filter if we have a list of flags in the request.
	if len(reqBody.Flags) > 0 {
		tmpFlags := map[string]flag.Flag{}
		for _, flagKey := range reqBody.Flags {
			if _, ok := flags[flagKey]; ok {
				tmpFlags[flagKey] = flags[flagKey]
			}
		}
		flags = tmpFlags
	}

	span.SetAttributes(attribute.Int("flagConfiguration.configurationSize", len(flags)))
	c.Response().Header().
		Set(echo.HeaderLastModified, flagset.GetCacheRefreshDate().
			Format(time.RFC1123))
	return c.JSON(
		http.StatusOK,
		FlagConfigurationResponse{
			EvaluationContextEnrichment: flagset.GetEvaluationContextEnrichment(),
			Flags:                       flags,
		},
	)
}
