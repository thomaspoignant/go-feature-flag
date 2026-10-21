package retriever

import (
	"context"
	"errors"
)

// NewChangeInfoResolver returns a function asking every retriever implementing ChangeInfoProvider
// who made the last change, and keeps the most recent one.
// Retrievers in error are ignored as long as one of them answers; it returns nil if none can answer.
func NewChangeInfoResolver(retrievers []Retriever) func(ctx context.Context) (*ChangeInfo, error) {
	providers := make([]ChangeInfoProvider, 0)
	for _, r := range retrievers {
		if p, ok := r.(ChangeInfoProvider); ok {
			providers = append(providers, p)
		}
	}
	if len(providers) == 0 {
		return nil
	}
	return func(ctx context.Context) (*ChangeInfo, error) {
		var latest *ChangeInfo
		var errs []error
		for _, p := range providers {
			info, err := p.LastChange(ctx)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if info != nil && (latest == nil || info.ChangedAt.After(latest.ChangedAt)) {
				latest = info
			}
		}
		if latest == nil && len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
		return latest, nil
	}
}
