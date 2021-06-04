# Custom exporter
To create a custom exporter you must have a `struct` that implements the [`exporter.Exporter`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/internal/exporter#Exporter) interface.


```go linenums="1"
type Exporter interface {
    // Export will send the data to the exporter.
    Export(context.Context, *log.Logger, []FeatureEvent) error

	// IsBulk return false if we should directly send the data as soon as it is produce
	// and true if we collect the data to send them in bulk.
	IsBulk() bool
}

```
`Export` is called asynchronously with a list of `FeatureEvent` that have been collected.  
It is your responsibility to store them where you want.

`IsBulk` function should return `false` if the exporter can handle the results in stream mode.  
If you decide to manage it in streaming mode, everytime we call a variation the `Export` function will be called
with only on event in the list.
