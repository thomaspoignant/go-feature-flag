package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
)

type health struct {
	monitoringService service.Monitoring
}

// NewHealth is a constructor to create a new controller for the health method
func NewHealth(monitoring service.Monitoring) Controller {
	return &health{
		monitoringService: monitoring,
	}
}

// Handler is the entry point for this API
// @Summary      Health
// @Description Making a **GET** request to the URL path `/health` will tell you if the relay proxy is ready to serve traffic.
// @Description
// @Description This is useful especially for loadbalancer to know that they can send traffic to the service.
// @Produce      json
// @Success      200  {object}   model.HealthResponse
// @Router       /health [get]
//
func (h *health) Handler(c echo.Context) error {
	return c.JSON(http.StatusOK, h.monitoringService.Health())
}
