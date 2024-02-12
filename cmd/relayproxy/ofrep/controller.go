package ofrep

import "github.com/labstack/echo/v4"

type Controller interface {
	OFREPHandler(c echo.Context) error
}
