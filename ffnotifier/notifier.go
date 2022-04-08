package ffnotifier

import (
	"sync"
)

type Notifier interface {
	Notify(cache DiffCache, waitGroup *sync.WaitGroup)
}
