---
sidebar_position: 6
---

# SQS Exporter

The **SQS exporter** will collect the data and create an event in the queue for each evaluation we receive.

## Configuration example
```go
ffclient.Config{ 
   // ...
	 cfg, _ := config.LoadDefaultConfig(context.TODO())
   DataExporter: ffclient.DataExporter{
        // ...
        Exporter: &sqsexporter.Exporter{
			      QueueURL: "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",
            AwsConfig: &cfg,
        },
    },
    // ...
}
```

## Configuration fields
| Field         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `QueueURL `     | URL of your SQS queue.<br/>_You can find it in your AWS console._                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `AwsConfig `  | An instance of `aws.Config` that configure your access to AWS *(see [this documentation for more info](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html))*.                                                                                                                                                                                                                                                                                                                                                          |

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/sqsexporter).
