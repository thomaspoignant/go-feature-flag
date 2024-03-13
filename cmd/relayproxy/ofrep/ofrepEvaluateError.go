package ofrep

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

type EvaluateError struct {
	err model.OFREPCommonErrorResponse
}

func NewEvaluateError(errorCode flag.ErrorCode, errorDetails string) *EvaluateError {
	return &EvaluateError{
		err: model.OFREPCommonErrorResponse{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		},
	}
}

func (m *EvaluateError) Error() string {
	return fmt.Sprintf("missing TargetingKey error: %v", m.err)
}

func (m *EvaluateError) ToOFRErrorResponse() model.OFREPCommonErrorResponse {
	return m.err
}
