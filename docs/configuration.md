# Configuration
The configuration is set with `ffclient.Config{}` and you can give it to ``ffclient.Init()`` the initialization
function.  
All the possible options are listed [here](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config).

## Example
```go 
ffclient.Init(ffclient.Config{ 
    PollInterval:   3,
    Logger:         log.New(file, "/tmp/log", 0),
    Context:        context.Background(),
    Retriever:      &ffclient.FileRetriever{Path: "testdata/flag-config.yaml"},
    FileFormat:     "yaml",
    Notifiers: []ffclient.NotifierConfig{
        &ffclient.WebhookConfig{
            EndpointURL: " https://example.com/hook",
            Secret:     "Secret",
            Meta: map[string]string{
                "app.name": "my app",
            },
        },
    },
    DataExporter: ffclient.DataExporter{
        FlushInterval:   10 * time.Second,
        MaxEventInMemory: 1000,
        Exporter: &ffexporter.File{
            OutputDir: "/output-data/",
        },
    },
    StartWithRetrieverError: false,
})
```

## Configuration fields

| Field | Description |
|---|---|
|`Retriever`  | The configuration retriever you want to use to get your flag file<br> *see [Where do I store my flags file](#where-do-i-store-my-flags-file) for the configuration details*.|
|`Context`  | *(optional)* The context used by the retriever.<br />Default: `context.Background()`|
|`DataExporter` | *(optional)* DataExporter defines how to export data on how your flags are used.<br> *see [export data section](#export-data) for more details*.|
|`FileFormat`| *(optional)* Format of your configuration file. Available formats are `yaml`, `toml` and `json`, if you omit the field it will try to unmarshal the file as a `yaml` file.<br>Default: `YAML`|
|`Logger`   | *(optional)* Logger used to log what `go-feature-flag` is doing.<br />If no logger is provided the module will not log anything.<br>Default: No log|
|`Notifiers` | *(optional)* List of notifiers to call when your flag file has changed.<br> *see [notifiers section](#notifiers) for more details*.|
|`PollInterval`   | (optional) Number of seconds to wait before refreshing the flags.<br />Default: 60|
|`StartWithRetrieverError` | *(optional)* If **true**, the SDK will start even if we did not get any flags from the retriever. It will serve only default values until the retriever returns the flags.<br>The init method will not return any error if the flag file is unreachable.<br>Default: **false**|