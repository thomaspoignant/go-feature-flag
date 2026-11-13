# AWS S3

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Retrieves the configuration from an AWS S3 bucket. This retriever is useful when you are using AWS and want to use S3 to store you configuration files. You can point it to a single file (using "item"), or leave "item" out to load a folder of your bucket and merge several flag files into one configuration.

<!-- -->

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->AWS S3<!-- --> retriever, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

retrievers:

  - kind: s3

    bucket: my-featureflag-bucket

    item: flag/flags.goff.yaml

# ...
```

info

If you are using the S3 provider, the easiest way to provide credentials is to set environment variables. It will be used by GO Feature Flag to identify to your S3 bucket.

```
export AWS_SECRET_ACCESS_KEY=xxxx

export AWS_ACCESS_KEY_ID=xxx

export AWS_DEFAULT_REGION=eu-west-1
```

| Field name  | Mandatory | Type   | Default  | Description                                                                                                                                                                                                       |
| ----------- | --------- | ------ | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`      | **        | string | **none** | **Value should be `s3`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                             |
| `bucket`    | **        | string | **none** | This is the name of your S3 bucket *(ex: `my-featureflag-bucket`)*.                                                                                                                                               |
| `item`      | **        | string | **none** | Path to the file inside the bucket *(ex: `config/flag/my-flags.yaml`)*.<br />*Set it to load one file; leave it out to load a folder, see [Load several files from a folder](#load-several-files-from-a-folder).* |
| `prefix`    | **        | string | **none** | *Folder mode only.* Key prefix *(ex: `flags/`)* restricting the files loaded. Empty or `/` is the root of the bucket.                                                                                             |
| `pattern`   | **        | string | **none** | *Folder mode only.* Optional glob matched against the file name only, not the whole path *(ex: `*.yaml`)*, so it can't contain a `/`. If empty, every file is used.                                               |
| `recursive` | **        | bool   | `false`  | *Folder mode only.* If `true`, files in nested "folders" are also loaded, otherwise only the direct children of the prefix.                                                                                       |

**Set `item` to load a single file, leave it out to load a folder.** `prefix`, `pattern` and `recursive` can only be used without `item`.

### Load several files from a folder[​](#load-several-files-from-a-folder "Direct link to Load several files from a folder")

When `item` is not set, the retriever lists the objects of the bucket *(restricted by `prefix`, `pattern` and `recursive`)*, downloads them and merges the flags into a single flag set. It is useful when several teams publish their own flag files to the same bucket without writing to one huge file.

goff-proxy.yaml

```
retrievers:

  - kind: s3

    bucket: my-featureflag-bucket

    prefix: flags/

    pattern: "*.yaml"

    recursive: true
```

warning

Without `item` **and** without `prefix`, the **root of the bucket** is read *(everything below it with `recursive: true`)*. Set a `pattern` when other files live there. Also beware that a typo in the `item` key silently switches the retriever to folder mode.

How it works:

* **Format**: it is derived from the file extension *(`.json`, `.yaml`/`.yml`, `.toml`)*. All the files must have the same format, otherwise the retrieval fails with an error naming the offending files. Use `pattern` to select only one format when your prefix contains several. A file with another extension is skipped with a warning in the logs *(once, as long as it stays there)*, it does not make the retrieval fail.
* **Merge**: the top-level flag keys of every file are merged. If the same flag key is defined in 2 files, the retrieval fails with an error naming the flag and both files.
* **Polling**: the retriever lists the objects at each poll and only downloads the files whose `ETag` changed (or that are new). An object listed without an `ETag` is downloaded at every poll. The other ones are served from memory.
* **Deletion**: if a file is removed from the bucket, its flags disappear from the flag set at the next poll. If no file matches, the flag set is empty *(this is not an error, but a warning is logged)*.
* Keys ending with `/` *(folder markers)* are ignored. Without `recursive`, only the direct children of the prefix are used.
* Without `prefix` *(or with `prefix: /`)* the root of the bucket is used: top-level files only, or the whole bucket with `recursive: true`. Leading `/` of a prefix are trimmed *(`/flags/` is `flags/`)* because S3 keys never start with a `/`.
* A trailing `/` is added to the prefix when missing, so `flags` and `flags/` are equivalent and `flags-old/...` is never matched.
* Empty files are accepted and contribute no flag, whatever their format.
* Files are parsed exactly like a single `item` file, so any file valid as an `item` is valid in folder mode.

IAM permissions

In folder mode the identity used by GO Feature Flag needs `s3:ListBucket` on the bucket *(`arn:aws:s3:::my-featureflag-bucket`)* in addition to `s3:GetObject` on the objects *(`arn:aws:s3:::my-featureflag-bucket/flags/*`)*.

### Compatibility with S3-Compatible Solutions[​](#compatibility-with-s3-compatible-solutions "Direct link to Compatibility with S3-Compatible Solutions")

The S3 retriever also works with S3-compatible object stores such as **MinIO**, Ceph or Garage: keep `kind: s3` and point the AWS SDK at your server with `AWS_ENDPOINT_URL_S3`.

info

Read the [**MinIO** page](https://gofeatureflag.org/docs/integrations/store-flags-configuration/minio.md) for the full setup, including the bucket addressing gotcha you will hit if your endpoint is a hostname rather than an IP address.

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->AWS S3<!-- --> retriever, you need to add the following configuration to your `ffclient.Config{}` object:

info

The [**S3 Retriever v2**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag@v1.23.1/retriever/s3retrieverv2) will use the [`aws-sdk-go-v2`](https://github.com/aws/aws-sdk-go-v2) to access your flag in an S3 bucket.

The [**S3 Retriever v1**](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag@v1.23.1/retriever/s3retriever) will use the deprecated [`aws-sdk-go`](https://github.com/aws/aws-sdk-go) to access your flag in an S3 bucket.

[AWS has announce end-of-support for AWS SDK for Go v1](https://aws.amazon.com/blogs/developer/announcing-end-of-support-for-aws-sdk-for-go-v1-on-july-31-2025/), and it's recommended to migrate from S3 Retriever v1 to v2.

example.go

```
awsConfig, _ := config.LoadDefaultConfig(context.Background())

err := ffclient.Init(ffclient.Config{

    PollingInterval: 3 * time.Second,

    Retriever: &s3retrieverv2.Retriever{

        Bucket: "my-featureflag-bucket",

        Item:   "flag/flags.goff.yaml",

        AwsConfig: &awsConfig,

    },

})

defer ffclient.Close()
```

To load several files from a folder *(leave `Item` empty)*:

example\_prefix.go

```
Retriever: &s3retrieverv2.Retriever{

    Bucket:    "my-featureflag-bucket",

    Prefix:    "flags/",

    Pattern:   "*.yaml",

    Recursive: true,

    AwsConfig: &awsConfig,

},
```

| Field           | Mandatory | Description                                                                                                                                                                                                                                                                                                                                 |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **`Bucket`**    | **        | The name of your bucket.                                                                                                                                                                                                                                                                                                                    |
| **`Item`**      | **        | The location of your file in the bucket. If not set, the retriever is in folder mode and merges several files (see [above](#load-several-files-from-a-folder)).                                                                                                                                                                             |
| **`Prefix`**    | **        | *Folder mode only.* Key prefix *(ex: `flags/`)*. Empty or `/` is the root of the bucket. Requires the `s3:ListBucket` permission.                                                                                                                                                                                                           |
| **`Pattern`**   | **        | *Folder mode only.* Glob matched against the file name only, it can't contain a `/` *(ex: `*.yaml`)*.                                                                                                                                                                                                                                       |
| **`Recursive`** | **        | *Folder mode only.* If `true`, files in nested "folders" are also loaded. Defaults to `false`.                                                                                                                                                                                                                                              |
| **`AwsConfig`** | **        | An instance of `aws.Config` that configure your access to AWS<br />*check [this documentation for more info](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html)*.<br />If not set, `config.LoadDefaultConfig()` is called for you, which reads the standard AWS environment variables and shared config files. |
