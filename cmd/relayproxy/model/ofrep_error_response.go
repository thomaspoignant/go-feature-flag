package model

type OFREPEvaluateErrorResponse struct {
	OFREPErrorResponse `json:",inline" yaml:",inline" toml:",inline"`
	Key                string `json:"key"`
}

type OFREPErrorResponse struct {
	ErrorCode    string `json:"errorCode"`
	ErrorDetails string `json:"errorDetails"`
}
