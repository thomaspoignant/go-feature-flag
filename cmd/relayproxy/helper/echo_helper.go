package helper

import "github.com/labstack/echo/v4"

// GetAPIKey extracts the API key from the Authorization header
// It supports both Bearer and Basic authentication schemes
// Returns the API key or an error if the header is missing or invalid
func GetAPIKey(c echo.Context) string {
	apiKey := c.Request().Header.Get("Authorization")
	if len(apiKey) > 7 && apiKey[:7] == "Bearer " {
		apiKey = apiKey[7:]
	}
	return apiKey
}
