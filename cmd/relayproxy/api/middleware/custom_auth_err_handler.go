package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AuthMiddlewareErrHandler(err error, _ echo.Context) error {
	return &echo.HTTPError{
		Code:    http.StatusUnauthorized,
		Message: err.Error(),
	}
}
