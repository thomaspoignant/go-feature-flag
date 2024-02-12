package customerr

import (
	"fmt"
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
)

type OfrepGenericError struct {
	err model.OFREPErrorResponse
}

func NewOFREPGenericError(errorCode string, errorDetails string) *OfrepGenericError {
	return &OfrepGenericError{
		err: model.OFREPErrorResponse{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		},
	}
}

func (m *OfrepGenericError) Error() string {
	return fmt.Sprintf("missing TargetingKey error: %v", m.err)
}

func (m *OfrepGenericError) ToOFRErrorResponse() model.OFREPErrorResponse {
	return m.err
}
