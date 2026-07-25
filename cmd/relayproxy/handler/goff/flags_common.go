package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/helper"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/service"
	"github.com/thomaspoignant/go-feature-flag/modules/core/dto"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

// flagKeyPattern matches the OpenAPI FlagKey schema (^[A-Za-z0-9._-]+$).
var flagKeyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// writeError writes the shared {errorCode, errorDetails} envelope used by every
// flag-management endpoint and returns a non-nil error so callers can propagate it as
// "already handled, stop here" (c.JSON itself only returns a non-nil error on a write
// failure, which would make a nil-returning helper indistinguishable from success).
// Echo's default error handler no-ops once the response is committed, so this never
// causes a second response to be written.
func writeError(httpContext echo.Context, status int, code flag.ErrorCode, details string) error {
	if responseWriteErr := httpContext.JSON(status, model.ErrorResponse{ErrorCode: code, ErrorDetails: details}); responseWriteErr != nil {
		return responseWriteErr
	}
	return echo.NewHTTPError(status, details)
}

// getFlagSet resolves the caller's flagSet from its API key. On failure it writes the
// error response itself and returns a non-nil error that the handler should return as-is.
func getFlagSet(httpContext echo.Context, flagSetManager service.FlagsetManager) (*ffclient.GoFeatureFlag, error) {
	flagSet, httpErr := helper.FlagSet(flagSetManager, helper.APIKey(httpContext))
	if httpErr != nil {
		errorDetails, hasDetails := httpErr.Message.(string)
		if !hasDetails {
			errorDetails = ""
		}
		return nil, writeError(httpContext, httpErr.Code, flag.ErrorCodeGeneral, errorDetails)
	}
	return flagSet, nil
}

// resolveWritableFlagSet resolves the caller's flagSet and requires it to be backed by
// exactly one WritableRetriever (currently: a single PostgreSQL retriever with an explicit
// flagSet configured). On failure it writes the error response itself.
func resolveWritableFlagSet(
	httpContext echo.Context, flagSetManager service.FlagsetManager,
) (*ffclient.GoFeatureFlag, retriever.WritableRetriever, error) {
	flagSet, resolveErr := getFlagSet(httpContext, flagSetManager)
	if resolveErr != nil {
		return nil, nil, resolveErr
	}
	flagSetWritableRetriever, hasWritableRetriever := flagSet.GetWritableRetriever()
	if !hasWritableRetriever {
		return nil, nil, writeError(httpContext, http.StatusForbidden, flag.ErrorFlagConfiguration,
			"the flag management API requires a flagSet backed by exactly one writable "+
				"(PostgreSQL) retriever")
	}
	return flagSet, flagSetWritableRetriever, nil
}

// mapWriteErr maps the sentinel errors returned by retriever.WritableRetriever to an HTTP
// status and errorCode.
func mapWriteErr(writeErr error) (status int, code flag.ErrorCode) {
	switch {
	case errors.Is(writeErr, retriever.ErrFlagNotFound):
		return http.StatusNotFound, flag.ErrorCodeFlagNotFound
	case errors.Is(writeErr, retriever.ErrFlagAlreadyExists):
		return http.StatusConflict, flag.ErrorFlagConfiguration
	case errors.Is(writeErr, retriever.ErrETagMismatch):
		return http.StatusPreconditionFailed, flag.ErrorFlagConfiguration
	case errors.Is(writeErr, retriever.ErrFlagsetNotConfigured):
		return http.StatusForbidden, flag.ErrorFlagConfiguration
	default:
		return http.StatusInternalServerError, flag.ErrorCodeGeneral
	}
}

// validateFlagKey checks the flag key against the OpenAPI FlagKey pattern.
func validateFlagKey(inputKey string) error {
	if inputKey == "" || !flagKeyPattern.MatchString(inputKey) {
		return errors.New("flag key must match ^[A-Za-z0-9._-]+$")
	}
	return nil
}

// readBody reads and returns the raw request body.
func readBody(httpContext echo.Context) ([]byte, error) {
	return io.ReadAll(httpContext.Request().Body)
}

// validateDefinition unmarshals a raw flag definition and validates it using the existing
// GOFF/OpenFeature validation rules (flag.InternalFlag.IsValid()).
func validateDefinition(rawFlagDefinition []byte) (dto.DTO, error) {
	var definitionDTO dto.DTO
	if unmarshalErr := json.Unmarshal(rawFlagDefinition, &definitionDTO); unmarshalErr != nil {
		return dto.DTO{}, unmarshalErr
	}
	internalFlagDefinition := definitionDTO.Convert()
	if validationErr := internalFlagDefinition.IsValid(); validationErr != nil {
		return dto.DTO{}, validationErr
	}
	return definitionDTO, nil
}

// computeETag returns a strong ETag for a flag definition, computed over the canonical JSON
// encoding of its dto.DTO (Go's encoder sorts map keys, giving a stable hash regardless of the
// original byte layout). This must produce the same result regardless of whether the
// definition came from the cache or from a WritableRetriever, so GET and write paths agree.
func computeETag(definitionDTO dto.DTO) (string, error) {
	serializedJsonDefinitionDto, marshalErr := json.Marshal(definitionDTO)
	if marshalErr != nil {
		return "", marshalErr
	}
	definitionChecksum := sha256.Sum256(serializedJsonDefinitionDto)
	return `"` + hex.EncodeToString(definitionChecksum[:]) + `"`, nil
}

// computeCollectionETag returns a strong ETag over the whole flag map, used for the
// GET /v1/flags collection ETag header. json.Marshal of a Go map deterministically sorts
// keys, so this is stable across calls as long as the content hasn't changed.
func computeCollectionETag(flags map[string]flag.Flag) (string, error) {
	collectionDefinitions := make(map[string]dto.DTO, len(flags))
	for flagKey, cachedFlag := range flags {
		if definitionDTO, hasDefinition := getDtoFromCachedFlag(cachedFlag); hasDefinition {
			collectionDefinitions[flagKey] = definitionDTO
		}
	}
	serializedJsonDefinitionDtos, marshalErr := json.Marshal(collectionDefinitions)
	if marshalErr != nil {
		return "", marshalErr
	}
	checksum := sha256.Sum256(serializedJsonDefinitionDtos)
	return `"` + hex.EncodeToString(checksum[:]) + `"`, nil
}

// ifMatchHeader returns the If-Match header value, or nil if absent.
func ifMatchHeader(httpContext echo.Context) *string {
	ifMatchValue := httpContext.Request().Header.Get("If-Match")
	if ifMatchValue == "" {
		return nil
	}
	return &ifMatchValue
}

// getDtoFromCachedFlag extracts the dto.DTO of a cached flag.Flag. GOFF's in-memory cache only
// ever stores *flag.InternalFlag behind the flag.Flag interface (internal/cache), so this
// assertion is expected to always succeed for flags coming from GetFlagsFromCache().
func getDtoFromCachedFlag(cachedFlag flag.Flag) (dto.DTO, bool) {
	internalFlag, isInternalFlag := cachedFlag.(*flag.InternalFlag)
	if !isInternalFlag {
		return dto.DTO{}, false
	}
	return dto.ConvertInternalFlagToDto(*internalFlag), true
}

// inferFlagType infers the OpenFeature FlagType from the first non-nil variation value.
func inferFlagType(definitionDTO dto.DTO) model.FlagType {
	if definitionDTO.Variations == nil {
		return model.FlagTypeObject
	}
	for _, variationValuePointer := range *definitionDTO.Variations {
		if variationValuePointer == nil || *variationValuePointer == nil {
			continue
		}
		switch typedVariationValue := (*variationValuePointer).(type) {
		case bool:
			return model.FlagTypeBoolean
		case string:
			return model.FlagTypeString
		case float64:
			if typedVariationValue == float64(int64(typedVariationValue)) {
				return model.FlagTypeInteger
			}
			return model.FlagTypeFloat
		default:
			return model.FlagTypeObject
		}
	}
	return model.FlagTypeObject
}

// buildFlagResponse assembles the model.FlagResponse for a single flag, computing its ETag
// if not already known (etag == "").
func buildFlagResponse(key string, definitionDTO dto.DTO, source string, editable bool, etag string) (model.FlagResponse, error) {
	if etag == "" {
		computedETag, computeETagErr := computeETag(definitionDTO)
		if computeETagErr != nil {
			return model.FlagResponse{}, computeETagErr
		}
		etag = computedETag
	}
	return model.FlagResponse{
		Key:        key,
		Type:       inferFlagType(definitionDTO),
		Definition: definitionDTO,
		Source:     source,
		Editable:   editable,
		ETag:       etag,
	}, nil
}

// reload forces the flagSet to reload its retrievers so that evaluation is
// immediately consistent with the write that was just performed.
func reload(httpContext echo.Context, flagSet *ffclient.GoFeatureFlag) {
	flagSet.ForceRefreshWithContext(httpContext.Request().Context())
}

// validateAndUpsertFlag validates a raw flag definition and upserts it through the writable
// retriever, triggering a flagSet reload on success. Shared by PUT (full body) and PATCH
// (merge-patch result), which only differ in how they arrive at raw and how they report the
// created/updated outcome. On any failure it writes the error response itself and returns a
// non-nil error that the caller should return as-is.
func validateAndUpsertFlag(
	httpContext echo.Context,
	threadContext context.Context,
	flagSet *ffclient.GoFeatureFlag,
	writableRetriever retriever.WritableRetriever,
	flagKey string,
	rawFlagDefinition []byte,
) (definitionDTO dto.DTO, isFlagCreated bool, flagETag string, writeErr error) {
	definitionDTO, validationErr := validateDefinition(rawFlagDefinition)
	if validationErr != nil {
		return dto.DTO{}, false, "", writeError(httpContext, http.StatusBadRequest, flag.ErrorFlagConfiguration, validationErr.Error())
	}

	isFlagCreated, flagETag, upsertErr := writableRetriever.UpsertFlag(threadContext, flagKey, rawFlagDefinition, ifMatchHeader(httpContext))
	if upsertErr != nil {
		status, code := mapWriteErr(upsertErr)
		return dto.DTO{}, false, "", writeError(httpContext, status, code, upsertErr.Error())
	}

	reload(httpContext, flagSet)
	return definitionDTO, isFlagCreated, flagETag, nil
}

// finalizeFlagUpdate finalizes a successful flag update (PATCH merge-patch or PATCH
// .../state): records update metrics, builds the response envelope, sets the ETag header,
// and writes 200 OK. Shared by patchFlag and setFlagState, which both only ever report an
// update (never a create) once they reach this point.
func finalizeFlagUpdate(
	httpContext echo.Context, metrics metric.Metrics, key string, definitionDTO dto.DTO, source, etag string,
) error {
	metrics.IncFlagUpdated(key)
	metrics.IncFlagChange()

	flagResponse, buildResponseErr := buildFlagResponse(key, definitionDTO, source, true, etag)
	if buildResponseErr != nil {
		return writeError(httpContext, http.StatusInternalServerError, flag.ErrorCodeGeneral, buildResponseErr.Error())
	}
	httpContext.Response().Header().Set("ETag", etag)
	return httpContext.JSON(http.StatusOK, flagResponse)
}
