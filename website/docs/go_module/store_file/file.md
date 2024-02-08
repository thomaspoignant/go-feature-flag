---
sidebar_position: 25
---

# File
The [**File Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/fileretriever/#Retriever) will read a local file to get your flags.

:::tip
Using a file to store your flags is not recommend, except if it is in a shared folder for all your services.
:::

## Example
```go showLineNumbers
import 	"github.com/thomaspoignant/go-feature-flag/retriever/file"
// ...

err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &fileretriever.Retriever{
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
