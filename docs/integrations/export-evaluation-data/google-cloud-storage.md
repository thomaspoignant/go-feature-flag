# Google Cloud Storage

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data to a Google Cloud Storage Bucket.

<!-- -->

Everytime the `FlushInterval` or `MaxEventInMemory` is reached a new file will be added to <!-- -->Google Cloud Storage<!-- -->.

info

If for some reason the <!-- -->Google Cloud Storage<!-- --> upload fails, we will keep the data in memory and retry to add the next time we reach `FlushInterval` or `MaxEventInMemory`.

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

To configure your relay proxy to use the <!-- -->Google Cloud Storage<!-- --> exporter, you need to add the following configuration to your relay proxy configuration file:

goff-proxy.yaml

```
# ...

exporters:

  - kind: googleStorage

    bucket: test-goff

# ...
```

| Field name                | Mandatory | Type   | Default                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------- | --------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kind`                    | **        | string | **none**                                                                                                                           | **Value should be `googleStorage`**.<br />*This field is mandatory and describes which retriever you are using.*                                                                                                                                                                                                                                                                        |
| `bucket`                  | **        | string | **none**                                                                                                                           | Name of your Google Cloud Storage Bucket.                                                                                                                                                                                                                                                                                                                                               |
| `path`                    | **        | string | **bucket root level**                                                                                                              | The location of the directory in Google Cloud Storage.                                                                                                                                                                                                                                                                                                                                  |
| `flushInterval`           | **        | int    | `60000`                                                                                                                            | The interval in millisecond between 2 calls to the webhook *(if the `maxEventInMemory` is reached before the flushInterval we will call the exporter before)*.                                                                                                                                                                                                                          |
| `maxEventInMemory`        | **        | int    | `100000`                                                                                                                           | If we hit that limit we will call the exporter.                                                                                                                                                                                                                                                                                                                                         |
| `format`                  | **        | string | `JSON`                                                                                                                             | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                 |
| `filename`                | **        | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                       | You can use a config template to define the name of your exported files. Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                               |
| `csvTemplate`             | **        | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br />This field will be ignored if you are using format other than CSV.<br />You can decide which fields you want in your CSV line with a go-template syntax, please check [`exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/exporter/feature_event.go) to see what are the fields available. |
| `parquetCompressionCodec` | **        | string | `SNAPPY`                                                                                                                           | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)                                                                                                                                                                                                          |

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Google Cloud Storage<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
config := ffclient.Config{

    // ...

   DataExporter: ffclient.DataExporter{

        // ...

        Exporter: &gcstorageexporter.Exporter{

            Bucket:   "test-goff",

            Format:   "json",

            Path:     "yourPath",

            Filename: "flag-variation-{{ .Timestamp}}.{{ .Format}}",

            Options:  []option.ClientOption{}, // Your google cloud SDK options

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field name                | Mandatory | Default                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                             |
| ------------------------- | --------- | ---------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Bucket`                  | **        | none                                                                                                                               | Name of your Google Cloud Storage Bucket.                                                                                                                                                                                                                                                                                                                                               |
| `Options`                 | **        | none                                                                                                                               | An instance of `option.ClientOption` that configures your access to Google Cloud.<br />Check [this documentation for more info](https://cloud.google.com/docs/authentication).                                                                                                                                                                                                          |
| `FlushInterval`           | **        | `60000`                                                                                                                            | The interval in millisecond between 2 calls to the webhook *(if the `maxEventInMemory` is reached before the flushInterval we will call the exporter before)*.                                                                                                                                                                                                                          |
| `MaxEventInMemory`        | **        | `100000`                                                                                                                           | If we hit that limit we will call the exporter.                                                                                                                                                                                                                                                                                                                                         |
| `Format`                  | **        | `JSON`                                                                                                                             | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                 |
| `Filename`                | **        | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                       | You can use a config template to define the name of your exported files. Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                               |
| `CsvTemplate`             | **        | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br />This field will be ignored if you are using format other than CSV.<br />You can decide which fields you want in your CSV line with a go-template syntax, please check [`exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/exporter/feature_event.go) to see what are the fields available. |
| `Path`                    | **        | **bucket root level**                                                                                                              | The location of the directory in S3.                                                                                                                                                                                                                                                                                                                                                    |
| `ParquetCompressionCodec` | **        | `SNAPPY`                                                                                                                           | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)                                                                                                                                                                                                          |
