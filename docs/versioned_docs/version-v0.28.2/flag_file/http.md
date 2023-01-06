# HTTP endpoint

The [__HTTP Retriever__](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/retriever/httpretriever/#Retriever)
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

| Field         | Description                                                                                                     |
|---------------|-----------------------------------------------------------------------------------------------------------------|
| __`URL`__     | Location where to retrieve the file <br/> _(ex: [http://mydomain.io/flag.yaml](http://mydomain.io/flag.yaml))_. |
| __`Method`__  | the HTTP method you want to use <br/>_(default is GET)_.                                                        |
| __`Body`__    | _(optional)_<br/>If you need a body to get the flags.                                                           |
| __`Header`__  | _(optional)_<br/>Header you should pass while calling the endpoint _(useful for authorization)_.                |
| __`Timeout`__ | _(optional)_<br/>Timeout for the HTTP call <br/>(default is 10 seconds).                                        |
