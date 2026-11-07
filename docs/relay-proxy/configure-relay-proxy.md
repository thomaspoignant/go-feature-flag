# 🔧 Configuration

## Overview[​](#overview "Direct link to Overview")

In this page, you will find all the configuration options available for the **relay-proxy**.

The **GO Feature Flag relay-proxy** is highly configurable and can be adapted to your needs. Despite that you can start with a minimum configuration and fine tune it later.

## Configuration file[​](#configuration-file "Direct link to Configuration file")

The configuration of the **relay proxy** is based on a configuration file that you have to provide. This is a `yaml` file and usage is to call it `goff-proxy.yaml` but you can name it as you will.

You can provide the configuration file with the option `--config=/path_to_your_file.yaml` when launching the relay proxy.

```
go-feature-flag --config=/goff-proxy.yaml
```

info

By default, if you omit the `--config` option, the relay proxy will look for a configuration file named `goff-proxy.yaml` in *(int this order)*:

1. `./` *(the current directory)*
2. `/goff/`
3. `/etc/opt/goff/`

## Configuration Patterns[​](#configuration-patterns "Direct link to Configuration Patterns")

[**](#getting-started)

### [One set of feature flags](#getting-started)

[Use one set of feature flags for all your applications. This is the default way of using GO Feature Flag, it requires a minimum configuration and is easy to setup](#getting-started)

[Get started**](#getting-started)

[**](#use-multiple-flag-sets)

### [Use multiple flag sets](#use-multiple-flag-sets)

[Separate your feature flags into different flag sets to give ownership to different teams or different usage, useful for multi-tenant applications or when you have different teams.](#use-multiple-flag-sets)

[Get started**](#use-multiple-flag-sets)

## Getting Started[​](#getting-started "Direct link to Getting Started")

The minimum configuration to start the **relay proxy** is to provide where to retrieve your feature flags configuration.

goff-proxy.yaml

```
retrievers:

  - kind: file

    path: /goff/flags.yaml # Location of your feature flag files
```

All the other options are optional and can be added to the configuration file to fine-tune the **relay proxy**.

## Environment variables[​](#environment-variables "Direct link to Environment variables")

The primary way to configure the relay proxy is through a configuration file, but you can also use environment variables if you prefer.<br /><!-- -->You can override file configurations using environment variables.

**Here are some things to know when you are using environment variables:**

* To set an option, the environment variable should have the same name as the configuration option in uppercase.<br /><!-- -->For example, to set the `logLevel` option, you can set the `LOGLEVEL` environment variable.

  ```
  export LOGLEVEL=debug
  ```

* If you want to set a nested field, you can use `_` to separate each field.<br /><!-- -->For example, to set the `server.port` option, you can set the `SERVER_PORT` environment variable.

  ```
  export SERVER_PORT=8080
  ```

* If you want to set an array of string, you can add multiple values separated by a comma.

  ```
  export AUTHORIZEDKEYS_EVALUATION=my-first-key,my-second-key
  ```

* If you are modifying an array of objects, you can use the index to set the value.

  ```
  export RETRIEVERS_0_KIND=github
  ```

info

If you want to avoid any collision with existing environment variables that are not specifically related to the relay proxy, you can prefix your environment variables.

You can configure the prefix using the [`envVariablePrefix`](#envvariableprefix) field in your configuration file.

## Configuration Options[​](#configuration-options "Direct link to Configuration Options")

### `server`[​](#server "Direct link to server")

This is the configuration related to the server behavior of the relay proxy. It allows the internal server behaviors such as your execution mode, monitoring port, etc ...

* option name: `server`
* type: [`serverConfig`](#type-serverconfig)
* default: **none**
* mandatory:**
* Check the [server configuration types](#type-serverconfig) to have more information about the available options.

### `retrievers`[​](#retrievers "Direct link to retrievers")

This is the configuration on how to retrieve the configuration of the files.

* option name: `retrievers`
* type: [`[]retriever`](#type-retriever)
* default: **none**
* **mandatory:****
* You can have multiple retrievers at the same time, check [*"Using multiple retrievers"*](https://gofeatureflag.org/docs/concepts/retriever.md#using-multiple-retrievers) to see how it works.
* Configuration of a retriever depends on the one you are choosing. Check the [list of retrievers](https://gofeatureflag.org/docs/integrations/store-flags-configuration.md#available-retrievers) and go on the detail page of your choice to configure it.

example goff-proxy.yaml

```
retrievers:

  - kind: file

    path: /goff/my-flags.yaml

  - kind: s3

    bucket: my-featureflag-bucket

    item: flag/flags.goff.yaml
```

### `startWithRetrieverError`[​](#startwithretrievererror "Direct link to startwithretrievererror")

By default, the **relay proxy** will crash if it is not able to retrieve the flags from the configuration.

If you don't want your relay proxy to crash, you can set `startWithRetrieverError` to true. Until the flag is retrievable the relay proxy will only answer with default values.

* option name: `startWithRetrieverError`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `persistentFlagConfigurationFile`[​](#persistentflagconfigurationfile "Direct link to persistentflagconfigurationfile")

If set, GO Feature Flag will store the flags configuration in this file to be able to serve the flags even if none of the retrievers is available during starting time.

info

By default, the flag configuration is not persisted and stays on the retriever system. By setting a file here, you ensure that GO Feature Flag will always start with a configuration but which can be out-dated.

* option name: `persistentFlagConfigurationFile`
* type: **string**
* default: **none**
* mandatory:**
* example: `/tmp/goff_persist_conf.yaml`

### `fileFormat`[​](#fileformat "Direct link to fileformat")

This is the format of your flag configuration file.

* option name: `fileFormat`
* type: **string**
* Acceptable values are :`yaml`, `json`, `toml`.
* default: **`yaml`**
* mandatory:**

### `pollingInterval`[​](#pollinginterval "Direct link to pollinginterval")

This is the time interval **in millisecond** when the relay proxy is reloading the configuration file.

* option name: `pollingInterval`
* type: **int**
* default: **`60000`**
* The minimum time accepted is 1000 millisecond (1 second).
* If you set a negative value, the polling will be disabled.
* mandatory:**

### `enablePollingJitter`[​](#enablepollingjitter "Direct link to enablepollingjitter")

Set to true if you want to avoid having true periodicity when retrieving your flags. It is useful to avoid having spike on your flag configuration storage in case your application is starting multiple instance at the same time.

info

We ensure a deviation that is maximum ±10% of your polling interval.

* option name: `enablePollingJitter`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `logLevel`[​](#loglevel "Direct link to loglevel")

The log level to use for the relay proxy.

Check [*"Logging"*](https://gofeatureflag.org/docs/relay-proxy/observability.md#logging) to have more information.

* option name: `logLevel`
* type: **string**
* default: **`info`**
* Acceptable values are `error`, `warn`, `info`, `debug`.
* mandatory:**
* When in `DEBUG` mode, the relay proxy will log the request body and response body, for large payloads this can be a lot of data and can impact the performance this is the reason why we truncate the body if it is larger than **8KiB**.

### `logFormat`[​](#logformat "Direct link to logformat")

The format to use for structured logs from the relay proxy.

Check [*"Logging"*](https://gofeatureflag.org/docs/relay-proxy/observability.md#logging) to have more information.

* option name: `logFormat`
* type: **string**
* default: **`json`**
* Acceptable values are `json` and `logfmt`.
* mandatory:**

### `enablePprof`[​](#enablepprof "Direct link to enablepprof")

This param is used to enable **pprof endpoints** if you wish to enable profiling.

Check [*"Profiling"*](https://gofeatureflag.org/docs/relay-proxy/profiling.md) to have more information.

* option name: `enablePprof`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `exporters`[​](#exporters "Direct link to exporters")

Exporters is the configuration used to define how to export the data from GO Feature Flag.

* option name: `exporters`
* type: **[\[\]exporter](#type-exporter)**
* default: **none**
* mandatory:**
* You can have multiple exporters at the same time, it will export the data to all the exporters you have configured in parallel.
* Configuration of an exporter depends on the one you are choosing. Check the [list of exporters](https://gofeatureflag.org/docs/integrations/export-evaluation-data.md#available-exporters) and go on the detail page of your choice to configure it.

example goff-proxy.yaml

```
exporters:

  - kind: s3

    bucket: evaluation-data-bucket

  - kind: kafka

    kafka:

      topic: "go-feature-flag-events"
```

### `notifiers`[​](#notifiers "Direct link to notifiers")

Notifiers is the configuration on where to notify a flag changes.

* option name: `notifiers`
* type: **[\[\]notifier](#type-notifier)**
* default: **none**
* mandatory:**
* Configuration of a notifier depends on the one you are choosing. Check the [list of notifiers](https://gofeatureflag.org/docs/integrations/notify-flags-changes.md#available-notifiers) and go on the detail page of your choice to configure it.

### `disableNotifierOnInit`[​](#disablenotifieroninit "Direct link to disablenotifieroninit")

If `disableNotifierOnInit` is set to **true**, the relay proxy will not call any notifier when the flags are loaded for the 1st time *(during initialization)*. This is useful if you do not want a Slack/Webhook notification saying that the flags have been added every time you start the proxy.

* option name: `disableNotifierOnInit`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `debug`[​](#debug "Direct link to debug")

If `debug` is set to true, we will set the log level to debug and set some components in debug mode (`labstack/echo`, `pprof` etc ...).

* option name: `debug`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `hideBanner`[​](#hidebanner "Direct link to hidebanner")

`hideBanner` can be used to hide the startup banner.

* option name: `hideBanner`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `swagger`[​](#swagger "Direct link to swagger")

Configuration related to the swagger behavior of the relay proxy.

* option name: `swagger`
* type: **[swagger](#type-swagger)**
* default: **none**
* mandatory:**

### `ofrepEventStream`[​](#ofrepeventstream "Direct link to ofrepeventstream")

Advertises the flag-change SSE endpoint in the OFREP bulk evaluation response so that OFREP providers can subscribe to changes instead of polling (see [OpenFeature ADR-0008](https://github.com/open-feature/protocol/blob/main/service/adrs/0008-sse-for-bulk-evaluation-changes.md)).

It is enabled when [`ofrepEventStream.baseUrl`](#ofrepeventstreambaseurl) is set.

* option name: `ofrepEventStream`
* type: **[ofrepEventStream](#type-ofrepeventstream)**
* default: **none**
* mandatory:**

### `disableVersionHeader`[​](#disableversionheader "Direct link to disableversionheader")

If `disableVersionHeader` is set to **`true`**, the relay proxy will not add the header `x-gofeatureflag-version` with the GO Feature Flag version in the HTTP response.

* option name: `disableVersionHeader`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `disableFlagDetailsInStream`[​](#disableflagdetailsinstream "Direct link to disableflagdetailsinstream")

If `disableFlagDetailsInStream` is set to **`true`**, WebSocket flag change payloads omit flag definitions, including targeting rules. The `deleted`, `added`, and `updated` objects retain the changed flag names with empty object values so clients can refresh their evaluated flags without receiving the underlying configuration. Enabling this setting requires restarting the relay proxy.

* option name: `disableFlagDetailsInStream`
* type: **boolean**
* default: **`false`**
* mandatory:**

### `authorizedKeys`[​](#authorizedkeys "Direct link to authorizedkeys")

List of authorized API keys.

info

If you don't set any API keys, the relay proxy will be fully open to all requests

* option name: `authorizedKeys`
* type: **[authorizedKeys](#type-authorizedkeys)**
* default: **none**
* mandatory:**

tip

**Runtime updates**: API keys can be updated at runtime without restarting the relay proxy.<br /><!-- -->When the configuration file changes, the relay proxy automatically detects and applies the new API key configuration.<br /><!-- -->This works in both default mode and flagset mode.

### `evaluationContextEnrichment`[​](#evaluationcontextenrichment "Direct link to evaluationcontextenrichment")

It is a free field that will be merged with the evaluation context sent during the evaluation.

It is useful to add common attributes to all the evaluations, such as a server version, environment, etc. These fields will be included in the custom attributes of the evaluation context.

warning

If in the evaluation context you have a field with the same name, it will be overridden by the `evaluationContextEnrichment`.

* option name: `evaluationContextEnrichment`
* type: **object**
* default: **none**
* mandatory:**

### `openTelemetryOtlpEndpoint`[​](#opentelemetryotlpendpoint "Direct link to opentelemetryotlpendpoint")

Endpoint of your **OpenTelemetry OTLP collector**, used to send traces.

You will be able to forward those traces to your OpenTelemetry solution directly from the collector.

* option name: `openTelemetryOtlpEndpoint`
* type: **string**
* default: **none**
* mandatory:**

### `flagSets`[​](#flagsets "Direct link to flagsets")

Flag sets allow you to organize your feature flags into separate groups, each with its own configuration, API keys, retrievers, exporters, and notifiers.

When you configure flag sets, the relay proxy operates in **flag set mode** instead of the default single configuration mode. Each flag set requires at least one API key and can have its own retrievers, exporters, and notifiers.

* option name: `flagSets`
* type: **[\[\]flagSet](#type-flagset)**
* default: **none**
* mandatory:**
* If flag sets are configured, the relay proxy will ignore the top-level `retrievers`, `exporters`, and `notifiers` configuration.

example goff-proxy.yaml

```
server:

  mode: http

  port: 1031

flagSets:

  - name: teamA

    apiKeys:

      - "team-a-api-key"

    retrievers:

      - kind: file

        path: /flags/team-a-flags.yaml

    exporters:

      - kind: log



  - name: teamB

    apiKeys:

      - "team-b-api-key"

    retrievers:

      - kind: s3

        bucket: my-feature-flag-bucket

        item: team-b/flags.yaml

    exporters:

      - kind: s3

        bucket: evaluation-data-bucket
```

**Runtime updates**: flag sets can be **added** or **removed** at runtime without restarting the relay proxy.<br /><!-- -->When the configuration file changes, the relay proxy detects it and reconciles the running flag sets: a newly added flag set is started immediately and a removed flag set is stopped.

warning

* **Modifying** an existing flag set (its `retrievers`, `exporters`, `notifiers`, `pollingInterval`, …) is **not supported at runtime**. The only exception is its `apiKeys`, which can be changed live. Any other modification is **ignored** and an error is logged; the change will only be applied the next time the relay proxy is **restarted**.
* Runtime add/remove only works for flag sets that have an explicit, unique [`name`](#flagsetname). Flag sets without a name (or named `default`) get a generated name at startup and cannot be added or removed at runtime.

### `envVariablePrefix`[​](#envvariableprefix "Direct link to envvariableprefix")

Prefix for your environment variables.<br /><!-- -->This useful if you want to avoid any collision with existing environment variables that are not specifically related to the relay proxy.

If you use this option you will have to prefix all your environment variables with the value you set here. For example if you want to override the port of the relay-proxy and the prefix is `GOFF_`.

```
export GOFF_SERVER_PORT=8080
```

* option name: `envVariablePrefix`
* type: **string**
* default: **empty string**
* mandatory:**

### `enableBulkMetricFlagNames`[​](#enablebulkmetricflagnames "Direct link to enablebulkmetricflagnames")

Enables per-flag metrics for bulk evaluation endpoints. This adds `flag_name` labels to the `all_flags_evaluations_total_with_flag` metric.

info

When enabled, this option provides more granular metrics by including the flag name in the metric labels for bulk evaluation endpoints. This can be useful for monitoring and alerting on specific flags, but may increase the cardinality of your metrics.

* option name: `enableBulkMetricFlagNames`
* type: **boolean**
* default: **`false`**
* mandatory:**

## Configuration Types[​](#configuration-types "Direct link to Configuration Types")

### type `authorizedKeys`[​](#type-authorizedkeys "Direct link to type-authorizedkeys")

`authorizedKeys` is an object were it is possible to configure the list of API Keys that are accepted by the relay proxy.

#### `authorizedKeys.evaluation`[​](#authorizedkeysevaluation "Direct link to authorizedkeysevaluation")

If set, we will check for each evaluation if an authorized key is provided.

Each request will need to provide one of authorized key inside `X-API-Key` header.

* option name: `evaluation`
* type: **\[]string**
* default: **none**
* mandatory:**

#### `authorizedKeys.admin`[​](#authorizedkeysadmin "Direct link to authorizedkeysadmin")

You need to set API keys in this field if you want to access the `/v1/admin/*` endpoints.

Each request will need to provide one of authorized key inside `X-API-Key` header.

* option name: `admin`
* type: **\[]string**
* default: **none**
* mandatory:**
* If no api key is configured the endpoint will be unreachable.

#### Example[​](#example "Direct link to Example")

goff-proxy.yaml

```
# ...

authorizedKeys:

  evaluation:

    - "my-first-key"

    - "my-second-key"

  admin:

    - "my-first-admin-key"

    - "my-second-admin-key"
```

### type `serverConfig`[​](#type-serverconfig "Direct link to type-serverconfig")

Configuration related to the server behavior of the relay proxy.

#### `server.mode`[​](#servermode "Direct link to servermode")

The execution mode of the relay proxy, you can choose to start the relay-proxy in different modes depending how you want to call the APIs.

* option name: `mode`
* type: **string**
* Acceptable values are :`http`, `lambda`, `unixsocket`.
* default: **http**
* mandatory:**
* `http`: Start the relay proxy as a HTTP server.
  <br />
  `lambda`: Start the relay proxy as an AWS Lambda handler.
  <br />
  `unixsocket`: Start the relay proxy and listen on a Unix socket, you have to set the `unixSocketPath` field to define where the socket file will be created.

#### `server.unixSocketPath`[​](#serverunixsocketpath "Direct link to serverunixsocketpath")

The path where the Unix socket file will be created when the relay proxy is started in `unixsocket` mode. If used in combination with `server.monitoringPort` monitoring endpoints will be served via http on the provided port.

* option name: `unixSocketPath`
* type: **string**
* default: **none**
* mandatory:**
* This field is mandatory if the `mode` is set to `unixsocket`.

#### `server.host`[​](#serverhost "Direct link to serverhost")

The server host used by the relay proxy to bind the HTTP server.

* option name: `host`
* type: **string**
* default: **`0.0.0.0`**
* mandatory:**
* This field is used only in `http` mode.

#### `server.port`[​](#serverport "Direct link to serverport")

The server port used by the relay proxy to bind the HTTP server.

* option name: `port`
* type: **int**
* default: **1031**
* mandatory:**
* This field is used only in `http` mode.

#### `server.monitoringPort`[​](#servermonitoringport "Direct link to servermonitoringport")

If set the monitoring endpoints will be served on this specific port.

Check [*"Use specific port for the monitoring"*](https://gofeatureflag.org/docs/relay-proxy/observability.md#use-specific-port-for-the-monitoring) to have more information.

* option name: `monitoringPort`
* type: **int**
* default: **none**
* mandatory:**
* This field is used only in `http` mode.

#### `server.awsLambdaAdapter`[​](#serverawslambdaadapter "Direct link to serverawslambdaadapter")

This parameter allow you to decide which type of AWS lambda handler you want to use.

* option name: `awsLambdaAdapter`
* type: **string**
* default: **`APIGatewayV2`**
* Acceptable values are `APIGatewayV2`, `APIGatewayV1`, `ALB`.
* mandatory:**
* This param is used only if `server.mode` is set to `lambda`.

#### `server.awsApiGatewayBasePath`[​](#serverawsapigatewaybasepath "Direct link to serverawsapigatewaybasepath")

Specifies the base path prefix for AWS API Gateway deployments when using non-root routes. The relay proxy will strip this base path from incoming requests before processing them. This is useful when deploying behind paths like `/api` or `/dev/feature-flags`.

* option name: `awsApiGatewayBasePath`
* type: **string**
* default: **none**
* mandatory:**
* This param is used only if `server.mode` is set to `lambda`.

**Example:** If set to `/api/feature-flags`, requests to `/api/feature-flags/health` will be processed as `/health`.

### type `retriever`[​](#type-retriever "Direct link to type-retriever")

A [retriever](https://gofeatureflag.org/docs/concepts/retriever.md) is the component in charge of loading your flag configuration from a remote source.

#### `retriever.kind`[​](#retrieverkind "Direct link to retrieverkind")

Configuring a retriever always start by setting the kind of retriever you want to use. This is the only common param between all the retrievers.

* option name: `kind`
* type: **string**
* default: **none**
* mandatory:**

info

To find the **`kind`** to use for your retriever and the **rest of the options**, check the [list of retrievers](https://gofeatureflag.org/docs/integrations/store-flags-configuration.md#available-retrievers) and go on the detail page of your choice to configure it.

**Available retrievers**:

* [HTTP(S)](https://gofeatureflag.org/../integrations/store-flags-configuration/http)
* [File System](https://gofeatureflag.org/../integrations/store-flags-configuration/file)
* [Kubernetes ConfigMap](https://gofeatureflag.org/../integrations/store-flags-configuration/kubernetes-configmap)
* [AWS S3](https://gofeatureflag.org/../integrations/store-flags-configuration/aws-s3)
* [MinIO](https://gofeatureflag.org/../integrations/store-flags-configuration/minio)
* [Google Cloud Storage](https://gofeatureflag.org/../integrations/store-flags-configuration/google-cloud-storage)
* [Azure Blob Storage](https://gofeatureflag.org/../integrations/store-flags-configuration/azure-blob-storage)
* [GitHub](https://gofeatureflag.org/../integrations/store-flags-configuration/github)
* [GitLab](https://gofeatureflag.org/../integrations/store-flags-configuration/gitlab)
* [Bitbucket](https://gofeatureflag.org/../integrations/store-flags-configuration/bitbucket)
* [MongoDB](https://gofeatureflag.org/../integrations/store-flags-configuration/mongodb)
* [Redis](https://gofeatureflag.org/../integrations/store-flags-configuration/redis)
* [Postgresql](https://gofeatureflag.org/../integrations/store-flags-configuration/postgresql)

### type `exporter`[​](#type-exporter "Direct link to type-exporter")

An [exporter](https://gofeatureflag.org/docs/concepts/exporter.md) is the component in charge of sending your evaluation data to a remote source.

example goff-proxy.yaml

```
# ...

exporters:

  - kind: s3

    bucket: evaluation-data-bucket

    flushInterval: 10000
```

info

Every `exporter` has its own configuration, check the [list of exporters](https://gofeatureflag.org/docs/integrations/export-evaluation-data.md#available-exporters) and go on the detail page of your choice to configure it.

**Available exporters**:

* [AWS S3](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-s3)
* [MinIO](https://gofeatureflag.org/../integrations/export-evaluation-data/minio)
* [Azure Blob Storage](https://gofeatureflag.org/../integrations/export-evaluation-data/azure-blob-storage)
* [File System](https://gofeatureflag.org/../integrations/export-evaluation-data/file)
* [Apache Kafka](https://gofeatureflag.org/../integrations/export-evaluation-data/kafka)
* [AWS Kinesis](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-kinesis)
* [Google Cloud Storage](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-storage)
* [Google Cloud PubSub](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-pubsub)
* [Google Cloud BigQuery](https://gofeatureflag.org/../integrations/export-evaluation-data/google-cloud-bigquery)
* [AWS SQS](https://gofeatureflag.org/../integrations/export-evaluation-data/aws-sqs)
* [Webhook](https://gofeatureflag.org/../integrations/export-evaluation-data/webhook)
* [Log](https://gofeatureflag.org/../integrations/export-evaluation-data/log)
* [OpenTelemetry](https://gofeatureflag.org/../integrations/export-evaluation-data/opentelemetry)

### type `flagSet`[​](#type-flagset "Direct link to type-flagset")

A flag set is a collection of flags that are used to evaluate features. It is used to group flags together and to apply the same configuration to them.

example goff-proxy.yaml

```
flagSets:

  - name: my-flagset

    apiKeys:

      - "my-api-key"

    retrievers:

      - kind: file

        path: /flags/my-flags.yaml

    exporters:

      - kind: s3

        bucket: my-evaluations

    notifiers:

      - kind: slack

        webhookUrl: "https://hooks.slack.com/services/..."
```

#### `flagSet.name`[​](#flagsetname "Direct link to flagsetname")

The name of the flag set. This allows you to identify the flag set.

* option name: `name`
* type: **string**
* default: **auto-generated UUID**
* mandatory:**

warning

If you want to **add/remove the flag set** or **update its API keys** at runtime without restarting the relay proxy, you **must provide a `name`** for your flag set.<br /><!-- -->Without a name (or using `default` as the name), a UUID is generated and the flag set cannot be managed at runtime.

#### `flagSet.apiKeys`[​](#flagsetapikeys "Direct link to flagsetapikeys")

The API keys for the flag set. This will add new API keys to the list of authorized keys for evaluation. This property is mandatory for every flag set, as it is used to filter the flags available.

* option name: `apiKeys`
* type: **\[]string**
* default: **none**
* mandatory:**

#### `flagSet.retrievers`[​](#flagsetretrievers "Direct link to flagsetretrievers")

Configuration for retrieving flag configurations. Supports the same options as the main configuration.

* [*see `retrievers`*](#retrievers)

#### `flagSet.exporters`[​](#flagsetexporters "Direct link to flagsetexporters")

Configuration for exporting evaluation data. Supports the same options as the main configuration.

* [*see `exporters`*](#exporters)

#### `flagSet.notifiers`[​](#flagsetnotifiers "Direct link to flagsetnotifiers")

Configuration for flag change notifications. Supports the same options as the main configuration.

* [*see `notifiers`*](#notifiers)

#### `flagSet.fileFormat`[​](#flagsetfileformat "Direct link to flagsetfileformat")

The format of the flag configuration files (yaml, json, toml).

* [*see `fileFormat`*](#fileformat)

#### `flagSet.pollingInterval`[​](#flagsetpollinginterval "Direct link to flagsetpollinginterval")

The time interval in milliseconds when the relay proxy is reloading the configuration file.

* [*see `pollingInterval`*](#pollinginterval)

#### `flagSet.startWithRetrieverError`[​](#flagsetstartwithretrievererror "Direct link to flagsetstartwithretrievererror")

If true, the flag set will start even if the retriever fails initially.

* [*see `startWithRetrieverError`*](#startwithretrievererror)

#### `flagSet.enablePollingJitter`[​](#flagsetenablepollingjitter "Direct link to flagsetenablepollingjitter")

Set to true if you want to avoid having true periodicity when retrieving your flags.

* [*see `enablePollingJitter`*](#enablepollingjitter)

#### `flagSet.disableNotifierOnInit`[​](#flagsetdisablenotifieroninit "Direct link to flagsetdisablenotifieroninit")

Set to true if you do not want to call any notifier when the flags are loaded for the first time.

* [*see `disableNotifierOnInit`*](#disablenotifieroninit)

#### `flagSet.evaluationContextEnrichment`[​](#flagsetevaluationcontextenrichment "Direct link to flagsetevaluationcontextenrichment")

Additional context data to merge with all evaluation requests for this flag set.

* [*see `evaluationContextEnrichment`*](#evaluationcontextenrichment)

#### `flagSet.persistentFlagConfigurationFile`[​](#flagsetpersistentflagconfigurationfile "Direct link to flagsetpersistentflagconfigurationfile")

File path to persist flag configurations for this flag set.

* [*see `persistentFlagConfigurationFile`*](#persistentflagconfigurationfile)

#### `flagSet.environment`[​](#flagsetenvironment "Direct link to flagsetenvironment")

Environment identifier for this flag set (e.g., "dev", "staging", "prod").

* [*see `environment`*](#environment)

### type `notifier`[​](#type-notifier "Direct link to type-notifier")

A [notifier](https://gofeatureflag.org/docs/concepts/notifier.md) is the component in charge of sending a notification when a flag changes to a remote system.

example goff-proxy.yaml

```
# ...

notifiers:

  - kind: discord

    webhookUrl: "https://discord.com/api/webhooks/yyyy/xxxxxxx"

  - kind: slack

    webhookUrl: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"

# ...
```

info

Every `notifier` has its own configuration, check the [list of notifiers](https://gofeatureflag.org/docs/integrations/notify-flags-changes.md#available-notifiers) and go on the detail page of your choice to configure it.

**Available notifiers**:

* [Slack](https://gofeatureflag.org/../integrations/notify-flags-changes/slack)
* [Discord](https://gofeatureflag.org/../integrations/notify-flags-changes/discord)
* [Microsoft Teams](https://gofeatureflag.org/../integrations/notify-flags-changes/microsoft-teams)
* [Webhook](https://gofeatureflag.org/../integrations/notify-flags-changes/webhook)

### type `swagger`[​](#type-swagger "Direct link to type-swagger")

Configuration related to the swagger behavior of the relay proxy.

#### `swagger.enabled`[​](#swaggerenabled "Direct link to swaggerenabled")

Enables Swagger for testing the APIs directly. If you are enabling Swagger you will have to provide the [`swagger.host`](#swaggerhost) configuration (default is `localhost`).

It will start the swagger server on the same port as the relay proxy and you can access it at `http://<host>:<server.port>/swagger/`.

* option name: `enabled`
* type: **boolean**
* default: **`false`**
* mandatory:**

#### `swagger.host`[​](#swaggerhost "Direct link to swaggerhost")

The DNS used by the swagger configuration to access the relay proxy.

* option name: `host`
* type: **string**
* default: **`localhost`**
* mandatory:**

### type `ofrepEventStream`[​](#type-ofrepeventstream "Direct link to type-ofrepeventstream")

Configuration used to advertise the flag-change SSE endpoint in the OFREP bulk evaluation response.

When [`ofrepEventStream.baseUrl`](#ofrepeventstreambaseurl) is set, the bulk evaluation response contains an `eventStreams` field pointing to the relay proxy SSE endpoint (`/stream/v1/sse/flag/change`).

Each time the flag configuration changes, this endpoint sends an ADR-0008 event that tells the provider to re-fetch its evaluations:

```
id: 1771622898000000000

event: message

data: {"type":"refetchEvaluation","lastModified":1771622898}
```

The stream also sends SSE comment lines, which spec-compliant clients ignore: `: connected` right after the connection is opened, and `: heartbeat` every 30 seconds so that proxies do not close idle connections.

If the OFREP request is authenticated, the API key of the caller is added to the advertised URL as the `apiKey` query parameter, so the provider can open the stream without extra credentials. As required by the ADR, this URL is sensitive and must not be logged or persisted.

#### `ofrepEventStream.baseUrl`[​](#ofrepeventstreambaseurl "Direct link to ofrepeventstreambaseurl")

The public base URL clients should use to reach the relay proxy, for example `https://gofeatureflag.example.com`. The path of the SSE endpoint is appended automatically, so a path prefix is kept (`https://example.com/goff` becomes `https://example.com/goff/stream/v1/sse/flag/change`).

When not set, the `eventStreams` field is not added and OFREP providers keep polling. Do not set it when the relay proxy runs as an AWS Lambda, because Lambda cannot stream SSE.

It must be an absolute `http` or `https` URL, without query parameters.

* option name: `baseUrl`
* type: **string**
* default: **none**
* mandatory:**

#### `ofrepEventStream.inactivityDelaySec`[​](#ofrepeventstreaminactivitydelaysec "Direct link to ofrepeventstreaminactivitydelaysec")

The delay in seconds advertised to the client, after which it should consider the stream inactive. When `0` the field is omitted.

* option name: `inactivityDelaySec`
* type: **integer**
* default: **`0`**
* mandatory:**

## Use multiple flag sets[​](#use-multiple-flag-sets "Direct link to Use multiple flag sets")

Flag sets allow you to organize your feature flags into separate groups, each with its own configuration, API keys, retrievers, exporters, and notifiers. This is particularly useful for:

* **Team isolation**: Different teams can manage their own flags independently
* **Environment separation**: Separate flags for different environments (dev, staging, prod)
* **Access control**: Different API keys for different applications or teams
* **Independent configurations**: Each flag set can have its own retrievers, exporters, and notifiers

### How flag sets work[​](#how-flag-sets-work "Direct link to How flag sets work")

When you configure flag sets, the relay proxy operates in **flag set mode** instead of the default single configuration mode. In flag set mode:

* Each flag set requires at least one API key
* API keys are used to route requests to the appropriate flag set
* Each flag set can have its own `retrievers`, `exporters`, and `notifiers`
* There is no inheritance between flag sets - each is completely independent

### Basic flag set configuration[​](#basic-flag-set-configuration "Direct link to Basic flag set configuration")

goff-proxy.yaml

```
server:

  mode: http

  port: 1031

flagSets:

  - name: teamA

    apiKeys:

      - "team-a-api-key-1"

      - "team-a-api-key-2"

    retrievers:

      - kind: file

        path: /path/to/team-a-flags.yaml

    exporters:

      - kind: log



  - name: teamB

    apiKeys:

      - "team-b-api-key"

    retrievers:

      - kind: s3

        bucket: my-feature-flag-bucket

        item: team-b/flags.yaml

    exporters:

      - kind: s3

        bucket: evaluation-data-bucket

    notifiers:

      - kind: slack

        webhookUrl: "https://hooks.slack.com/services/..."
```

### Flag set configuration options[​](#flag-set-configuration-options "Direct link to Flag set configuration options")

warning

If you are using flag sets, all top level configuration options will be ignored.

Each flag set supports all the same configuration options as the main relay proxy configuration. You can find the list of options in the [flag set types](#type-flagset) section.

### Using flag sets with API keys[​](#using-flag-sets-with-api-keys "Direct link to Using flag sets with API keys")

When using flag sets, all evaluation requests must include an API key.<br /><!-- -->You can do so by adding the API key in the initialization of your providers.

The relay proxy will:

1. Extract the API key from the `Authorization` header
2. Find the flag set associated with that API key
3. Route the evaluation request to that flag set
4. Return the evaluation result from that flag set's flags

### Runtime API key updates[​](#runtime-api-key-updates "Direct link to Runtime API key updates")

GO Feature Flag supports updating API keys at runtime without restarting the relay proxy.<br /><!-- -->See the [Runtime configuration updates](#runtime-configuration-updates) section for detailed information about runtime updates in flagset mode.

warning

To enable runtime API key updates, your flag sets **must have a `name` configured**.<br /><!-- -->Without a name, a UUID is generated automatically and runtime updates won't work.

```
flagSets:

  - name: team-a  # ✅ Required for runtime API key updates

    apiKeys:

      - "team-a-api-key"

    # ...
```

### Example: Multi-team setup[​](#example-multi-team-setup "Direct link to Example: Multi-team setup")

goff-proxy.yaml

```
server:

  mode: http

  port: 1031



logLevel: info



flagSets:

  # Frontend team flags

  - name: frontend

    apiKeys:

      - "frontend-api-key"

    retrievers:

      - kind: file

        path: /flags/frontend-flags.yaml

    exporters:

      - kind: s3

        bucket: frontend-evaluations

    notifiers:

      - kind: slack

        webhookUrl: "https://hooks.slack.com/services/frontend-webhook"



  # Backend team flags

  - name: backend

    apiKeys:

      - "backend-api-key"

    retrievers:

      - kind: github

        repository: my-org/backend-flags

        path: flags.yaml

    exporters:

      - kind: kafka

        topic: backend-evaluations

    notifiers:

      - kind: discord

        webhookUrl: "https://discord.com/api/webhooks/backend-webhook"



  # Mobile team flags

  - name: mobile

    apiKeys:

      - "mobile-api-key"

    retrievers:

      - kind: s3

        bucket: mobile-feature-flags

        item: flags.yaml

    exporters:

      - kind: file

        outputDir: /logs/mobile-evaluations
```

## Runtime configuration updates[​](#runtime-configuration-updates "Direct link to Runtime configuration updates")

GO Feature Flag relay proxy supports updating certain configuration options at runtime without requiring a restart.<br /><!-- -->When the configuration file changes, the relay proxy automatically detects and applies the changes.

### Default mode runtime updates[​](#default-mode-runtime-updates "Direct link to Default mode runtime updates")

In **default mode** (when not using flag sets), you can update the following configuration at runtime:

* **`authorizedKeys.evaluation`**: API keys for evaluation endpoints
* **`authorizedKeys.admin`**: API keys for admin endpoints
* **`apiKeys`**: Deprecated field (still supported for backward compatibility)

warning

In default mode, **only API keys can be updated at runtime**.<br /><!-- -->All other configuration changes (retrievers, exporters, notifiers, polling intervals, etc.) are not supported and will be ignored with a warning log message.

**Example:**

goff-proxy.yaml

```
# Initial configuration

authorizedKeys:

  evaluation:

    - "key-1"

    - "key-2"

  admin:

    - "admin-key-1"



# Updated configuration (runtime update)

authorizedKeys:

  evaluation:

    - "key-1"

    - "key-2"

    - "key-3"  # ✅ New key added

  admin:

    - "admin-key-1"

    - "admin-key-2"  # ✅ New admin key added
```

### Flagset mode runtime updates[​](#flagset-mode-runtime-updates "Direct link to Flagset mode runtime updates")

In **flagset mode** (when using flag sets), you can update API keys for each flag set at runtime.

**Supported operations:**

* Adding new API keys to an existing flag set
* Removing API keys from a flag set
* Replacing API keys with new ones

warning

To enable runtime API key updates, your flag sets **must have a `name` configured**.<br /><!-- -->Without a name, a UUID is generated automatically and runtime updates won't work.

```
flagSets:

  - name: team-a  # ✅ Required for runtime API key updates

    apiKeys:

      - "team-a-api-key"

    # ...
```

### How runtime updates work[​](#how-runtime-updates-work "Direct link to How runtime updates work")

1. The relay proxy detects the configuration file change
2. It validates the new configuration *(if the configuration is invalid, the relay proxy will not apply the changes)*
3. It updates the internal API key mappings
4. New requests immediately use the updated routing

info

During the update, there is no downtime. Requests continue to be processed normally, and the new API key configuration takes effect immediately after the update is applied.
