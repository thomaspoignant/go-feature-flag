# 🚚 Export evaluation data

The module supports different ways of exporting evaluation data.

To retrieve a file you need to provide an [exporter](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/) in your `ffclient.Config{}` during the initialization.

## Available exporters:[​](#available-exporters "Direct link to Available exporters:")

* [AWS S3](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-s3)
* [MinIO](https://gofeatureflag.org/../integrations/export-evaluation-data/minio)
* [Azure Blob Storage](https://gofeatureflag.org/../integrations/export-evaluation-data/azure-blob-storage)
* [File System](https://gofeatureflag.org/../integrations/export-evaluation-data/file)
* [Apache Kafka](https://gofeatureflag.org/../integrations/export-evaluation-data/kafka)
* [AWS Kinesis](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-kinesis)
* [Google Cloud Storage](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-storage)
* [Google Cloud PubSub](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-pubsub)
* [Google Cloud BigQuery](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-bigquery)
* [AWS SQS](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-sqs)
* [Webhook](https://gofeatureflag.org/../integrations/export-evaluation-data/webhook)
* [Log](https://gofeatureflag.org/../integrations/export-evaluation-data/log)
* [OpenTelemetry](https://gofeatureflag.org/../integrations/export-evaluation-data/opentelemetry)

info

In each exporter documentation, you will find the configuration needed to set up the exporter.
