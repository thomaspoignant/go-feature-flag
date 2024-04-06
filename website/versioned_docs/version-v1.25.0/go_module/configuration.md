---
sidebar_position: 10
description: How to configure the GO module to use it directly in your code. 
---

# Configuration
`go-feature-flag` needs to be initialized to be used.  
During the initialization you must give a [`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) configuration object.  

[`ffclient.Config{}`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#Config) is the only location where you can put the configuration.

## Configuration fields

| Field                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Retriever`                   | The configuration retriever you want to use to get your flag file<br/> *See [Store your flag file](./store_file/index.md) for the configuration details*.<br /><br /> *This field is optional if `Retrievers`* is configured.                                                                                                                                                                                                                                                                  |
| `Retrievers`                  | `Retrievers` is exactly the same thing as `Retriever` but you can configure more than 1 source for your flags.<br/>All flags are retrieved in parallel, but we are applying them in the order you provided them _(it means that a flag can be overridden by another flag)_. <br/>*See [Store your flag file](./store_file/index.md) for the configuration details*. <br /><br /> *This field is optional if `Retrievers`* is configured.                                                       |
| `Context`                     | *(optional)*<br/>The context used by the retriever.<br />Default: **`context.Background()`**                                                                                                                                                                                                                                                                                                                                                                                                   |
| `Environment`                 | <a name="option_environment"></a>*(optional)*<br/>The environment the app is running under, can be checked in feature flag rules.<br />Default: `""`<br/>*Check [**"environments"** section](../configure_flag/flag_format/#environments) to understand how to use this parameter.*                                                                                                                                                                                                            |
| `DataExporter`                | *(optional)*<br/>DataExporter defines the method for exporting data on the usage of your flags.<br/> *see [export data section](data_collection/index.md) for more details*.                                                                                                                                                                                                                                                                                                                              |
| `FileFormat`                  | *(optional)*<br/>Format of your configuration file. Available formats are `yaml`, `toml` and `json`, if you omit the field it will try to unmarshal the file as a `yaml` file.<br/>Default: **`YAML`**                                                                                                                                                                                                                                                                                         |
| `Logger`                      | *(optional)*<br/>Logger is used to log what `go-feature-flag` is doing.<br />If no logger is provided the module will not log anything.<br/>Default: **No log**                                                                                                                                                                                                                                                                                                                                   |
| `Notifiers`                   | *(optional)*<br/>List of notifiers to call when your flag file has been changed.<br/> *See [notifiers section](./notifier/index.md) for more details*.                                                                                                                                                                                                                                                                                                                                         |
| `PollingInterval`             | (optional) Duration to wait before refreshing the flags.<br/>The minimum polling interval is 1 second.<br/>Default: **60 * time.Second**                                                                                                                                                                                                                                                                                                                                                       |
| `EnablePollingJitter`         | (optional) Set to true if you want to avoid having true periodicity when retrieving your flags. It is useful to avoid having spike on your flag configuration storage in case your application is starting multiple instance at the same time.<br/>We ensure a deviation that is maximum ±10% of your polling interval.<br />Default: **false**                                                                                                                                          |
| `StartWithRetrieverError`     | *(optional)* If **true**, the SDK will start even if we did not get any flags from the retriever. It will serve only default values until the retriever returns the flags.<br/>The init method will not return any error if the flag file is unreachable.<br/>Default: **false**                                                                                                                                                                                                               |
| `Offline`                     | *(optional)* If **true**, the SDK will not try to retrieve the flag file and will not export any data. No notifications will be sent either.<br/>Default: **false**                                                                                                                                                                                                                                                                                                                            |
| `EvaluationContextEnrichment` | *(optional)* It is a free `map[string]interface{}` field that will be merged with the evaluation context sent during the evaluations. It is useful to add common attributes to all the evaluation, such as a server version, environment, ...<br/>All those fields will be included in the custom attributes of the evaluation context.<br/>If in the evaluation context you have a field with the same name, it will be overriden by the `evaluationContextEnrichment`.<br/> Default: **nil** |

## Example
```go
ffclient.Init(ffclient.Config{ 
    PollingInterval:   3 * time.Second,
    Logger:         log.New(file, "/tmp/log", 0),
    Context:        context.Background(),
    Environment:    os.Getenv("MYAPP_ENV"),
    Retriever:      &fileretriever.Retriever{Path: "testdata/flag-config.goff.yaml"},
    FileFormat:     "yaml",
    Notifiers: []notifier.Notifier{
        &webhooknotifier.Notifier{
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
        Exporter: &file.Exporter{
            OutputDir: "/output-data/",
        },
    },
    StartWithRetrieverError: false,
})
```

## Multiple configuration flag files
`go-feature-flag` comes ready to use out of the box by calling the `Init` function and, it will be available everywhere.  
Since most applications will want to use a single central flag configuration, the package provides this. It is similar to a singleton.

In all the examples above, they demonstrate using `go-feature-flag` in its singleton style approach.

### Working with multiple go-feature-flag

You can also create many `go-feature-flag` clients to use in your application.  

Each will have its own unique set of configurations and flags. Each can read from a different config file and from different places.  
All the functions that `go-feature-flag` package supports are mirrored as methods on a [`GoFeatureFlag`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#GoFeatureFlag).

#### Example

```go showLineNumbers
x, err := ffclient.New(Config{ Retriever: &httpretriever.Retriever{{URL: "http://example.com/flag-config.goff.yaml",}})
defer x.Close()

y, err := ffclient.New(Config{ Retriever: &httpretriever.Retriever{{URL: "http://example.com/test2.goff.yaml",}})
defer y.Close()

user := ffcontext.NewEvaluationContext("user-key")
x.BoolVariation("test-flag", user, false)
y.BoolVariation("test-flag", user, false)

// ...
```

When working with multiple [`GoFeatureFlag`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#GoFeatureFlag), it is up to the user to keep track of different [`GoFeatureFlag`](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag#GoFeatureFlag) instances.

## Offline mode
In some situations, you might want to stop making remote calls and fall back to default values for your feature flags.  
For example, if your software is both cloud-hosted and distributed to customers to run on-premise, it might make sense 
to fall back to defaults when running on-premise.

You can do this by setting `Offline` mode in the client's Config.

## Advanced configuration

- [Export data from your flag variations](./data_collection/index.md)
- [Be notified when your flags change](./notifier/index.md)
