package cache

import "sync"

type Notifier interface {
	Notify(cache diffCache, waitGroup *sync.WaitGroup)
}
