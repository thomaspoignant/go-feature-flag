package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func AuthMiddlewareErrHandler(_ *echo.Context, err error) error {
	return &echo.HTTPError{
		Code:    http.StatusUnauthorized,
		Message: err.Error(),
	}
}
