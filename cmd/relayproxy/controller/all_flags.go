package controller

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type allFlags struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewAllFlags(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &allFlags{
		goFF:    goFF,
		metrics: metrics,
	}
}

// Handler is the entry point for the allFlags endpoint
// @Summary      All flags variations for a user
// @Tags GO Feature Flag Evaluation API
// @Description  Making a **POST** request to the URL `/v1/allflags` will give you the values of all the flags for
// @Description this user.
// @Description
// @Description To get a variation you should provide information about the user.
// @Description For that you should provide some user information in JSON in the request body.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept		 json
// @Param 	     data body model.AllFlagRequest true "Payload of the user we want to challenge against the flag."
// @Success      200  {object} modeldocs.AllFlags "Success"
// @Failure      400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /v1/allflags [post]
func (h *allFlags) Handler(c echo.Context) error {
	h.metrics.IncAllFlag()

	reqBody := new(model.AllFlagRequest)
	if err := c.Bind(reqBody); err != nil {
		return err
	}
	// validation that we have a reqBody key
	if err := assertRequest(reqBody); err != nil {
		return err
	}

	evaluationCtx, err := evaluationContextFromRequest(reqBody)
	if err != nil {
		return err
	}
	tracer := otel.GetTracerProvider().Tracer(config.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "AllFlagsState")
	defer span.End()
	allFlags := h.goFF.AllFlagsState(evaluationCtx)
	span.SetAttributes(
		attribute.Bool("AllFlagsState.valid", allFlags.IsValid()),
		attribute.Int("AllFlagsState.numberEvaluation", len(allFlags.GetFlags())),
	)
	return c.JSON(http.StatusOK, allFlags)
}
