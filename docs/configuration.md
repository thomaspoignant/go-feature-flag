!!! Danger 
    TODO Update links
# Configuration
`go-feature-flag` needs to be initialized to be used.  
During the initialization you must give a [`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) configuration object.  
All available fields are describe in the [configuration fields](#configuration-fields) section.

[`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) is the only location where you can put the configuration.

## Configuration fields

| Field | Description |
|---|---|
|`Retriever`  | The configuration retriever you want to use to get your flag file<br> *see [Store your flag file](flag_file/index.md) for the configuration details*.|
|`Context`  | *(optional)*<br>The context used by the retriever.<br />Default: `context.Background()`|
|`DataExporter` | *(optional)*<br>DataExporter defines how to export data on how your flags are used.<br> *see [export data section](#export-data) for more details*.|
|`FileFormat`| *(optional)*<br>Format of your configuration file. Available formats are `yaml`, `toml` and `json`, if you omit the field it will try to unmarshal the file as a `yaml` file.<br>Default: `YAML`|
|`Logger`   | *(optional)*<br>Logger used to log what `go-feature-flag` is doing.<br />If no logger is provided the module will not log anything.<br>Default: No log|
|`Notifiers` | *(optional)*<br>List of notifiers to call when your flag file has changed.<br> *see [notifiers section](#notifiers) for more details*.|
|`PollInterval`   | (optional) Number of seconds to wait before refreshing the flags.<br />Default: 60|
|`StartWithRetrieverError` | *(optional)*<br>If **true**, the SDK will start even if we did not get any flags from the retriever. It will serve only default values until the retriever returns the flags.<br>The init method will not return any error if the flag file is unreachable.<br>Default: **false**|

## Example
```go linenums="1"
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


