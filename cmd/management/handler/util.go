package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/thomaspoignant/go-feature-flag/cmd/management/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/management/service"
)

func ok(c echo.Context, data any) error {
	return c.JSON(http.StatusOK, model.APIResponse{Success: true, Data: data})
}

func created(c echo.Context, data any) error {
	return c.JSON(http.StatusCreated, model.APIResponse{Success: true, Data: data})
}

func badRequest(c echo.Context, msg string) error {
	return c.JSON(http.StatusBadRequest, model.APIResponse{Success: false, Message: msg})
}

func notFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, model.APIResponse{Success: false, Message: msg})
}

func internal(c echo.Context, msg string) error {
	return c.JSON(http.StatusInternalServerError, model.APIResponse{Success: false, Message: msg})
}

func writeServiceError(c echo.Context, err error, opMsg string) error {
	var verrs model.ValidationErrors
	if errors.As(err, &verrs) {
		return c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false, Message: "validation failed", Errors: verrs.ToMap(),
		})
	}
	if errors.Is(err, service.ErrNotFound) {
		return notFound(c, "not found")
	}
	if errors.Is(err, service.ErrForbidden) {
		return c.JSON(http.StatusForbidden, model.APIResponse{Success: false, Message: "forbidden"})
	}
	if errors.Is(err, service.ErrConflict) {
		return c.JSON(http.StatusConflict, model.APIResponse{Success: false, Message: err.Error()})
	}
	c.Logger().Error(opMsg, err)
	return internal(c, opMsg)
}
