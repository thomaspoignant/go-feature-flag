# AWS SQS

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data inside a AWS SQS queue.

<!-- -->

info

AWS SQS<!-- --> is an exporter of type queue, it means that it send events as soon as he receives them it does not work in bulk but in near real time.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->AWS SQS<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: sqs

    queueUrl: "https://sqs.us-east-1.amazonaws/XXXX/test-queue"

# ...
```

| Field name | Mandatory | Type   | Default  | Description                                                                                            |
| ---------- | --------- | ------ | -------- | ------------------------------------------------------------------------------------------------------ |
| `kind`     | **        | string | **none** | **Value should be `sqs`**.<br />*This field is mandatory and describes which retriever you are using.* |
| `queueUrl` | **        | string | **none** | URL of your SQS queue.<br />*You can find it in your AWS console.*                                     |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->AWS SQS<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
awsConfig, _ := config.LoadDefaultConfig(context.Background())

config := ffclient.Config{

   // ...

   DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &sqsexporter.Exporter{

            QueueURL: "https://sqs.eu-west-1.amazonaws.com/XXX/test-queue",

            AwsConfig: &awsConfig,

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field       | Mandatory | Description                                                                                                                                                                                |
| ----------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `QueueURL`  | **        | URL of your SQS queue.<br />*You can find it in your AWS console.*                                                                                                                         |
| `AwsConfig` | **        | An instance of `aws.Config` that configures your access to AWS *(see [this documentation for more info](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html))*. |
