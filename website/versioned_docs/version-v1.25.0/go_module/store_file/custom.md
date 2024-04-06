---
sidebar_position: 30
---

# Custom Retriever

## Simple retriever
To create a custom retriever you must have a `struct` that implements the [`Retriever`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/#Retriever) interface.

```go showLineNumbers
type Retriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}
```

The `Retrieve` 	function is supposed to load the file and to return a `[]byte` of your flag configuration file.

You can check existing `Retriever` *([file](https://github.com/thomaspoignant/go-feature-flag/blob/main/retriever/fileretriever/retriever.go),
[s3](https://github.com/thomaspoignant/go-feature-flag/blob/main/retriever/s3retriever/retriever.go), ...)* to have an idea on how to do build your own.

## Initializable retriever
Sometimes you need to initialize your retriever before using it.
For example, if you want to connect to a database, you need to initialize the connection before using it.

To help you with that, you can use the [`InitializableRetriever`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/#InitializableRetriever) interface.

The only difference with the `Retriever` interface is that the `Init` func of your retriever will be called at the start of the application and the `Shutdown` func will be called when closing GO Feature Flag.

```go
type InitializableRetriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
	Init(ctx context.Context, logger *log.Logger) error
	Shutdown(ctx context.Context) error
	Status() retriever.Status
}
```
To avoid any issue to call the `Retrieve` function before the `Init` function, you have to manage the status of your retriever.
GO Feature Flag will try to call the `Retrieve` function only if the status is `RetrieverStatusReady`.
