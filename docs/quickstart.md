# Getting started

## Installation
```bash
go get github.com/thomaspoignant/go-feature-flag
```

## SDK Initialisation

First, you need to initialize the `ffclient` with the location of your backend file.
```go linenums="1"
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds (if you omit the
PollingInterval, the default value is 60 seconds).*

## Evaluate your flags
Now you can evaluate your flags anywhere in your code.

```go linenums="1"
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```
You can find more examples in the [examples/](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples) directory.
