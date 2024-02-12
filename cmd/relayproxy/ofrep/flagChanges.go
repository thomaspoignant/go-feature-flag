package ofrep

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/metric"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/ofrep/customerr"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"hash/crc32"
	"net/http"
)

type ofrepFlagChangesCtrl struct {
	goFF    *ffclient.GoFeatureFlag
	metrics metric.Metrics
}

func NewOFREPFlagChanges(goFF *ffclient.GoFeatureFlag, metrics metric.Metrics) Controller {
	return &ofrepFlagChangesCtrl{
		goFF:    goFF,
		metrics: metrics,
	}
}

type flagChangesResponse struct {
	ETag string `json:"ETag"`
	Key  string `json:"Key"`
}

// OFREPHandler is the entry point to get the ETag in bulk of flags to implement OpenFeature Remote Evaluation Protocol.
// @Summary     Compute all the ETag in bulk using the OpenFeature Remote Evaluation Protocol
// @Description Making a **POST** request to the URL `/ofrep/v1/flag/changes` will give you the value of the list of
// @Description ETags for your feature flags.
// @Description
// @Description If no flags are provided, the API will compite all available flags in the configuration.
// @Security     ApiKeyAuth
// @Produce      json
// @Accept	 	 json
// @Param 		 data body []string true "List of flags to evaluate"
// @Success      200  {object} model.OFREPBulkEvaluateSuccessResponse "Success"
// @Failure      400 {object}  model.OFREPErrorResponse "Bad Request"
// @Failure      401 {object}  modeldocs.HTTPErrorDoc "Unauthorized"
// @Failure      500 {object}  modeldocs.HTTPErrorDoc "Internal server error"
// @Router       /ofrep/v1/flag/changes [post]
func (h *ofrepFlagChangesCtrl) OFREPHandler(c echo.Context) error {
	request := new([]string)
	if err := c.Bind(request); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			customerr.NewOFREPGenericError("INVALID_CONTEXT", err.Error()).ToOFRErrorResponse())
	}

	flags, err := h.goFF.GetFlagsFromCache()
	if err != nil {
		return c.JSON(http.StatusInternalServerError,
			customerr.NewOFREPGenericError("INTERNAL_ERROR", err.Error()).ToOFRErrorResponse())
	}

	if request == nil || len(*request) == 0 {
		// no flag as input, we compute the ET
		response := []flagChangesResponse{}
		for k, v := range flags {
			response = append(response, flagChangesResponse{
				ETag: fmt.Sprintf("%x", flagCheckSum(v)),
				Key:  k,
			})
		}
		return c.JSON(http.StatusOK, response)
	}

	response := []flagChangesResponse{}
	for _, key := range *request {
		if f, ok := flags[key]; ok {
			response = append(response, flagChangesResponse{
				ETag: fmt.Sprintf("%x", flagCheckSum(f)),
				Key:  key,
			})
		}
	}
	return c.JSON(http.StatusOK, response)
}

func flagCheckSum(f flag.Flag) string {
	jsonData, _ := json.Marshal(f)
	return fmt.Sprintf("%x", crc32.ChecksumIEEE(jsonData))
}
