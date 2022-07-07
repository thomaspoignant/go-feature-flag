# HTTP endpoint
The [**HTTP Retriever**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/httpretriever/#Retriever) 
will perform an HTTP Request with your configuration to get your flags.

## Example
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &httpretriever.Retriever{
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
|**`URL`**| Location where to retrieve the file <br> _(ex: http://mydomain.io/flag.yaml)_.|
|**`Method`**| the HTTP method you want to use <br>*(default is GET)*.|
|**`Body`**| *(optional)*<br>If you need a body to get the flags.|
|**`Header`**| *(optional)*<br>Header you should pass while calling the endpoint *(useful for authorization)*.|
|**`Timeout`**| *(optional)*<br>Timeout for the HTTP call <br>(default is 10 seconds).|
