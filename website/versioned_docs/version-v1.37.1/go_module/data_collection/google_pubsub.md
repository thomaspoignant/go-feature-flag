---
sidebar_position: 7
---

# PubSub Exporter

The **PubSub exporter** will collect the data and publish an event on the topic for each evaluation we receive.

## Configuration example
```go
ffclient.Config{ 
   // ... 
    cfg, _ := config.LoadDefaultConfig(context.TODO())
    DataExporter: ffclient.DataExporter{
        // ... 
        Exporter: &pubsubexporter.Exporter{
            ProjectID: "project-id", // required
            Topic: "topic", // required 
            Options: []option.ClientOption{...},
            PublishSettings: &pubsub.PublishSettings{...},
            EnableMessageOrdering: true,
        },
    },
    // ...
}
```

## Configuration fields

| Field                   | Description                                                                                          |
|-------------------------|------------------------------------------------------------------------------------------------------|
| `ProjectID `            | ID of GCP project.<br/>_You can find it in your GCP console_                                         |
| `Topic `                | Name of topic on which messages will be published                                                    |
| `Options`               | PubSub client options *(see [docs](https://pkg.go.dev/google.golang.org/api/option))*                |
| `PublishSettings`       | Topic related settings *(see [docs](https://pkg.go.dev/cloud.google.com/go/pubsub#PublishSettings))* |
| `EnableMessageOrdering` | Enables delivery of ordered keys                                                                     |

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/pubsubexporter).
