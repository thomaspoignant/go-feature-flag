---
sidebar_position: 0
---

# Export data

If you want to export data about how your flags are used, you can use the **`DataExporter`**.  
It collects all the variations events and can save these events on several locations:

- [File](file.md) *- create local files with the variation usages.*
- [Log](log.md) *- use your logger to write the variation usages.*
- [S3](s3.md) *- export your variation usages to S3.*
- [Webhook](webhook.md) *- export your variation usages by calling a webhook.*
- [Google Cloud Storage](google_cloud_storage.md) *- export your variation usages by calling a webhook.*
- [Kafka](kafka.md) *- export your variation usages by producing messages to a Kafka topic.*

If the existing exporter does not work with your system you can extend the system and use a [custom exporter](custom.md).

## Data format

Currently, we are supporting only feature events.  
They represent individual flag evaluations and are considered "full fidelity" events.

### Example

```json showLineNumbers
{
    "kind": "feature",
    "contextKind": "anonymousUser",
    "userKey": "ABCD",
    "creationDate": 1618228297,
    "key": "test-flag",
    "variation": "Default",
    "value": false,
    "default": false,
    "source": "SERVER"
}
```

### Configuration fields

| Field              | Description                                                                                                                                                                                                                                                                                             |
|--------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **`kind`**         | The kind for a feature event is feature. A feature event will only be generated if the trackEvents attribute of the flag is set to true.                                                                                                                                                                |
| **`contextKind`**  | The kind of context which generated an event. This will only be "**anonymousUser**" for events generated on behalf of an anonymous user or the reserved word "**user**" for events generated on behalf of a non-anonymous user.                                                                          |
| **`userKey`**      | The key of the user object used in a feature flag evaluation.                                                                                                                                                                                                                                           |
| **`creationDate`** | When the feature flag was requested at Unix epoch time in milliseconds.                                                                                                                                                                                                                                 |
| **`key`**          | The key of the feature flag requested.                                                                                                                                                                                                                                                                  |
| **`variation`**    | The variation of the flag requested. Available values are:<br/>**True**: if the flag was evaluated to True <br/>**False**: if the flag was evaluated to False<br/>**Default**: if the flag was evaluated to Default<br/>**SdkDefault**: if something wrong happened and the SDK default value was used. |
| **`value`**        | The value of the feature flag returned by feature flag evaluation.                                                                                                                                                                                                                                      |
| **`source`**       | Where the event is generated. This is set to SERVER when the event is evaluated from the relay-proxy and PROVIDER_CACHE when it is evaluated from the cache.             
| **`default`**      | (Optional) This value is set to true if feature flag evaluation failed, in which case, the value returned is the default value passed to variation.                                                                                                                                                     |

Events are collected and send in bulk to avoid spamming your exporter *(see details in [how to configure data export](#how-to-configure-data-export)*)

## How to configure data export?

In your `ffclient.Config` add the `DataExporter` field and configure your export location.

To avoid spamming your location everytime you have a variation called, `go-feature-flag` is storing in memory all the events and sends them in bulk to the exporter.
You can decide the threshold on when to send the data with the properties `FlushInterval` and `MaxEventInMemory`. The first threshold hit will export the data.

If there are some flags that you don't want to export, you can use `trackEvents` fields on these specific flags to disable the data export *(see [flag file format](../../configure_flag/flag_format.mdx))*.

### Example

```go showLineNumbers
ffclient.Config{ 
    // ...
   DataExporter: ffclient.DataExporter{
        FlushInterval:   10 * time.Second,
        MaxEventInMemory: 1000,
        Exporter: &fileexporter.Exporter{
            OutputDir: "/output-data/",
        },
    },
    // ...
}
```

### Configuration fields

| Field              | Description |
|--------------------|-------------|
| `Exporter`         | The configuration of the exporter you want to use. All the exporters are available in the `exporter` package.                          |
| `FlushInterval`    | *(optional)*<br/>Time to wait before exporting the data.<br/>**Default: 60 seconds**.                                                  |
| `MaxEventInMemory` | *(optional)*<br/>If `MaxEventInMemory` is reach before the `FlushInterval` a intermediary export will be done<br/>**Default: 100000**. |

## Don't track a flag

By default, all flags are trackable, and their data is exported.

If you want to exclude a specific flag from the data export, you can set the property `trackEvents` to `false` on your flag, and you will have no export for it.

### YAML

```yaml
test-flag:
  percentage: 50
  true: "B"
  false: "A"
  default: "Default"
  trackEvents: false
```

### JSON

<details>
  <summary>JSON example</summary>

```json
{
  "test-flag": {
    "percentage": 50,
    "true": "B",
    "false": "A",
    "default": "Default",
    # highlight-next-line
    "trackEvents": false
  }
}
```

</details>

### TOML

<details>
  <summary>TOML example</summary>

```toml
[test-flag]
percentage = 50.0
true = "B"
false = "A"
default = "Default"
# highlight-next-line
trackEvents = false
```

</details>
