package stream

import "github.com/thomaspoignant/go-feature-flag/notifier"

type options struct {
	includeFlagDetails bool
}

// Option configures flag change streams.
type Option func(*options)

// WithFlagDetails controls whether stream payloads include flag definitions.
func WithFlagDetails(include bool) Option {
	return func(options *options) {
		options.includeFlagDetails = include
	}
}

func newOptions(opts ...Option) options {
	options := options{includeFlagDetails: true}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

type flagChangeNames struct {
	Deleted map[string]struct{} `json:"deleted"`
	Added   map[string]struct{} `json:"added"`
	Updated map[string]struct{} `json:"updated"`
}

func flagChangePayload(diff notifier.DiffCache, includeFlagDetails bool) any {
	if includeFlagDetails {
		return diff
	}
	return flagChangeNames{
		Deleted: mapKeys(diff.Deleted),
		Added:   mapKeys(diff.Added),
		Updated: mapKeys(diff.Updated),
	}
}

func mapKeys[T any](values map[string]T) map[string]struct{} {
	keys := make(map[string]struct{}, len(values))
	for key := range values {
		keys[key] = struct{}{}
	}
	return keys
}
