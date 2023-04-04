package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type flagEval struct {
	goFF *ffclient.GoFeatureFlag
}

func NewFlagEval(goFF *ffclient.GoFeatureFlag) Controller {
	return &flagEval{
		goFF: goFF,
	}
}

// Handler is the entry point for the flag eval endpoint
// @Summary     Evaluate a feature flag
// @Description Making a **POST** request to the URL `/v1/feature/<your_flag_name>/eval` will give you the value of the
// @Description flag for this user.
// @Description
// @Description To get a variation you should provide information about the user:
// @Description - User information in JSON in the request body.
// @Description - A default value in case there is an error while evaluating the flag.
// @Description
// @Description Note that you will always have a usable value in the response, you can use the field `failed` to know if
// @Description an issue has occurred during the validation of the flag, in that case the value returned will be the
// @Description default value.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept			 json
// @Param 			 data body model.EvalFlagRequest true "Payload of the user we want to get all the flags from."
// @Param        flag_key path string true "Name of your feature flag"
// @Success      200  {object}   modeldocs.EvalFlagDoc "Success"
// @Failure 		 400 {object} modeldocs.HTTPErrorDoc "Bad Request"
// @Failure      500 {object} modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /v1/feature/{flag_key}/eval [post]
func (h *flagEval) Handler(c echo.Context) error {
	flagKey := c.Param("flagKey")
	if flagKey == "" {
		return fmt.Errorf("impossible to find the flag key in the URL")
	}

	metrics := c.Get(metric.CustomMetrics).(*metric.Metrics)
	metrics.IncFlagEvaluation(flagKey)

	reqBody := new(model.EvalFlagRequest)
	if err := c.Bind(reqBody); err != nil {
		return err
	}

	// validation that we have a reqBody key
	if err := assertRequest(&reqBody.AllFlagRequest); err != nil {
		return err
	}
	goFFUser, err := userRequestToUser(reqBody.User)
	if err != nil {
		return err
	}

	// get flag name from the URL
	flagValue, _ := h.goFF.RawVariation(flagKey, goFFUser, reqBody.DefaultValue)
	return c.JSON(http.StatusOK, flagValue)
}
