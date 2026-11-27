# AWS Kinesis

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data inside a Kafka Kinesis stream.

<!-- -->

info

AWS Kinesis<!-- --> is an exporter of type queue, it means that it send events as soon as he receives them it does not work in bulk but in near real time.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->AWS Kinesis<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: kinesis

    streamArn: "arn:aws:kinesis:us-east-1:XXXX:stream/test-stream"

# ...
```

| Field name   | Mandatory | Type   | Default  | Description                                                                                                |
| ------------ | --------- | ------ | -------- | ---------------------------------------------------------------------------------------------------------- |
| `kind`       | **        | string | **none** | Value should be **`kinesis`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `streamArn`  | **        | string | **none** | The ARN of your kinesis stream.                                                                            |
| `streamName` | **        | string | **none** | The name of your kinesis stream.                                                                           |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->AWS Kinesis<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
awsConfig, _ := config.LoadDefaultConfig(context.Background())

config := ffclient.Config{

   // ...

   DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &kinesisexporter.Exporter{

            Settings: kinesisexporter.NewSettings(

                kinesisexporter.WithStreamName("test-stream"),

                kinesisexporter.WithPartitionKey("0"),

                kinesisexporter.WithExplicitHashKey("0"),

            ),

            AwsConfig: &config, // aws custom configuration

        },



    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field                  | Mandatory | Description                                                                                                                                                                                                                                                                                                                                                  |
| ---------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `StreamName`           | **        | Name of the Kinesis stream to publish messages to                                                                                                                                                                                                                                                                                                            |
| `PartitionKey`         | **        | Function that takes 'FeatureEvent' as an input and returns calculated string 'Partition Key'. If not specified, then by default string 'default' will be used for all events. Effectively this will assign all events to a single Kinesis shard. There is nothing bad in using single shard, but for performance consideration this field might be utilised. |
| `ExplicitPartitionKey` | **        | String key to identify which stream shard event data belongs to. Overrides PartitionKey setting                                                                                                                                                                                                                                                              |
| `AwsConfig`            | **        | An instance of `*aws.Config` that holds additional settings connect to AWS                                                                                                                                                                                                                                                                                   |
