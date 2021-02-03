package notifier

import (
	"sync"

	"github.com/thomaspoignant/go-feature-flag/internal/model"
)

type Notifier interface {
	Notify(cache model.DiffCache, waitGroup *sync.WaitGroup)
}
