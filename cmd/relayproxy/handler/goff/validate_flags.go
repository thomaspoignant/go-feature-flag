package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
)

// maxValidateBodySize is the maximum size of the flag file accepted by the validate endpoint.
const maxValidateBodySize = 1 << 20 // 1 MiB

type validateFlags struct {
	version string
}

// ValidateFlagsError is the validation error of a single flag.
type ValidateFlagsError struct {
	Flag    string `json:"flag"    example:"my-flag"`
	Message string `json:"message" example:"variations are required"`
}

// ValidateFlagsResponse is the response of the validate flags endpoint.
type ValidateFlagsResponse struct {
	Version string               `json:"version" example:"v1.0.0"`
	Valid   bool                 `json:"valid"`
	Errors  []ValidateFlagsError `json:"errors"`
}

// NewValidateFlags initialize the controller for the /admin/v1/flags/validate endpoint
func NewValidateFlags(version string) Controller {
	return &validateFlags{version: version}
}

// Handler validates a flag file without loading it in the relay proxy.
// @Summary      Validate a flag file
// @Tags Admin API to manage GO Feature Flag
// @Description  Validates the flag file sent in the body of the request, using the same rules as the
// @Description  `goff lint` command.
// @Description  The flag file is not loaded, nothing is changed on the flags served by the relay proxy.
// @Description  The format of the body is selected with the `Content-Type` header (`application/json`,
// @Description  `application/yaml`, `application/x-yaml`, `text/yaml` or `application/toml`).
// @Security     ApiKeyAuth
// @Accept       json
// @Accept       application/yaml
// @Accept       application/toml
// @Produce      json
// @Param        data body object true "The flag file to validate"
// @Success      200  {object} ValidateFlagsResponse "All the flags are valid"
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Unparseable body or unsupported content type"
// @Failure      413  {object} modeldocs.HTTPErrorDoc "Body too large"
// @Failure      422  {object} ValidateFlagsResponse "At least one flag is invalid"
// @Router       /admin/v1/flags/validate [post]
func (h *validateFlags) Handler(c *echo.Context) error {
	format, ok := formatFromContentType(c.Request().Header.Get(echo.HeaderContentType))
	if !ok {
		return echo.NewHTTPError(
			http.StatusBadRequest,
			"unsupported or missing Content-Type, use application/json, application/yaml or application/toml",
		)
	}

	body := http.MaxBytesReader(c.Response(), c.Request().Body, maxValidateBodySize)
	data, err := io.ReadAll(body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
		}
		return echo.NewHTTPError(http.StatusBadRequest, "impossible to read the request body")
	}

	flags, err := configfile.ParseConfig(data, format)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp := ValidateFlagsResponse{
		Version: h.version,
		Valid:   true,
		Errors:  make([]ValidateFlagsError, 0),
	}
	for _, flagErr := range configfile.ValidateFlags(flags) {
		resp.Errors = append(resp.Errors, ValidateFlagsError{
			Flag:    flagErr.Flag,
			Message: flagErr.Err.Error(),
		})
	}
	status := http.StatusOK
	if len(resp.Errors) > 0 {
		resp.Valid = false
		status = http.StatusUnprocessableEntity
	}
	return c.JSON(status, resp)
}

// formatFromContentType returns the flag file format matching the content type.
func formatFromContentType(contentType string) (string, bool) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", false
	}
	switch mediaType {
	case "application/json":
		return "json", true
	case "application/yaml", "application/x-yaml", "text/yaml":
		return "yaml", true
	case "application/toml":
		return "toml", true
	default:
		return "", false
	}
}
