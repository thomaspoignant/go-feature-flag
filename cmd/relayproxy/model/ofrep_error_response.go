package model

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

type OFREPEvaluateResponseError struct {
	OFREPCommonResponseError `       json:",inline" yaml:",inline" toml:",inline"`
	Key                      string `json:"key"`
}

type OFREPCommonResponseError struct {
	ErrorCode    flag.ErrorCode `json:"errorCode"`
	ErrorDetails string         `json:"errorDetails"`
}

func (o *OFREPCommonResponseError) Error() string {
	return fmt.Sprintf("[%s] %s", o.ErrorCode, o.ErrorDetails)
}
