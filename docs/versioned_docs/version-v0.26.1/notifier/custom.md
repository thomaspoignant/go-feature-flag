# Custom Notifier
To create a custom notifier you must have a `struct` that implements the [`notifier.Notifier`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/notifier/notifier) interface.

```go
import (
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/notifier/notifier"
	"sync"
)

type Notifier struct{}
func (c *Notifier) Notify(diff notifier.DiffCache, wg *sync.WaitGroup) error {
	defer waitGroup.Done() // don't forget this line, if you don't have it you can break your notifications
	
	// ...
	// do whatever you want here
}
```
