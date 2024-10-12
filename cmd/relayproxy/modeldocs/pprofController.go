// nolint: lll
package modeldocs

import "github.com/labstack/echo/v4"

// FakePprofController is a fake endpoint for swagger documentation of pprof endpoint
//
// @Summary      pprof endpoint
// @Tags Profiling
// @Description  This endpoint is provided by the echo pprof middleware.
// @Description  To know more please check this blogpost from the GO team https://go.dev/blog/pprof.
// @Description  Visit the page /debug/pprof/ to see the available endpoints, all endpoint are not in the swagger documentation because they are standard pprof endpoints.
// @Description  This endpoint is only available in debug mode.
// @Produce      plain
// @Success      200 {object}	string
// @Router       /debug/pprof/ [get]
func FakePprofController(_ echo.Context) {
	// This is a fake controller, the real entry point is provided by the prometheus middleware.
}
