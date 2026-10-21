package flaghistory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

// toMap converts a flag into a generic map to be able to compare 2 versions of the same flag.
func toMap(f flag.Flag) (map[string]any, error) {
	raw, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("impossible to marshal the flag: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("impossible to unmarshal the flag: %w", err)
	}
	return m, nil
}

// hashFlag returns a stable hash of the flag.
// encoding/json sorts the map keys, so the result does not depend on the order of the fields.
func hashFlag(m map[string]any) (string, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:]), nil
}

// summarize gives the minimal description stored for an added or deleted flag.
func summarize(m map[string]any) map[string]any {
	summary := map[string]any{}
	if variations, ok := m["variations"].(map[string]any); ok {
		summary["variations"] = slices.Sorted(maps.Keys(variations))
	}
	if disable, ok := m["disable"].(bool); ok {
		summary["disable"] = disable
	}
	return summary
}

// diffMaps lists the fields that differ between before and after, recursively.
func diffMaps(path string, before, after map[string]any) []FieldChange {
	keys := make(map[string]struct{}, len(before)+len(after))
	for k := range before {
		keys[k] = struct{}{}
	}
	for k := range after {
		keys[k] = struct{}{}
	}

	changes := make([]FieldChange, 0)
	for _, k := range slices.Sorted(maps.Keys(keys)) {
		p := k
		if path != "" {
			p = path + "." + k
		}
		b, inBefore := before[k]
		a, inAfter := after[k]
		bm, bIsMap := b.(map[string]any)
		am, aIsMap := a.(map[string]any)
		switch {
		case inBefore && inAfter && bIsMap && aIsMap:
			changes = append(changes, diffMaps(p, bm, am)...)
		case !reflect.DeepEqual(b, a):
			changes = append(changes, FieldChange{Path: p, Old: b, New: a})
		}
	}
	return changes
}

// EntriesFromDiff converts the diff computed by the notification service into history entries.
func EntriesFromDiff(flagset, source string, diff notifier.DiffCache) ([]Entry, error) {
	entries := make([]Entry, 0, len(diff.Added)+len(diff.Deleted)+len(diff.Updated))
	base := baseEntry(flagset, source, diff.Change)

	for _, key := range slices.Sorted(maps.Keys(diff.Added)) {
		m, err := toMap(diff.Added[key])
		if err != nil {
			return nil, err
		}
		hash, err := hashFlag(m)
		if err != nil {
			return nil, err
		}
		e := base
		e.FlagKey, e.Action, e.AfterHash = key, ActionAdded, hash
		e.Diff = Diff{Summary: summarize(m)}
		entries = append(entries, e)
	}

	for _, key := range slices.Sorted(maps.Keys(diff.Deleted)) {
		m, err := toMap(diff.Deleted[key])
		if err != nil {
			return nil, err
		}
		e := base
		e.FlagKey, e.Action = key, ActionDeleted
		e.Diff = Diff{Summary: summarize(m)}
		entries = append(entries, e)
	}

	for _, key := range slices.Sorted(maps.Keys(diff.Updated)) {
		before, err := toMap(diff.Updated[key].Before)
		if err != nil {
			return nil, err
		}
		after, err := toMap(diff.Updated[key].After)
		if err != nil {
			return nil, err
		}
		hash, err := hashFlag(after)
		if err != nil {
			return nil, err
		}
		e := base
		e.FlagKey, e.Action, e.AfterHash = key, ActionUpdated, hash
		e.Diff = Diff{Changes: diffMaps("", before, after)}
		entries = append(entries, e)
	}
	return entries, nil
}

func baseEntry(flagset, source string, change *notifier.ChangeInfo) Entry {
	e := Entry{Flagset: flagset, Source: source, Author: UnknownAuthor}
	if change == nil {
		return e
	}
	if change.Author != "" {
		e.Author = change.Author
	}
	e.AuthorEmail = change.Email
	e.ChangeRef = change.Reference
	return e
}
