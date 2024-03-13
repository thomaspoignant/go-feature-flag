package ofrep

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type OfrepEvaluateError struct {
	err model.OFREPCommonErrorResponse
}

func NewOFREPEvaluateError(errorCode flag.ErrorCode, errorDetails string) *OfrepEvaluateError {
	return &OfrepEvaluateError{
		err: model.OFREPCommonErrorResponse{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		},
	}
}

func (m *OfrepEvaluateError) Error() string {
	return fmt.Sprintf("missing TargetingKey error: %v", m.err)
}

func (m *OfrepEvaluateError) ToOFRErrorResponse() model.OFREPCommonErrorResponse {
	return m.err
}
