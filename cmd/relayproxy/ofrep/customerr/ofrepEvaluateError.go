package customerr

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type OfrepEvaluateError struct {
	err model.OFREPEvaluateErrorResponse
}

func NewOFREPEvaluateError(flagKey string, errorCode flag.ErrorCode, errorDetails string) *OfrepEvaluateError {
	return &OfrepEvaluateError{
		err: model.OFREPEvaluateErrorResponse{
			Key: flagKey,
			OFREPErrorResponse: model.OFREPErrorResponse{
				ErrorCode:    errorCode,
				ErrorDetails: errorDetails,
			},
		},
	}
}

func (m *OfrepEvaluateError) Error() string {
	return fmt.Sprintf("missing TargetingKey error: %v", m.err)
}

func (m *OfrepEvaluateError) ToOFRErrorResponse() model.OFREPEvaluateErrorResponse {
	return m.err
}
