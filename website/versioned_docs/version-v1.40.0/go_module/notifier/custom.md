---
sidebar_position: 30
---

# Custom Notifier

To create a custom notifier you must have a `struct` that implements the
[`notifier.Notifier`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/notifier/notifier) interface.

In parameter you will receive a `notifier.DiffCache` struct that will tell you what has changed in your flag configuration.

```go
import (
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/notifier/notifier"
	"sync"
)

type Notifier struct{}
func (c *Notifier) Notify(diff notifier.DiffCache) error {
	// ...
	// do whatever you want here
}
```
