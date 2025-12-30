package helper

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
)

// APIKey extracts the API key from the request headers.
// It checks headers in the following order of precedence:
// 1. X-API-Key header (raw value)
// 2. Authorization header (with "Bearer " prefix removed if present)
// Returns an empty string if no API key is found.
func APIKey(c echo.Context) string {
	// First, check X-API-Key header (takes precedence)
	if xAPIKey := c.Request().Header.Get(XAPIKeyHeader); xAPIKey != "" {
		return xAPIKey
	}

	// Fall back to Authorization header
	apiKey := c.Request().Header.Get(AuthorizationHeader)
	if len(apiKey) >= len(BearerPrefix) && strings.EqualFold(apiKey[:len(BearerPrefix)], BearerPrefix) {
		return strings.TrimSpace(apiKey[len(BearerPrefix):])
	}
	return apiKey
}

// FlagSet retrieves the flagset for the given API key from the flagset manager
// This layer ensure that the flagset manager is initialized and that the API key is valid
func FlagSet(flagsetManager service.FlagsetManager, apiKey string) (*ffclient.GoFeatureFlag, *echo.HTTPError) {
	if flagsetManager == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "flagset manager is not initialized")
	}
	flagset, err := flagsetManager.FlagSet(apiKey)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "error while getting flagset: "+err.Error())
	}
	return flagset, nil
}
