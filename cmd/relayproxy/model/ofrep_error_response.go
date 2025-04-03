package model

import (
	"fmt"

	"github.com/thomaspoignant/go-feature-flag/internal/flag"
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
