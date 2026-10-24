package notification_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
)

type captureNotifier struct {
	mu    sync.Mutex
	diffs []notifier.DiffCache
}

func (n *captureNotifier) Notify(d notifier.DiffCache) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.diffs = append(n.diffs, d)
	return nil
}

func Test_notificationService_changeInfo(t *testing.T) {
	info := &notifier.ChangeInfo{Author: "jane", Reference: "abc123"}
	tests := []struct {
		name         string
		resolver     notification.ChangeInfoResolver
		sameCache    bool
		wantChange   *notifier.ChangeInfo
		wantResolved bool
	}{
		{
			name: "author added to the diff",
			resolver: func(context.Context) (*notifier.ChangeInfo, error) {
				return info, nil
			},
			wantChange:   info,
			wantResolved: true,
		},
		{
			name: "resolver error does not block the notification",
			resolver: func(context.Context) (*notifier.ChangeInfo, error) {
				return nil, fmt.Errorf("api down")
			},
			wantChange:   nil,
			wantResolved: true,
		},
		{
			name: "resolver not called without difference",
			resolver: func(context.Context) (*notifier.ChangeInfo, error) {
				return info, nil
			},
			sameCache:    true,
			wantResolved: false,
		},
		{
			name:       "no resolver configured",
			resolver:   nil,
			wantChange: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved := false
			var opts []notification.Option
			if tt.resolver != nil {
				r := tt.resolver
				opts = append(opts, notification.WithChangeInfoResolver(
					func(ctx context.Context) (*notifier.ChangeInfo, error) {
						resolved = true
						return r(ctx)
					}))
			}
			n := &captureNotifier{}
			svc := notification.NewService([]notifier.Notifier{n}, opts...)
			oldCache := map[string]flag.Flag{"f": &flag.InternalFlag{Version: testconvert.String("1")}}
			newCache := map[string]flag.Flag{"f": &flag.InternalFlag{Version: testconvert.String("2")}}
			if tt.sameCache {
				newCache = oldCache
			}
			svc.Notify(oldCache, newCache, nil)
			svc.Close()

			assert.Equal(t, tt.wantResolved, resolved)
			if tt.sameCache {
				assert.Empty(t, n.diffs)
				return
			}
			assert.Len(t, n.diffs, 1)
			assert.Equal(t, tt.wantChange, n.diffs[0].Change)
		})
	}
}

func Test_notificationService_changeInfoTimeoutDeadline(t *testing.T) {
	var deadlineSet bool
	svc := notification.NewService([]notifier.Notifier{&captureNotifier{}},
		notification.WithChangeInfoResolver(func(ctx context.Context) (*notifier.ChangeInfo, error) {
			_, deadlineSet = ctx.Deadline()
			return nil, nil
		}))
	svc.Notify(
		map[string]flag.Flag{"f": &flag.InternalFlag{Version: testconvert.String("1")}},
		map[string]flag.Flag{"f": &flag.InternalFlag{Version: testconvert.String("2")}},
		nil,
	)
	svc.Close()
	assert.True(t, deadlineSet)
}
