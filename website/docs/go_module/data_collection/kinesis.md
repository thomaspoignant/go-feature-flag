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
| `PartitionKey `     | (Required) Key kinesis use to identify which stream shard data belongs to                                                                                                                                     |
| `ExplicitPartitionKey `     | (Optional) Key kinesis use to identify which stream shard data belongs to overrides PartitionKey setting                                                                                                                                     |
| `AwsConfig `    | (Optional) An instance of `*aws.Config` that holds additional settings connect to AWS |                                                                                                                                         |                                                                                                                                                     |
