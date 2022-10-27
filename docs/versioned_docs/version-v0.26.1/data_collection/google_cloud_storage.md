# Google Cloud Storage Exporter

The **Google Cloud Storage exporter** will collect the data and create a new file in a specific folder everytime we send the data.

Everytime the `FlushInterval` or `MaxEventInMemory` is reached a new file will be added to S3.

:::info
If for some reason the Google Cloud Storage upload failed, we will keep the data in memory and retry to add the next time we reach `FlushInterval` or `MaxEventInMemory`.
:::

Check this [complete example](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/data_export_googlecloudstorage) to see how to export the data in S3.

## Configuration example
```go linenums="1"
ffclient.Config{ 
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
```

## Configuration fields
| Field         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
|---------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `Bucket `     | Name of your Google Cloud Storage Bucket.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| `CsvTemplate` | *(optional)* CsvTemplate is used if your output format is CSV. This field will be ignored if you are using another format than CSV. You can decide which fields you want in your CSV line with a go-template syntax, please check [internal/exporter/feature_event.go](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available.<br/>**Default:** `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}}\n` |
| `Filename`    | *(optional)* Filename is the name of your output file. You can use a templated config to define the name of your exported files.<br/>Available replacement are `{{ .Hostname}}`, `{{ .Timestamp}`} and `{{ .Format}}`<br/>Default: `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                                                                                                                                                                                                      |
| `Format`      | *(optional)* Format is the output format you want in your exported file. Available format are **`JSON`** and **`CSV`**. *(Default: `JSON`)*                                                                                                                                                                                                                                                                                                                                                                                                        |
| `Options`     | *(optional)* An instance of `option.ClientOption` that configures your access to Google Cloud. <br/> Check [this documentation for more info](https://cloud.google.com/docs/authentication).                                                                                                                                                                                                                                                                                                                                                        |
| `Path `       | *(optional)* The location of the directory in your bucket.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/gcstorageexporter).
