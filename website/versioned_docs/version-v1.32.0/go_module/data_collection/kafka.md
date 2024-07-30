---
sidebar_position: 6
---

# Kafka Exporter
The **Kafka exporter** produces messages to a Kafka topic for each event generated.

## Configuration example
```go
ffclient.Config{ 
   // ...
   DataExporter: ffclient.DataExporter{
        // ...
        Exporter: &kafkaexporter.Exporter{
           Settings: kafkaexporter.Settings{
              Topic: "go-feature-flag-events",
              Addresses: []string{"cluster1", "cluster2"},
           },
        },
    },
    // ...
}
```

## Configuration fields
| Field        | Description                                                                                                                                                                                              |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Topic `     | Name of the topic to publish messages                                                                                                                                                           |
| `Addresses ` | The list of addresses for the Kafka boostrap servers                                                                                                                                                     |
| `Config `    | (Optional) An instance of `*sarama.Config` that holds additional settings for the producer, such as timeouts, TLS settings, etc. If not populated, a default will be used by calling `sarama.NewConfig()` |                                                                                                                                         |                                                                                                                                                     |

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/kafkaexporter).
