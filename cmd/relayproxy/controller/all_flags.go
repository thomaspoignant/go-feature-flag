package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type allFlags struct {
	goFF *ffclient.GoFeatureFlag
}

func NewAllFlags(goFF *ffclient.GoFeatureFlag) Controller {
	return &allFlags{
		goFF: goFF,
	}
}

// Handler is the entry point for the allFlags endpoint
// @Summary      allflags returns all the flag for a specific user.
// @Description  allflags returns all the flag for a specific user.
// @Tags         flags
// @Produce      json
// @Accept			 json
// @Param 			 data body model.RelayProxyRequest true "Payload of the user we want to challenge against the flag."
// @Success      200  {object}   modeldocs.AllFlags "Success"
// @Failure 		 400 {object} modeldocs.HTTPError "Bad Request"
// @Failure      500 {object} modeldocs.HTTPError "Internal server error"
// @Router       /v1/allflags [post]
func (h *allFlags) Handler(c echo.Context) error {
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
		return err
	}

	allFlags := h.goFF.AllFlagsState(goFFUser)
	return c.JSON(http.StatusOK, allFlags)
}
