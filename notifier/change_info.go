package notifier

import "time"

// ChangeInfo describes who made a change on the flag configuration and when.
// It is filled by the retrievers able to provide this information (git providers, databases...).
type ChangeInfo struct {
	// Author is the name of the person (or the system) who made the change.
	Author string `json:"author,omitempty"`
	// Email is the email of the author.
	Email string `json:"email,omitempty"`
	// Reference identifies the change in the source system (commit sha, version...).
	Reference string `json:"reference,omitempty"`
	// ChangedAt is the date of the change in the source system.
	ChangedAt time.Time `json:"changedAt"`
}
