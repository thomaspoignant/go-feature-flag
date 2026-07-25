package controller

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/cmdhelpers/configfile"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

const (
	defaultListOffset = 0
	defaultListLimit  = 100
	maxListLimit      = 500
)

type listFlags struct {
	flagsetManager service.FlagsetManager
	metrics        metric.Metrics
}

// NewListFlags initializes the controller for GET /v1/flags.
func NewListFlags(flagsetManager service.FlagsetManager, metrics metric.Metrics) Controller {
	return &listFlags{flagsetManager: flagsetManager, metrics: metrics}
}

// Handler is the entry point for the listFlags endpoint.
// @Summary      List feature flags
// @Tags Flag Management API
// @Description  Returns every flag definition known to the relay proxy for the caller's flagset.
// @Security     ApiKeyAuth
// @Produce      json
// @Param        offset query int false "Pagination offset"
// @Param        limit query int false "Pagination limit (max 500)"
// @Param        type query string false "Filter by flag type"
// @Param        q query string false "Case-insensitive substring match on flag key"
// @Param        disabled query bool false "Filter by disabled state"
// @Success      200  {object} model.FlagListResponse "Success"
// @Failure      401 {object} model.ErrorResponse "Unauthorized"
// @Router       /v1/flags [get]
func (h *listFlags) Handler(httpContext echo.Context) error {
	tracer := otel.GetTracerProvider().Tracer(configfile.OtelTracerName)
	requestContext, span := tracer.Start(httpContext.Request().Context(), "listFlags")
	defer span.End()

	flagset, err := getFlagSet(httpContext, h.flagsetManager)
	if err != nil {
		return err
	}

	flags, cacheReadErr := flagset.GetFlagsFromCacheWithContext(requestContext)
	if cacheReadErr != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, cacheReadErr.Error())
	}

	writable, editable := flagset.GetWritableRetriever()
	source := ""
	if editable {
		source = writable.Source()
	}

	responses := make([]model.FlagResponse, 0, len(flags))
	for key, cachedFlag := range flags {
		flagDefinition, ok := getDtoFromCachedFlag(cachedFlag)
		if !ok {
			continue
		}
		resp, buildErr := buildFlagResponse(key, flagDefinition, source, editable, "")
		if buildErr != nil {
			return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, buildErr.Error())
		}
		responses = append(responses, resp)
	}
	sort.Slice(responses, func(i, j int) bool { return responses[i].Key < responses[j].Key })

	responses = filterFlags(responses, httpContext.QueryParam("type"), httpContext.QueryParam("q"), httpContext.QueryParam("disabled"))
	total := len(responses)
	offset := queryInt(httpContext.QueryParam("offset"), defaultListOffset)
	limit := queryInt(httpContext.QueryParam("limit"), defaultListLimit)
	if limit <= 0 || limit > maxListLimit {
		limit = defaultListLimit
	}
	responses = paginate(responses, offset, limit)

	collectionETag, etagErr := computeCollectionETag(flags)
	if etagErr != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, etagErr.Error())
	}
	httpContext.Response().Header().Set("ETag", collectionETag)

	span.SetAttributes(attribute.Int("listFlags.total", total), attribute.Int("listFlags.returned", len(responses)))
	return httpContext.JSON(http.StatusOK, model.FlagListResponse{
		Flags: responses,
		Meta:  model.FlagListPagination{Total: total, Offset: offset, Limit: limit},
	})
}

func filterFlags(flags []model.FlagResponse, typeFilter, keyQuery, disabled string) []model.FlagResponse {
	if typeFilter == "" && keyQuery == "" && disabled == "" {
		return flags
	}
	filteredFlags := make([]model.FlagResponse, 0, len(flags))
	for _, flagResponse := range flags {
		if typeFilter != "" && string(flagResponse.Type) != typeFilter {
			continue
		}
		if keyQuery != "" && !strings.Contains(strings.ToLower(flagResponse.Key), strings.ToLower(keyQuery)) {
			continue
		}
		if disabled != "" {
			wantDisabled, err := strconv.ParseBool(disabled)
			if err == nil {
				isDisabled := flagResponse.Definition.Disable != nil && *flagResponse.Definition.Disable
				if isDisabled != wantDisabled {
					continue
				}
			}
		}
		filteredFlags = append(filteredFlags, flagResponse)
	}
	return filteredFlags
}

func paginate(flags []model.FlagResponse, offset, limit int) []model.FlagResponse {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(flags) {
		return []model.FlagResponse{}
	}
	endIndex := offset + limit
	if endIndex > len(flags) {
		endIndex = len(flags)
	}
	return flags[offset:endIndex]
}

func queryInt(rawValue string, defaultValue int) int {
	if rawValue == "" {
		return defaultValue
	}
	parsedValue, err := strconv.Atoi(rawValue)
	if err != nil || parsedValue < 0 {
		return defaultValue
	}
	return parsedValue
}
