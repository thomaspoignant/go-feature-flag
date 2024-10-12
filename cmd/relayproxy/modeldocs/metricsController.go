package modeldocs

import "github.com/labstack/echo/v4"

// FakeMetricsController is a fake entry point for swagger documentation
//
// @Summary      Prometheus endpoint
// @Tags Monitoring
// @Description  This endpoint is providing metrics about the relay proxy in the prometheus format.
// @Produce      plain
// @Success      200 {object}	string
// @Router       /metrics [get]
func FakeMetricsController(_ echo.Context) {
	// This is a fake controller, the real entry point is provided by the prometheus middleware.
}
