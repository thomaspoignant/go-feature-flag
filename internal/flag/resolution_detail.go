package flag

type ResolutionDetails struct {
	// Variant indicates the name of the variant used when evaluating the flag
	Variant string

	// Reason indicates the reason of the decision
	Reason ResolutionReason

	// ErrorCode indicates the error code for this evaluation
	ErrorCode ErrorCode
}
