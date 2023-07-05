package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/server/service"
)

type info struct {
	monitoringService service.Monitoring
}

func NewInfo(monitoring service.Monitoring) Controller {
	return &info{
		monitoringService: monitoring,
	}
}

// Handler is the entry point for the Info API
// @Summary      Info
// @Description  Making a **GET** request to the URL path `/info` will give you information about the actual state
// @Description  of the server.
// @Description
// @Description	As of Today the level of information is small be we can improve this endpoint to returns more
// @Description information.
// @Produce      json
// @Success      200  {object}   model.InfoResponse
// @Router       /info [get]
func (h *info) Handler(c echo.Context) error {
	return c.JSON(http.StatusOK, h.monitoringService.Info())
}
