package model

// OFREPSSEEventTypeRefetchEvaluation tells the provider that the flag configuration has changed
// and that it must re-fetch its evaluations.
const OFREPSSEEventTypeRefetchEvaluation = "refetchEvaluation"

// OFREPSSEEvent is the payload of the events sent on the OFREP SSE stream,
// as defined by OpenFeature ADR-0008.
type OFREPSSEEvent struct {
	// Type of the event, providers must ignore the types they don't know.
	Type string `json:"type" example:"refetchEvaluation"`
	// LastModified is the unix timestamp (in seconds) of the flag configuration change.
	LastModified int64 `json:"lastModified,omitempty" example:"1771622898"`
}
