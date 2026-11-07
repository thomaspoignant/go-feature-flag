# Apache Kafka

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data inside a Kafka topic.

<!-- -->

info

Apache Kafka<!-- --> is an exporter of type queue, it means that it send events as soon as he receives them it does not work in bulk but in near real time.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Apache Kafka<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: kafka

    kafka:

      topic: "go-feature-flag-events"

      addresses:

        - "localhost:9092"

# ...
```

| Field name        | Mandatory | Type      | Default           | Description                                                                                                                                                                                                                                                                                                |
| ----------------- | --------- | --------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`            | **        | string    | **none**          | **Value should be `kafka`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                                                                                                   |
| `kafka.topic`     | **        | string    | **none**          | Kafka topic to bind to.                                                                                                                                                                                                                                                                                    |
| `kafka.addresses` | **        | \[]string | **none**          | List of bootstrap addresses for the Kafka cluster.                                                                                                                                                                                                                                                         |
| `kafka.config`    | **        | object    | *see description* | This field allows fine tuning of the Kafka reader. This object should contain the [Sarama configuration](https://pkg.go.dev/github.com/IBM/sarama#Config) that the reader will use. On empty, a sensible default is created using [sarama.NewConfig()](https://pkg.go.dev/github.com/IBM/sarama#NewConfig) |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Apache Kafka<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
config := ffclient.Config{

   // ...

   DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &kafkaexporter.Exporter{

           Settings: kafkaexporter.Settings{

              Topic: "go-feature-flag-events",

              Addresses: []string{"localhost:9092", "cluster2:9092"},

           },

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field      | Mandatory | Description                                                                                                                                                                                               |
| ---------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Format`   | **        | The format to produce in the topic (only `JSON` is supported for now).                                                                                                                                    |
| `Settings` | **        | A extension of the `*sarama.Config` struct that holds additional settings for the producer, such as timeouts, TLS settings, etc. If not populated, a default will be used by calling `sarama.NewConfig()` |

In the settings we have the following fields:

| Field       | Mandatory | Description                                          |
| ----------- | --------- | ---------------------------------------------------- |
| `Topic`     | **        | Name of the topic to publish messages                |
| `Addresses` | **        | The list of addresses for the Kafka boostrap servers |
