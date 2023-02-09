package flag

// ResolutionReason is an enum following the open-feature specs about resolution reasons.
type ResolutionReason = string

const (
	// ReasonTargetingMatch Indicates that the feature flag is targeting
	// 100% of the targeting audience,
	// e.g. 100% rollout percentage
	ReasonTargetingMatch ResolutionReason = "TARGETING_MATCH"

	// ReasonSplit Indicates that the feature flag is targeting
	// a subset of the targeting audience,
	// e.g. less than 100% rollout percentage
	ReasonSplit ResolutionReason = "SPLIT"

	// ReasonDisabled Indicates that the feature flag is disabled
	ReasonDisabled ResolutionReason = "DISABLED"

	// ReasonDefault Indicates that the feature flag evaluated to the default value
	ReasonDefault ResolutionReason = "DEFAULT"

	// ReasonStatic	Indicates that the feature flag evaluated to a
	// static value, for example, the default value for the flag
	//
	// Note: Typically means that no dynamic evaluation has been
	// executed for the feature flag
	ReasonStatic ResolutionReason = "STATIC"

	// ReasonUnknown Indicates an unknown issue occurred during evaluation
	ReasonUnknown ResolutionReason = "UNKNOWN"

	// ReasonError Indicates that an error occurred during evaluation
	// Note: The `errorCode`-field contains the details of this error
	ReasonError ResolutionReason = "ERROR"

	ReasonOffline ResolutionReason = "OFFLINE"
)
