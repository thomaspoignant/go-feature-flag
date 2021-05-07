package model

// DiffCache contains the changes made in the cache, to be able
// to notify the user that something has changed (logs, webhook ...)
type DiffCache struct {
	Deleted map[string]Flag        `json:"deleted"`
	Added   map[string]Flag        `json:"added"`
	Updated map[string]DiffUpdated `json:"updated"`
}

// HasDiff check if we have differences
func (d *DiffCache) HasDiff() bool {
	return len(d.Deleted) > 0 || len(d.Added) > 0 || len(d.Updated) > 0
}

type DiffUpdated struct {
	Before Flag `json:"old_value"`
	After  Flag `json:"new_value"`
}
