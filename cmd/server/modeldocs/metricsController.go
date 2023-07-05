package modeldocs

import "github.com/labstack/echo/v4"

// FakeMetricsController is the entry point for the allFlags endpoint
//
// @Summary      Prometheus endpoint
// @Description  This endpoint is providing metrics in the prometheus format.
// @Produce      plain
// @Success      200 {object}	string
// @Router       /metrics [get]
func FakeMetricsController(_ echo.Context) {
	// This is a fake controller, the real entry point is provided by the prometheus middleware.
}
