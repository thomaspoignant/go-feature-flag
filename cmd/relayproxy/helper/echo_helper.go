package helper

import (
	"net/http"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
)

// GetAPIKey extracts the API key from the Authorization header.
// It removes the "Bearer " prefix if it exists.
// For other schemes, it returns the raw header value or an empty string if the header is missing.
func GetAPIKey(c echo.Context) string {
	apiKey := c.Request().Header.Get("Authorization")
	if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
		apiKey = apiKey[7:]
	}
	return apiKey
}

// GetFlagSet retrieves the flagset for the given API key from the flagset manager
// This layer ensure that the flagset manager is initialized and that the API key is valid
func GetFlagSet(flagsetManager service.FlagsetManager, apiKey string) (*ffclient.GoFeatureFlag, *echo.HTTPError) {
	if flagsetManager == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "flagset manager is not initialized")
	}
	flagset, err := flagsetManager.GetFlagSet(apiKey)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "error while getting flagset: "+err.Error())
	}
	return flagset, nil
}
