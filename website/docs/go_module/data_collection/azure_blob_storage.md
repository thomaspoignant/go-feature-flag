---
sidebar_position: 2
---

# Azure Blob Storage Exporter

The **Azure Blob Storage exporter** will collect the data and create a new file in a specific folder everytime we send the data.

Everytime the `FlushInterval` or `MaxEventInMemory` is reached, a new file will be added to Google Cloud Storage.

:::info
If for some reason the Azure Blob Storage upload failed, we will keep the data in memory and retry to add it the next time we reach `FlushInterval` or `MaxEventInMemory`.
:::

## Configuration example
```go showLineNumbers
ffclient.Config{
    // ...
   DataExporter: ffclient.DataExporter{
        // ...
        Exporter: &azureexporter.Exporter{
            Container:   "test-goff",
            Format:   "json",
            Path:     "yourPath",
            Filename: "flag-variation-{{ .Timestamp}}.{{ .Format}}",
            AccountName: "goff",
            AccountKey: "XXXX",
        },
    },
    // ...
}
```

## Configuration fields
| Field                     | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
|---------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Container `              | Name of your Azure Blob Storage Container.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `AccountName `            | This is the name of your Azure Blob Storage account.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| `AccountKey `             | *(optional)* This is the secret key of your Azure Blob Storage account.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| `CsvTemplate`             | *(optional)* CsvTemplate is used if your output format is CSV. This field will be ignored if you are using format other than CSV. You can decide which fields you want in your CSV line with a go-template syntax, please check [internal/exporter/feature_event.go](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available.<br/>**Default:** `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` |
| `Filename`                | *(optional)* Filename is the name of your output file. You can use a templated config to define the name of your exported files.<br/>Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}`} and `{{ .Format}}`<br/>Default: `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                                                                                                                                                                                                               |
| `Format`                  | *(optional)* Format is the output format you want in your exported file. Available formats are **`JSON`**, **`CSV`**, **`Parquet`**. *(Default: `JSON`)*                                                                                                                                                                                                                                                                                                                                                                                                       |
| `Options`                 | *(optional)* An instance of `option.ClientOption` that configures your access to Google Cloud. <br/> Check [this documentation for more info](https://cloud.google.com/docs/authentication).                                                                                                                                                                                                                                                                                                                                                                   |
| `Path `                   | *(optional)* The location of the directory in your container.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `ParquetCompressionCodec` | *(optional)* ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md) *(Default: `SNAPPY`)*                                                                                                                                                                                                                                                                                                                                              |`

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/azureexporter).
