# Log

## Overview[​](#overview "Direct link to Overview")

<!-- -->

Export evaluation data inside the application logger.

<!-- -->

## Configure the relay proxy[​](#configure-the-relay-proxy "Direct link to Configure the relay proxy")

info

This exporter is set up by default in the relay-proxy.

## Configure the GO Module[​](#configure-the-go-module "Direct link to Configure the GO Module")

To configure your GO module to use the <!-- -->Log<!-- --> exporter, you need to add the following configuration to your `ffclient.Config{}` object:

example.go

```
config := ffclient.Config{

    // ...

   DataExporter: ffclient.DataExporter{

        Exporter: &logsexporter.Exporter{

            LogFormat: "[{{ .FormattedDate}}] user=\"{{ .UserKey}}\", flag=\"{{ .Key}}\", value=\"{{ .Value}}\"",

        },

    },

    // ...

}

err := ffclient.Init(config)

defer ffclient.Close()
```

| Field name  | Mandatory | Default                                                                             | Description                                                                                                                                                                                                                                                      |
| ----------- | --------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `LogFormat` | **        | `[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"` | LogFormat is the [template](https://golang.org/pkg/text/template/) configuration of the output format of your log.You can use all the key from the `exporter.FeatureEvent` + a key called `FormattedDate` that represents the date with the **RFC 3339** Format. |
