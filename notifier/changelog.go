package notifier

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/luci/go-render/render"
	"github.com/r3labs/diff/v3"
)

// FieldChange is a single field that changed in a flag, ready to be displayed in a notification.
type FieldChange struct {
	// Path is the dot-separated path of the field that changed.
	Path string
	// From is the readable value before the change, "null" if the field did not exist.
	From string
	// To is the readable value after the change, "null" if the field was removed.
	To string
}

// FlagChanges returns every field that changed between two versions of a flag,
// sorted by path, including fields that were added or removed.
func FlagChanges(before, after any) []FieldChange {
	changelog, _ := diff.Diff(before, after, diff.AllowTypeMismatch(true))
	changes := make([]FieldChange, 0, len(changelog))
	for _, change := range changelog {
		changes = append(changes, FieldChange{
			Path: strings.Join(change.Path, "."),
			From: ReadableValue(change.From),
			To:   ReadableValue(change.To),
		})
	}
	sort.SliceStable(changes, func(i, j int) bool { return changes[i].Path < changes[j].Path })
	return changes
}

// ReadableValue renders a value for humans: JSON for anything JSON can encode,
// so pointers and time internals are not shown, and Go syntax otherwise.
func ReadableValue(value any) string {
	if out, err := json.Marshal(value); err == nil {
		return string(out)
	}
	return render.Render(value)
}
