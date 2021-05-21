# File
The [**FileRetriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#FileRetriever) will read a local file to get your flags.

!!! tip
    Using a file to store your flags is not recommend, except if it is in a shared folder for all your services.

## Example
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &ffclient.FileRetriever{
        Path: "file-example.yaml",
    },
})
defer ffclient.Close()
```

## Configuration fields
To configure your File retriever:

| Field | Description |
|---|---|
|**`Path`**| location of your file on the file system.|
