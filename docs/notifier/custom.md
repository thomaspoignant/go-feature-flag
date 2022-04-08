# Custom Notifier
To create a custom notifier you must have a `struct` that implements the [`ffnotifier.notifier`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/ffnotifier/notifier) interface.

```go linenums="1"
import (
	"fmt"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffnotifier"
	"sync"
)

// Your config object to create a new Notifier
type CustomNotifierConfig struct{}

// GetNotifier returns a notfier that implement ffnotifier.Notifier
func (c *CustomNotifier) GetNotifier(config ffclient.Config) (ffnotifier.Notifier, error) {
	return &CustomNotifier{}, nil
}

type CustomNotifier struct{}
func (w *CustomNotifier) Notify(cache ffnotifier.DiffCache, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done() // don't forget this line, if you don't have it you can break your notifications
	
	// ...
	// do whatever you want here
}
```
