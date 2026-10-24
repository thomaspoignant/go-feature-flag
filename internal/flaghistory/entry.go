package flaghistory

import "time"

// Action is the kind of modification made on a flag.
type Action string

const (
	ActionAdded   Action = "added"
	ActionUpdated Action = "updated"
	ActionDeleted Action = "deleted"
)

// UnknownAuthor is stored when no retriever is able to tell who made the change.
const UnknownAuthor = "unknown"

// FieldChange is a modification of a single field of a flag.
type FieldChange struct {
	Path string `json:"path"`
	Old  any    `json:"old,omitempty"`
	New  any    `json:"new,omitempty"`
}

// Diff is the only thing we store about a modification, never the full flag.
type Diff struct {
	// Changes is the list of fields modified (updated flags).
	Changes []FieldChange `json:"changes,omitempty"`
	// Summary is a minimal description of the flag (added and deleted flags).
	Summary map[string]any `json:"summary,omitempty"`
	// Reconciled is true when the change was detected at startup, by comparing the flags
	// with the latest recorded state, so the detail of the fields is not available.
	Reconciled bool `json:"reconciled,omitempty"`
}

// Entry is a line of the history.
type Entry struct {
	ID          int64     `json:"id"`
	Flagset     string    `json:"flagset"`
	FlagKey     string    `json:"flagKey"`
	Action      Action    `json:"action"`
	Diff        Diff      `json:"diff"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"authorEmail,omitempty"`
	ChangeRef   string    `json:"changeRef,omitempty"`
	Source      string    `json:"source,omitempty"`
	DetectedAt  time.Time `json:"detectedAt"`

	// AfterHash is the hash of the flag after the change (empty for a deletion).
	// It is used to avoid recording the same change several times.
	AfterHash string `json:"-"`
}

// Filter are the criteria to list the history.
type Filter struct {
	Flagset string
	FlagKey string
	Action  Action
	Author  string
	From    time.Time
	To      time.Time
	// Limit is the maximum number of entries returned.
	Limit int
	// Cursor is the id of the last entry of the previous page (0 for the first page).
	Cursor int64
}

// Page is a page of the history.
type Page struct {
	Entries []Entry `json:"entries"`
	// NextCursor is the cursor to use to get the next page, 0 when there is no more entry.
	NextCursor int64 `json:"nextCursor,omitempty"`
}
