# Getting started
First, you need to initialize the `ffclient` with the location of your backend file.
```go
err := ffclient.Init(ffclient.Config{
    PollInterval: 3,
    Retriever: &ffclient.HTTPRetriever{
        URL:    "http://example.com/flag-config.yaml",
    },
})
defer ffclient.Close()
```
*This example will load a file from an HTTP endpoint and will refresh the flags every 3 seconds (if you omit the
PollInterval, the default value is 60 seconds).*

Now you can evaluate your flags anywhere in your code.

```go
user := ffuser.NewUser("user-unique-key")
hasFlag, _ := ffclient.BoolVariation("test-flag", user, false)
if hasFlag {
    // flag "test-flag" is true for the user
} else {
    // flag "test-flag" is false for the user
}
```
You can find more example programs in the [examples/](examples) directory.
