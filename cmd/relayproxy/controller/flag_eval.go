package controller

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
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
// @Summary      Evaluate the users with the corresponding flag and return the value for the user.
// @Description  Evaluate the users with the corresponding flag and return the value for the user.
// @Description  Note that you will always have a usable value in the response, you can use the field failed to know if
// @Description  an issue has occurred during the validation of the flag, in that case the value returned will be the
// @Description  default value.
// @Tags         flags
// @Produce      json
// @Accept			 json
// @Param 			 data body model.RelayProxyRequest true "Payload of the user we want to get all the flags from."
// @Param        flag_key path string true "Name of your feature flag"
// @Success      200  {object}   model.FlagEval "Success"
// @Failure 		 400 {object} echo.HTTPError "Bad Request"
// @Failure      500 {object} echo.HTTPError "Internal server error"
// @Router       /v1/feature/{flag_key}/eval [post]
func (h *flagEval) Handler(c echo.Context) error {
	reqBody := new(model.RelayProxyRequest)
	if err := c.Bind(reqBody); err != nil {
		return err
	}

	// validation that we have a reqBody key
	if err := assertRequest(reqBody); err != nil {
		return err
	}
	goFFUser, err := userRequestToUser(reqBody.User)
	if err != nil {
		return nil
	}

	// get flag name from the URL
	flagKey := c.Param("flagKey")
	if flagKey == "" {
		return fmt.Errorf("impossible to find the flag key in the URL")
	}

	flagValue, _ := h.goFF.RawVariation(flagKey, goFFUser, reqBody.DefaultValue)
	return c.JSON(http.StatusOK, flagValue)
}
