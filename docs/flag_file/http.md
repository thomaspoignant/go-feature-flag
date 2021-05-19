# HTTP endpoint
The [**HTTPRetriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#HTTPRetriever) will perform an HTTP Request with your configuration to get your flags.

## Example
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
        Timeout: 2 * time.Second,
    },
})
defer ffclient.Close()
```
## Configuration fields
To configure your HTTP endpoint:

| Field | Description |
|---|---|
|**`URL`**| location of your file.|
|**`Method`**| the HTTP method you want to use <br>*(default is GET)*.|
|**`Body`**| *(optional)*<br>If you need a body to get the flags.|
|**`Header`**| *(optional)*<br>Header you should pass while calling the endpoint *(useful for authorization)*.|
|**`Timeout`**| *(optional)*<br>Timeout for the HTTP call <br>(default is 10 seconds).|
