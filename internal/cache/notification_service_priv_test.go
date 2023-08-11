//go:build !race

package cache_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/thomaspoignant/go-feature-flag/internal/cache"
	"github.com/thomaspoignant/go-feature-flag/internal/flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"github.com/thomaspoignant/go-feature-flag/testutils"
	"github.com/thomaspoignant/go-feature-flag/testutils/testconvert"
	"log"
	"os"
	"testing"
	"time"
)

func Test_notificationService_callNotifier(t *testing.T) {
	n := &NotifierMock{}
	c := cache.NewNotificationService([]notifier.Notifier{n})
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
	c := cache.NewNotificationService([]notifier.Notifier{n})
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
	tempFile, _ := os.CreateTemp("", "tempFile")
	logger := log.New(tempFile, "", 0)
	n := &NotifierMock{WithError: true}
	c := cache.NewNotificationService([]notifier.Notifier{n})
	oldCache := map[string]flag.Flag{
		"yo": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	newCache := map[string]flag.Flag{
		"yo-new": &flag.InternalFlag{Version: testconvert.String("1.0")},
	}
	c.Notify(oldCache, newCache, logger)
	time.Sleep(100 * time.Millisecond)

	content, _ := os.ReadFile(tempFile.Name())
	assert.Regexp(t, "\\["+testutils.RFC3339Regex+"\\] error while calling the notifier: error\n", string(content))
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
