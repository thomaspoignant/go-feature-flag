package model

import "github.com/thomaspoignant/go-feature-flag/internal/flag"

type OFREPEvaluateErrorResponse struct {
	OFREPErrorResponse `json:",inline" yaml:",inline" toml:",inline"`
	Key                string `json:"key"`
}

type OFREPErrorResponse struct {
	ErrorCode    flag.ErrorCode `json:"errorCode"`
	ErrorDetails string         `json:"errorDetails"`
}
