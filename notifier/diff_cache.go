package notifier

import (
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
)

// DiffCache contains the changes made in the cache, to be able
// to notify the user that something has changed (logs, webhook ...)
type DiffCache struct {
	Deleted map[string]flag.Flag   `json:"deleted"`
	Added   map[string]flag.Flag   `json:"added"`
	Updated map[string]DiffUpdated `json:"updated"`
	// Change is the origin of the modification (author, reference...).
	// It is nil when none of the retrievers is able to provide it.
	Change *ChangeInfo `json:"change,omitempty"`
}

// HasDiff check if we have differences
func (d *DiffCache) HasDiff() bool {
	return len(d.Deleted) > 0 || len(d.Added) > 0 || len(d.Updated) > 0
}

type DiffUpdated struct {
	Before flag.Flag `json:"old_value"`
	After  flag.Flag `json:"new_value"`
}
