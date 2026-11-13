# MinIO

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data to MinIO, or any other S3-compatible object store such as Ceph or Garage. There is no dedicated MinIO exporter: you use the S3 exporter and point it at your own server.

<!-- -->

[MinIO](https://min.io) is an S3-compatible object store you can run yourself, on-premises or in your own cloud. Everything below uses the [AWS S3 exporter](https://gofeatureflag.org/docs/integrations/export-evaluation-data/aws-s3.md) with `kind: s3` — only the endpoint changes.

Everytime the `flushInterval` or `maxEventInMemory` is reached, a new file is added to your MinIO bucket.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

The exporter configuration is the regular `s3` one, MinIO is selected through the AWS SDK environment variables:

goff-proxy.yaml

```
# ...

exporters:

  - kind: s3

    bucket: evaluation-data-bucket

    path: /go-feature-flag/variations/

# ...
```

```
export AWS_ENDPOINT_URL_S3=http://127.0.0.1:9000

export AWS_ACCESS_KEY_ID=minioadmin

export AWS_SECRET_ACCESS_KEY=minioadmin

export AWS_DEFAULT_REGION=us-east-1
```

Only `kind` and `bucket` are mandatory. All the other fields *(`format`, `filename`, `path`, `flushInterval`, `maxEventInMemory`, `csvTemplate`, `parquetCompressionCodec`)* behave exactly as documented on the [AWS S3 exporter page](https://gofeatureflag.org/docs/integrations/export-evaluation-data/aws-s3.md) — MinIO changes nothing about them.

Use an IP address in `AWS_ENDPOINT_URL_S3`, or read below

The AWS SDK addresses buckets as `http://{bucket}.{your-endpoint}` *(virtual-host style)*, **except** when your endpoint is an IP address. So `http://127.0.0.1:9000` works out of the box, but `http://minio:9000` makes the relay proxy call `http://evaluation-data-bucket.minio:9000` and fail.

The relay proxy has no setting to force path-style addressing today. If you cannot use an IP address, set `MINIO_DOMAIN` on your MinIO server and make `{bucket}.{host}` resolve to it — a full Docker Compose example is on the [MinIO retriever page](https://gofeatureflag.org/docs/integrations/store-flags-configuration/minio.md#using-a-hostname-endpoint).

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

The GO Module exposes `S3ClientOptions`, which lets you force path-style addressing. This works with any endpoint, hostname or IP, and requires no MinIO-side configuration:

example.go

```
awsConfig, _ := config.LoadDefaultConfig(context.Background(),

    config.WithBaseEndpoint("http://minio:9000"))



config := ffclient.Config{

    // ...

    DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &s3exporterv2.Exporter{

            Format:    "json",

            Bucket:    "evaluation-data-bucket",

            S3Path:    "/go-feature-flag/variations/",

            Filename:  "flag-variation-{{ .Timestamp}}.{{ .Format}}",

            AwsConfig: &awsConfig,

            S3ClientOptions: []func(*s3.Options){

                func(o *s3.Options) { o.UsePathStyle = true },

            },

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field                 | Mandatory | Description                                                                                                                                                                                                                                |
| --------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`Bucket`**          | **        | The name of your bucket.                                                                                                                                                                                                                   |
| **`AwsConfig`**       | **        | An instance of `aws.Config`. Set your MinIO endpoint with `config.WithBaseEndpoint()`. If you omit it, `config.LoadDefaultConfig()` is called for you, which reads `AWS_ENDPOINT_URL_S3` and the other standard AWS environment variables. |
| **`S3ClientOptions`** | **        | Functional options passed to the S3 client. Set `o.UsePathStyle = true` to address buckets as `{endpoint}/{bucket}`, which is what MinIO expects by default.                                                                               |

info

`AwsConfig` is optional. If you leave it out, GO Feature Flag calls `config.LoadDefaultConfig()` itself, so setting `AWS_ENDPOINT_URL_S3`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_DEFAULT_REGION` in your environment is enough — you only need `S3ClientOptions` to force path-style addressing.

The remaining fields are shared with the [AWS S3 exporter](https://gofeatureflag.org/docs/integrations/export-evaluation-data/aws-s3.md#configure-the-go-module).

## Troubleshooting[​](#troubleshooting "Direct link to Troubleshooting")

#### `NoSuchBucket: The specified bucket does not exist` — but the bucket does exist[​](#nosuchbucket-the-specified-bucket-does-not-exist--but-the-bucket-does-exist "Direct link to nosuchbucket-the-specified-bucket-does-not-exist--but-the-bucket-does-exist")

Your endpoint is a hostname, so the SDK is using virtual-host addressing and MinIO reads the first path segment as the bucket name. Use an IP endpoint, configure [`MINIO_DOMAIN` and a DNS alias](https://gofeatureflag.org/docs/integrations/store-flags-configuration/minio.md#using-a-hostname-endpoint), or set `UsePathStyle` if you are on the GO Module.

#### `resolve auth scheme: resolve endpoint: endpoint rule error, Invalid region`[​](#resolve-auth-scheme-resolve-endpoint-endpoint-rule-error-invalid-region "Direct link to resolve-auth-scheme-resolve-endpoint-endpoint-rule-error-invalid-region")

The AWS SDK requires a region even when talking to MinIO. Any value works:

```
export AWS_DEFAULT_REGION=us-east-1
```

#### `WARN Response has no supported checksum. Not validating response payload.`[​](#warn-response-has-no-supported-checksum-not-validating-response-payload "Direct link to warn-response-has-no-supported-checksum-not-validating-response-payload")

Harmless, but if it floods your logs you can silence it with:

```
export AWS_RESPONSE_CHECKSUM_VALIDATION=when_required
```
