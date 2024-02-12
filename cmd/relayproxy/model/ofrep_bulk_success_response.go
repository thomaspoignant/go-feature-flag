package model

type OFREPBulkEvaluateSuccessResponse = []OFREPFlagBulkEvaluateSuccessResponse

type OFREPFlagBulkEvaluateSuccessResponse struct {
	OFREPEvaluateSuccessResponse
	ErrorCode    string `json:"errorCode"`
	ErrorDetails string `json:"errorDetails"`
}
