---
sidebar_position: 6
---

# Kinesis Exporter
The **Kinesis exporter** produces messages to a Kinesis stream for each event generated. Currently all messages are populated into single shard of your choice.

## Configuration example
```go
ffclient.Config{ 
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
```

## Configuration fields
| Field        | Description                                                                                                                                                                                              |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `StreamName `     | (Required) Name of the Kinesis stream to publish messages to                                                                                                                                                           |
| `PartitionKey `     | (Optional) Function that takes 'FeatureEvent' as an input and returns calculated string 'Partition Key'. If not specified, then by default string 'default' will be used for all events. Effectively this will assign all events to a single Kinesis shard. There is nothing bad in using single shard, but for performance consideration this field might be utilised.                                                                                                    |
| `ExplicitPartitionKey `     | (Optional) String key to identify which stream shard event data belongs to. Overrides PartitionKey setting                                                                                                                                     |
| `AwsConfig `    | (Optional) An instance of `*aws.Config` that holds additional settings connect to AWS |                                                                                                                                         |                                                                                                                                                     |
