# Google Cloud PubSub

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data inside a GCP PubSub topic.

<!-- -->

info

Google Cloud PubSub<!-- --> is an exporter of type queue, it means that it send events as soon as he receives them it does not work in bulk but in near real time.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Google Cloud PubSub<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: pubsub

    projectID: "my-project-id"

    topic: "goff-feature-events"

# ...
```

| Field name  | Mandatory | Type   | Default  | Description                                                                                               |
| ----------- | --------- | ------ | -------- | --------------------------------------------------------------------------------------------------------- |
| `kind`      | **        | string | **none** | **Value should be `pubsub`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `projectID` | **        | string | **none** | Value should be ID of GCP project you are using.                                                          |
| `topic`     | **        | string | **none** | Topic name on which messages will be published.                                                           |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Google Cloud PubSub<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

Migration from v1

If you are migrating from the legacy `pubsubexporter` (v1), simply change the import and struct name from `pubsubexporter.Exporter` to `pubsubexporterv2.Exporter`. The v2 exporter uses `cloud.google.com/go/pubsub/v2` which provides improved performance and additional features.

example.go

```
config := ffclient.Config{

   // ...

    DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &pubsubexporterv2.Exporter{

            ProjectID: "project-id", // required

            Topic: "topic", // required

            Options: []option.ClientOption{...},

            PublishSettings: &pubsub.PublishSettings{...},

            EnableMessageOrdering: true,

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field                   | Mandatory | Description                                                                                                 |
| ----------------------- | --------- | ----------------------------------------------------------------------------------------------------------- |
| `ProjectID`             | **        | ID of GCP project.<br />*You can find it in your GCP console*                                               |
| `Topic`                 | **        | Name of topic on which messages will be published                                                           |
| `Options`               | **        | PubSub client options *(see [docs](https://pkg.go.dev/google.golang.org/api/option))*                       |
| `PublishSettings`       | **        | Publisher related settings *(see [docs](https://pkg.go.dev/cloud.google.com/go/pubsub/v2#PublishSettings))* |
| `EnableMessageOrdering` | **        | Enables delivery of ordered keys                                                                            |
