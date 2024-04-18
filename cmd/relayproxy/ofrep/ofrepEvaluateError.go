package ofrep

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
)

func NewEvaluateError(key string, errorCode flag.ErrorCode, errorDetails string) *model.OFREPEvaluateErrorResponse {
	return &model.OFREPEvaluateErrorResponse{
		OFREPCommonErrorResponse: model.OFREPCommonErrorResponse{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		},
		Key: key,
	}
}

func NewOFREPCommonError(errorCode flag.ErrorCode, errorDetails string) *model.OFREPCommonErrorResponse {
	return &model.OFREPCommonErrorResponse{
		ErrorCode:    errorCode,
		ErrorDetails: errorDetails,
	}
}
