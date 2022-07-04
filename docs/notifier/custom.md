# Custom Notifier
To create a custom notifier you must have a `struct` that implements the [`notifier.Notifier`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/notifier/notifier) interface.

```go linenums="1"
import (
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/notifier"
	"sync"
)

// Your config object to create a new Notifier
type CustomNotifierConfig struct{}

// GetNotifier returns a notfier that implement notifier.Notifier
func (c *CustomNotifierConfig) GetNotifier(config ffclient.Config) (notifier.Notifier, error) {
	return &CustomNotifier{}, nil
}

type CustomNotifier struct{}
func (w *CustomNotifier) Notify(cache notifier.DiffCache, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done() // don't forget this line, if you don't have it you can break your notifications
	
	// ...
	// do whatever you want here
}
```
