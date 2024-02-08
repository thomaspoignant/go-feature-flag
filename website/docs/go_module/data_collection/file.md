---
sidebar_position: 3
---

# File Exporter
The file exporter will collect the data and create a new file in a specific folder everytime we send the data.  
This file should be in the local instance.

Check this [complete example](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/data_export_file) to see how to export the data in a file.

## Configuration example
```go showLineNumbers
ffclient.Config{ 
    // ...
   DataExporter: ffclient.DataExporter{
        // ...
        Exporter: &fileexporter.Exporter{
            OutputDir: "/output-data/",
            Format: "csv",
            FileName: "flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}",
            CsvTemplate: "{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n"
        },
    },
    // ...
}
```

## Configuration fields

| Field  | Description  |
|---|---|
|`OutputDir`   | OutputDir is the location of the directory to store the exported files.<br/>It should finish with a `/`.  |
|`Format`   |   _(Optional)_ Format is the output format you want in your exported file.<br/>Available format: **`JSON`**, **`CSV`**, **`Parquet`**.<br/>**Default: `JSON`** |
|`Filename`   | _(Optional)_ Filename is the name of your output file.<br/>You can use a templated config to define the name of your exported files.<br/>Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}}`<br/>**Default: `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`**|
|`CsvTemplate`   | _(Optional)_ CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using format other than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [internal/exporter/feature_event.go](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see the available fields.<br/>**Default:** `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}}\n` |
| `ParquetCompressionCodec` | _(Optional)_ ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)<br/>**Default: `SNAPPY`** |`

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag/exporter/fileexporter).
