//go:build !race

package notification_test

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thejerf/slogassert"
	"github.com/thomaspoignant/go-feature-flag/internal/notification"
	"github.com/thomaspoignant/go-feature-flag/modules/core/flag"
	"github.com/thomaspoignant/go-feature-flag/modules/core/testutils/testconvert"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/utils/fflog"
)

func Test_notificationService_callNotifier(t *testing.T) {
	n := &NotifierMock{}
	c := notification.NewService([]notifier.Notifier{n})
	oldCache := map[string]flag.Flag{
		"yo": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	newCache := map[string]flag.Flag{
		"yo-new": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	c.Notify(oldCache, newCache, nil)
	time.Sleep(20 * time.Millisecond)
	assert.True(t, n.HasBeenCalled)
}

func Test_notificationService_no_difference(t *testing.T) {
	n := &NotifierMock{}
	c := notification.NewService([]notifier.Notifier{n})
	oldCache := map[string]flag.Flag{
		"yo": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	newCache := map[string]flag.Flag{
		"yo": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	c.Notify(oldCache, newCache, nil)
	time.Sleep(20 * time.Millisecond)
	assert.False(t, n.HasBeenCalled)
}

func Test_notificationService_with_error(t *testing.T) {
	handler := slogassert.New(t, slog.LevelDebug, nil)
	logger := slog.New(handler)
	n := &NotifierMock{WithError: true}
	c := notification.NewService([]notifier.Notifier{n})
	oldCache := map[string]flag.Flag{
		"yo": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	newCache := map[string]flag.Flag{
		"yo-new": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	c.Notify(oldCache, newCache, &fflog.FFLogger{LeveledLogger: logger})
	time.Sleep(100 * time.Millisecond)

	handler.AssertMessage("error while calling the notifier")
	assert.False(t, n.HasBeenCalled)
}

type NotifierMock struct {
	WithError     bool
	HasBeenCalled bool
}

func (n *NotifierMock) Notify(cache notifier.DiffCache) error {
	if n.WithError {
		return fmt.Errorf("error")
	}
	n.HasBeenCalled = true
	return nil
}
