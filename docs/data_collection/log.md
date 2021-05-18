# Log Exporter
The log exporter is here mostly for backward compatibility *(originally every variation were logged, but it can be a lot of data for a default configuration)*.  
It will use your logger `ffclient.Config.Logger` to log every variation changes.

You can configure your output log with the `Format` field.  
It uses a [go template](https://golang.org/pkg/text/template/) format.

## Configuration example
```go linenums="1"
ffclient.Config{
    // ...
   DataExporter: ffclient.DataExporter{
        Exporter: &ffexporter.Log{
            Format: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",
        },
    },
    // ...
}
```

## Configuration fields
| Field  | Description  |
|---|---|
|`Format`   | *(optional)*<br>Format is the [template](https://golang.org/pkg/text/template/) configuration of the output format of your log.<br>You can use all the key from the `exporter.FeatureEvent` + a key called `FormattedDate` that represent the date with the **RFC 3339** Format.<br><br>**Default: `[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"`**  |

Check the [godoc for full details](https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag@v0.11.0/ffexporter#Log).
