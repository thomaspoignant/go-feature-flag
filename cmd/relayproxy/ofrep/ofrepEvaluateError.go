package ofrep

import (
	"github.com/thomaspoignant/go-feature-flag/cmd/relayproxy/model"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

func NewEvaluateError(
	key string,
	errorCode flag.ErrorCode,
	errorDetails string,
) *model.OFREPEvaluateResponseError {
	return &model.OFREPEvaluateResponseError{
		OFREPCommonResponseError: model.OFREPCommonResponseError{
			ErrorCode:    errorCode,
			ErrorDetails: errorDetails,
		},
		Key: key,
	}
}

func NewOFREPCommonError(
	errorCode flag.ErrorCode,
	errorDetails string,
) *model.OFREPCommonResponseError {
	return &model.OFREPCommonResponseError{
		ErrorCode:    errorCode,
		ErrorDetails: errorDetails,
	}
}
