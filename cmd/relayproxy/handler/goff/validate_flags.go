package controller

import (
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/config"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type validateFlags struct {
	version string
	metrics metric.Metrics
	limits  config.ValidateFlagsEndpoint
}

// FlagValidationError is the validation error of a single flag.
type FlagValidationError struct {
	Flag    string `json:"flag"    example:"my-flag"`
	Message string `json:"message" example:"missing default rule"`
}

// FlagValidationResponse is the response of the validate flags endpoint.
type FlagValidationResponse struct {
	Version string                `json:"version" example:"v1.0.0"`
	Valid   bool                  `json:"valid"`
	Errors  []FlagValidationError `json:"errors"`
}

// NewValidateFlags initializes the controller for the /admin/v1/flags/validate endpoint.
// The unset (zero) limits are replaced by their default.
func NewValidateFlags(
	version string, metrics metric.Metrics, limits config.ValidateFlagsEndpoint,
) Controller {
	return &validateFlags{version: version, metrics: metrics, limits: limits.WithDefaults()}
}

// Handler validates a flag file without loading it in the relay proxy.
// @Summary      Validate a flag file
// @Tags Admin API to manage GO Feature Flag
// @Description  Validates the flag file sent in the body of the request, using the same rules as the
// @Description  `goff lint` command.
// @Description  The flag file is not loaded, nothing is changed on the flags served by the relay proxy.
// @Description  The format of the body is selected with the `Content-Type` header (`application/json`,
// @Description  `application/yaml`, `application/x-yaml`, `text/yaml` or `application/toml`).
// @Description  The maximum size of the body is configurable per format with `endpoints.validateFlags`
// @Description  (defaults: 1 MiB for JSON, 64 KiB for YAML and TOML).
// @Security     ApiKeyAuth
// @Accept       json
// @Accept       application/yaml
// @Accept       application/x-yaml
// @Accept       text/yaml
// @Accept       application/toml
// @Produce      json
// @Param        data body object true "The flag file to validate"
// @Success      200  {object} FlagValidationResponse "All the flags are valid"
// @Failure      400  {object} modeldocs.HTTPErrorDoc "Unparseable body"
// @Failure      413  {object} modeldocs.HTTPErrorDoc "Body too large"
// @Failure      415  {object} modeldocs.HTTPErrorDoc "Missing or unsupported content type"
// @Failure      422  {object} FlagValidationResponse "At least one flag is invalid"
// @Router       /admin/v1/flags/validate [post]
func (h *validateFlags) Handler(c *echo.Context) error {
	outcome, err := h.validate(c)
	h.metrics.IncValidateFlags(outcome)
	return err
}

// validate runs the validation and returns the outcome used for the metric alongside the handler error.
func (h *validateFlags) validate(c *echo.Context) (string, error) {
	const outcomeError = metric.ValidateFlagsOutcomeError
	format, maxSize, ok := h.formatFromContentType(c.Request().Header.Get(echo.HeaderContentType))
	if !ok {
		return outcomeError, echo.NewHTTPError(
			http.StatusUnsupportedMediaType,
			"unsupported or missing Content-Type, use application/json, application/yaml or application/toml",
		)
	}

	if c.Request().ContentLength > maxSize {
		return outcomeError, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
	}
	body := http.MaxBytesReader(c.Response(), c.Request().Body, maxSize)
	data, err := io.ReadAll(body)
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			return outcomeError, echo.NewHTTPError(http.StatusRequestEntityTooLarge, "request body too large")
		}
		return outcomeError, echo.NewHTTPError(http.StatusBadRequest, "impossible to read the request body")
	}

	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	_, span := tracer.Start(c.Request().Context(), "validateFlags")
	defer span.End()

	flags, err := configfile.ParseConfig(data, format)
	if err != nil {
		return outcomeError, echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp := FlagValidationResponse{
		Version: h.version,
		Valid:   true,
		Errors:  make([]FlagValidationError, 0),
	}
	for _, flagErr := range configfile.ValidateFlags(flags) {
		resp.Errors = append(resp.Errors, FlagValidationError{
			Flag:    flagErr.Flag,
			Message: flagErr.Err.Error(),
		})
	}
	span.SetAttributes(
		attribute.Int("validateFlags.flagCount", len(flags)),
		attribute.Int("validateFlags.invalidCount", len(resp.Errors)),
	)
	status := http.StatusOK
	outcome := metric.ValidateFlagsOutcomeValid
	if len(resp.Errors) > 0 {
		resp.Valid = false
		status = http.StatusUnprocessableEntity
		outcome = metric.ValidateFlagsOutcomeInvalid
	}
	return outcome, c.JSON(status, resp)
}

// formatFromContentType returns the flag file format matching the content type and the maximum
// size of the body for this format.
func (h *validateFlags) formatFromContentType(contentType string) (string, int64, bool) {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", 0, false
	}
	switch mediaType {
	case "application/json":
		return "json", int64(h.limits.MaxJSONBodySize), true
	case "application/yaml", "application/x-yaml", "text/yaml":
		return "yaml", int64(h.limits.MaxYAMLBodySize), true
	case "application/toml":
		return "toml", int64(h.limits.MaxTOMLBodySize), true
	default:
		return "", 0, false
	}
}
