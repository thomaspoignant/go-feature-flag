# MinIO

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Retrieves the configuration from MinIO, or any other S3-compatible object store such as Ceph or Garage. There is no dedicated MinIO retriever: you use the S3 retriever and point it at your own server.

<!-- -->

[MinIO](https://min.io) is an S3-compatible object store you can run yourself, on-premises or in your own cloud. Everything below uses the [AWS S3 retriever](https://gofeatureflag.org/docs/integrations/store-flags-configuration/aws-s3.md) with `kind: s3` — only the endpoint changes.

The same applies the other way around, to [export your evaluation data to MinIO](https://gofeatureflag.org/docs/integrations/export-evaluation-data/minio.md).

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

The retriever configuration is the regular `s3` one, MinIO is selected through the AWS SDK environment variables:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: s3

    bucket: my-featureflag-bucket

    item: flag/flags.goff.yaml

# ...
```

```
export AWS_ENDPOINT_URL_S3=http://127.0.0.1:9000

export AWS_ACCESS_KEY_ID=minioadmin

export AWS_SECRET_ACCESS_KEY=minioadmin

export AWS_DEFAULT_REGION=us-east-1
```

| Field name                       | Mandatory | Type                 | Default  | Description                                                                                                                                                                                                                                                                                                               |
| -------------------------------- | --------- | -------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`                           | **        | string               | **none** | **Value should be `s3`**.<br />*There is no `minio` kind, MinIO is served by the S3 retriever.*                                                                                                                                                                                                                           |
| `bucket`                         | **        | string               | **none** | This is the name of your MinIO bucket *(ex: `my-featureflag-bucket`)*.                                                                                                                                                                                                                                                    |
| `item`                           | **        | string               | **none** | Path to the file inside the bucket *(ex: `config/flag/my-flags.yaml`)*.<br />*Leave it out to load a folder, see the [AWS S3 page](https://gofeatureflag.org/docs/integrations/store-flags-configuration/aws-s3.md#load-several-files-from-a-folder).*                                                                    |
| `prefix`, `pattern`, `recursive` | **        | string, string, bool | **none** | Folder mode filters, only usable without `item`. `pattern` is matched against the file name only *(no `/`)*, and files with an unsupported extension are skipped with a warning. See the [AWS S3 page](https://gofeatureflag.org/docs/integrations/store-flags-configuration/aws-s3.md#load-several-files-from-a-folder). |

Use an IP address in `AWS_ENDPOINT_URL_S3`, or read below

The AWS SDK addresses buckets as `http://{bucket}.{your-endpoint}` *(virtual-host style)*, **except** when your endpoint is an IP address. So `http://127.0.0.1:9000` works out of the box, but `http://minio:9000` makes the relay proxy call `http://my-featureflag-bucket.minio:9000` and fail.

The relay proxy has no setting to force path-style addressing today, so with a hostname endpoint you must make the virtual-host name resolve — see [Using a hostname endpoint](#using-a-hostname-endpoint).

### Using a hostname endpoint[​](#using-a-hostname-endpoint "Direct link to Using a hostname endpoint")

If you cannot use an IP address *(Docker Compose, Kubernetes, …)*, tell MinIO which domain it is served on with `MINIO_DOMAIN` **and** make `{bucket}.{host}` resolve to MinIO.

In Docker Compose this means one environment variable and one network alias per bucket:

compose.yaml

```
services:

  minio:

    image: minio/minio

    command: server /data --console-address ":9001"

    environment:

      MINIO_ROOT_USER: minioadmin

      MINIO_ROOT_PASSWORD: minioadmin

      MINIO_DOMAIN: minio

    networks:

      default:

        aliases:

          - my-featureflag-bucket.minio

    ports:

      - "9000:9000"

      - "9001:9001"



  relay-proxy:

    image: gofeatureflag/go-feature-flag:latest

    depends_on:

      - minio

    volumes:

      - ./goff-proxy.yaml:/goff/goff-proxy.yaml

    environment:

      AWS_ENDPOINT_URL_S3: http://minio:9000

      AWS_ACCESS_KEY_ID: minioadmin

      AWS_SECRET_ACCESS_KEY: minioadmin

      AWS_DEFAULT_REGION: us-east-1

    ports:

      - "1031:1031"
```

info

The network alias is required because Docker's DNS does not resolve wildcards: without `my-featureflag-bucket.minio` the relay proxy cannot reach the bucket even with `MINIO_DOMAIN` set.

If you use the **GO Module** instead of the relay proxy, you don't need any of this: use `S3ClientOptions` to force path-style addressing *(see below)*.

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

The GO Module exposes `S3ClientOptions`, which lets you force path-style addressing. This works with any endpoint, hostname or IP, and requires no MinIO-side configuration:

example.go

```
awsConfig, _ := config.LoadDefaultConfig(context.Background(),

    config.WithBaseEndpoint("http://minio:9000"))



err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &s3retrieverv2.Retriever{

        Bucket:    "my-featureflag-bucket",

        Item:      "flag/flags.goff.yaml",

        AwsConfig: &awsConfig,

        S3ClientOptions: []func(*s3.Options){

            func(o *s3.Options) { o.UsePathStyle = true },

        },

    },

})

defer ffclient.Close()
```

| Field                 | Mandatory | Description                                                                                                                                                                                                                                |
| --------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`Bucket`**          | **        | The name of your bucket.                                                                                                                                                                                                                   |
| **`Item`**            | **        | The location of your file in the bucket.                                                                                                                                                                                                   |
| **`AwsConfig`**       | **        | An instance of `aws.Config`. Set your MinIO endpoint with `config.WithBaseEndpoint()`. If you omit it, `config.LoadDefaultConfig()` is called for you, which reads `AWS_ENDPOINT_URL_S3` and the other standard AWS environment variables. |
| **`S3ClientOptions`** | **        | Functional options passed to the S3 client. Set `o.UsePathStyle = true` to address buckets as `{endpoint}/{bucket}`, which is what MinIO expects by default.                                                                               |

info

`AwsConfig` is optional. If you leave it out, GO Feature Flag calls `config.LoadDefaultConfig()` itself, so setting `AWS_ENDPOINT_URL_S3`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_DEFAULT_REGION` in your environment is enough — you only need `S3ClientOptions` to force path-style addressing:

example.go

```
Retriever: &s3retrieverv2.Retriever{

    Bucket: "my-featureflag-bucket",

    Item:   "flag/flags.goff.yaml",

    S3ClientOptions: []func(*s3.Options){

        func(o *s3.Options) { o.UsePathStyle = true },

    },

},
```

## Troubleshooting[​](#troubleshooting "Direct link to Troubleshooting")

#### `NoSuchBucket: The specified bucket does not exist` — but the bucket does exist[​](#nosuchbucket-the-specified-bucket-does-not-exist--but-the-bucket-does-exist "Direct link to nosuchbucket-the-specified-bucket-does-not-exist--but-the-bucket-does-exist")

Your endpoint is a hostname, so the SDK is using virtual-host addressing and MinIO reads the first path segment as the bucket name. Use an IP endpoint, configure [`MINIO_DOMAIN` and a DNS alias](#using-a-hostname-endpoint), or set `UsePathStyle` if you are on the GO Module.

#### `dial tcp ...: no route to host` on `http://my-featureflag-bucket.minio:9000`[​](#dial-tcp--no-route-to-host-on-httpmy-featureflag-bucketminio9000 "Direct link to dial-tcp--no-route-to-host-on-httpmy-featureflag-bucketminio9000")

Same cause, but the virtual-host name doesn't resolve at all. The bucket name in the URL is the tell.

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
