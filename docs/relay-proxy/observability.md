# 👀 Observability

## Overview[​](#overview "Direct link to Overview")

Observability is crucial to know how your applications are working. And this is probably something you want to have for your feature flag solution as well. This is why GO Feature Flag relay-proxy offers out of the box some ways of monitor what is happening.

## Logging[​](#logging "Direct link to Logging")

GO Feature Flag relay proxy is using [Zap](https://pkg.go.dev/go.uber.org/zap) as the logging library.

By default, the relay proxy is logging to the standard output with a JSON format in level `INFO`.

You can change the log level by setting the `logLevel` in your configuration file and you can choose the log format by setting the `logFormat` options.

goff-proxy.yaml

```
#...

logLevel: ERROR

logFormat: json

#...
```

| Field name | Mandatory | Type   | Default    | Description                                                                                                                                                                                                                       |
| ---------- | --------- | ------ | ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| logLevel   | **        | string | **`INFO`** | The log level to use for the relay proxy.<br />Available values are `ERROR`, `WARN`, `INFO`, `DEBUG`.                                                                                                                             |
| logFormat  | **        | string | **`json`** | The format to use for structured logs from the relay proxy.<br />Valid values are `json` and `logfmt`.- **json**: Will output the logs in `JSON` format (default).<br />- **logfmt**: Will output the log as text in the console. |

## Tracing[​](#tracing "Direct link to Tracing")

The relay proxy will attempt to send traces to an OpenTelemetry collector or compatible agent if an OpenTelemetry exporter endpoint is configured.

To set the endpoint, set the `OTEL_EXPORTER_OTLP_ENDPOINT` environment variable or set `otel.exporter.otlp.endpoint` in the configuration file.

* Environment variable
* Configuration file

```
export OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

goff-proxy.yaml

```
# ...

otel:

  exporter:

    otlp:

      endpoint: "http://localhost:4317"
```

See [the OpenTelemetry documentation](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) for more information.

All your requests will be traced and sent to the collector with the service name **`go-feature-flag`**.

note

If you want to try the OpenTelemetry integration locally, follow this [example](https://github.com/thomaspoignant/go-feature-flag/tree/main/cmd/relayproxy/testdata/opentelemetry) to set up Jaeger and see your traces.

***

## Monitoring[​](#monitoring "Direct link to Monitoring")

The **relay proxy** offers some endpoints for you to be able to see how it behaves.

### `/health`[​](#health "Direct link to health")

Making a **GET** request to the URL path `/health` will tell you if the relay proxy is ready to serve traffic.

This is useful especially for loadbalancer to know that they can send traffic to the service.

### `/info`[​](#info "Direct link to info")

Making a **GET** request to the URL path `/info` will give you information about the actual state of the relay proxy.

### `/metrics`[​](#metrics "Direct link to metrics")

This endpoint is providing metrics about the relay proxy in the prometheus format.

note

The relay proxy supports [Prometheus native histograms](https://prometheus.io/docs/specs/native_histograms/) for improved latency distribution precision. When scraping with Prometheus (especially using protobuf format), histogram metrics will include both classic bucket data and native histogram data for higher-fidelity percentile calculations. Classic histogram buckets are preserved for backward compatibility.

***

## Use specific port for the monitoring[​](#use-specific-port-for-the-monitoring "Direct link to Use specific port for the monitoring")

You can configure a different port for the monitoring endpoints.<br /><!-- -->This is useful if you want to expose the monitoring endpoints on a different port than the main service.

goff-proxy.yaml

```
# ...

monitoringPort: 1032 # only the monitoring endpoints will be exposed on this port

# ...
```

note

By default the monitoring endpoints are exposed on the same port as the main service.
