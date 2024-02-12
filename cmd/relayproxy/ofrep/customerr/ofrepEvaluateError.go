package customerr

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type OfrepEvaluateError struct {
	err model.OFREPEvaluateErrorResponse
}

func NewOFREPEvaluateError(flagKey string, errorCode string, errorDetails string) *OfrepEvaluateError {
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
