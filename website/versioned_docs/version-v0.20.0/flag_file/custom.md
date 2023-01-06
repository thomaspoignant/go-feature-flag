# Custom Retriever

To create a custom retriever you must have a `struct` that implements the [`Retriever`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Retriever) interface.

```go linenums="1"
type Retriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}
```

The `Retrieve` 	function is supposed to load the file and to return a []byte of your flag configuration file.

If you want to specify the format of the file, you can use the `ffclient.Config.FileFormat` option to specify if it is 
a `YAML`, `JSON` or `TOML` file.

You can check existing `Retriever` *([file](https://github.com/thomaspoignant/go-feature-flag/blob/main/retriever_file.go),
[s3](https://github.com/thomaspoignant/go-feature-flag/blob/main/retriever_s3.go), ...)* to have an idea on how to do build your own.
