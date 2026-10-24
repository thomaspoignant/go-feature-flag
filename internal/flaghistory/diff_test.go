package flaghistory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

func newFlag(version string, disable bool) *flag.InternalFlag {
	variations := map[string]*any{"on": testconvert.Interface(true), "off": testconvert.Interface(false)}
	return &flag.InternalFlag{
		Variations:  &variations,
		DefaultRule: &flag.Rule{VariationResult: testconvert.String("on")},
		Version:     testconvert.String(version),
		Disable:     testconvert.Bool(disable),
	}
}

func TestEntriesFromDiff(t *testing.T) {
	tests := []struct {
		name        string
		diff        notifier.DiffCache
		wantActions []Action
		wantKeys    []string
		check       func(t *testing.T, entries []Entry)
	}{
		{
			name: "added flag stores a summary, not the flag",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{"new-flag": newFlag("1", false)},
			},
			wantActions: []Action{ActionAdded},
			wantKeys:    []string{"new-flag"},
			check: func(t *testing.T, entries []Entry) {
				assert.Equal(t, []string{"off", "on"}, entries[0].Diff.Summary["variations"])
				assert.Equal(t, false, entries[0].Diff.Summary["disable"])
				assert.Empty(t, entries[0].Diff.Changes)
				assert.NotEmpty(t, entries[0].AfterHash)
			},
		},
		{
			name: "deleted flag has no hash",
			diff: notifier.DiffCache{
				Deleted: map[string]flag.Flag{"old-flag": newFlag("1", false)},
			},
			wantActions: []Action{ActionDeleted},
			wantKeys:    []string{"old-flag"},
			check: func(t *testing.T, entries []Entry) {
				assert.Empty(t, entries[0].AfterHash)
			},
		},
		{
			name: "updated flag lists only the modified fields",
			diff: notifier.DiffCache{
				Updated: map[string]notifier.DiffUpdated{
					"f": {Before: newFlag("1", false), After: newFlag("2", true)},
				},
			},
			wantActions: []Action{ActionUpdated},
			wantKeys:    []string{"f"},
			check: func(t *testing.T, entries []Entry) {
				assert.Equal(t, []FieldChange{
					{Path: "disable", Old: false, New: true},
					{Path: "version", Old: "1", New: "2"},
				}, entries[0].Diff.Changes)
				assert.Nil(t, entries[0].Diff.Summary)
			},
		},
		{
			name: "author comes from the change info",
			diff: notifier.DiffCache{
				Added:  map[string]flag.Flag{"f": newFlag("1", false)},
				Change: &notifier.ChangeInfo{Author: "jane", Email: "jane@x.io", Reference: "abc"},
			},
			wantActions: []Action{ActionAdded},
			wantKeys:    []string{"f"},
			check: func(t *testing.T, entries []Entry) {
				assert.Equal(t, "jane", entries[0].Author)
				assert.Equal(t, "jane@x.io", entries[0].AuthorEmail)
				assert.Equal(t, "abc", entries[0].ChangeRef)
			},
		},
		{
			name: "author is unknown without change info",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{"f": newFlag("1", false)},
			},
			wantActions: []Action{ActionAdded},
			wantKeys:    []string{"f"},
			check: func(t *testing.T, entries []Entry) {
				assert.Equal(t, UnknownAuthor, entries[0].Author)
			},
		},
		{
			name: "entries are sorted by flag key",
			diff: notifier.DiffCache{
				Added: map[string]flag.Flag{"b": newFlag("1", false), "a": newFlag("1", false)},
			},
			wantActions: []Action{ActionAdded, ActionAdded},
			wantKeys:    []string{"a", "b"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries, err := EntriesFromDiff("default", "instance-1", tt.diff)
			require.NoError(t, err)
			require.Len(t, entries, len(tt.wantKeys))
			for i := range entries {
				assert.Equal(t, tt.wantKeys[i], entries[i].FlagKey)
				assert.Equal(t, tt.wantActions[i], entries[i].Action)
				assert.Equal(t, "default", entries[i].Flagset)
				assert.Equal(t, "instance-1", entries[i].Source)
			}
			if tt.check != nil {
				tt.check(t, entries)
			}
		})
	}
}

func TestHashFlag_stable(t *testing.T) {
	m1, err := toMap(newFlag("1", false))
	require.NoError(t, err)
	m2, err := toMap(newFlag("1", false))
	require.NoError(t, err)
	m3, err := toMap(newFlag("2", false))
	require.NoError(t, err)

	h1, _ := hashFlag(m1)
	h2, _ := hashFlag(m2)
	h3, _ := hashFlag(m3)
	assert.Equal(t, h1, h2)
	assert.NotEqual(t, h1, h3)
}
