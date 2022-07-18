package flag

// ErrorCode is an enum following the open-feature specs about error code.
type ErrorCode = string

const (
	ErrorCodeProviderNotReady ErrorCode = "PROVIDER_NOT_READY"
	ErrorCodeFlagNotFound     ErrorCode = "FLAG_NOT_FOUND"
	ErrorCodeParseError       ErrorCode = "PARSE_ERROR"
	ErrorCodeTypeMismatch     ErrorCode = "TYPE_MISMATCH"
	ErrorCodeGeneral          ErrorCode = "GENERAL"
)
