package model

type OFREPBulkEvaluateSuccessResponse = []OFREPFlagBulkEvaluateSuccessResponse

type OFREPFlagBulkEvaluateSuccessResponse struct {
	OFREPEvaluateSuccessResponse `json:"OFREPEvaluateSuccessResponse"`
	ErrorCode                    string `json:"errorCode,omitempty"`
	ErrorDetails                 string `json:"errorDetails,omitempty"`
	ETag                         string `json:"ETag"`
}
