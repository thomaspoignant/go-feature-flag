---
sidebar_position: 30
title: Configuration
description: How to configure the relay proxy to serve your feature flags.
---

# Configure the relay proxy

## Getting Started

The configuration of the **relay proxy** is based on a configuration file that you have to provide.

The only mandatory information you need to start the server is to provide where to retrieve your feature flags
configuration.

```yaml
retriever:
  kind: file
  path: /goff/flags.yaml # Location of your feature flag files
```

## Global configuration

:::tip Use environment variables.
You can override file configurations using environment variables.

Note that all environment variables should be uppercase.
If you want to replace a nested fields, please use `_` to separate each field _(ex: `RETRIEVER_KIND`)_.

In case of an array of string, you can add multiple values separated by a comma _(
ex: `AUTHORIZEDKEYS_EVALUATION=my-first-key,my-second-key`)_.
:::

| Field name                        | Type                                   | Default            | Description                                                                                                                                                                                                                                                                                                                                                                                                                                      |
|-----------------------------------|----------------------------------------|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `retriever`                       | [retriever](#retriever)                | **none**           | **(mandatory)** This is the configuration on how to retrieve the configuration of the files.<br /><br />_Note: this field is mandatory only if `retrievers` is not set._                                                                                                                                                                                                                                                                         |                                                                                                                                                                                                  |
| `retrievers`                      | [[]retriever](#retriever)              | **none**           | **(mandatory)** Exactly the same things as `retriever` except that you can provide more than 1 retriever.<br /><br />_Note: this field is mandatory only if `retriever` is not set._                                                                                                                                                                                                                                                             |
| `listen`                          | int                                    | `1031`             | This is the port used by the relay proxy when starting the HTTP server.                                                                                                                                                                                                                                                                                                                                                                          |
| `pollingInterval`                 | int                                    | `60000`            | This is the time interval **in millisecond** when the relay proxy is reloading the configuration file.<br/>The minimum time accepted is 1000 millisecond.                                                                                                                                                                                                                                                                                        |
| `enablePollingJitter`             | boolean                                | `false`            | Set to true if you want to avoid having true periodicity when retrieving your flags. It is useful to avoid having spike on your flag configuration storage in case your application is starting multiple instance at the same time.<br/>We ensure a deviation that is maximum ±10% of your polling interval.<br />Default: false                                                                                                                 |
| `DisableNotifierOnInit`           | boolean                                | `false`            | If **true**, the relay proxy will not call any notifier when the flags are loaded during initialization. This is useful if you do not want a Slack/Webhook notification saying that the flags have been added every time you start the proxy.<br/>Default: **false**                                                                                                                                                                             |
| `hideBanner`                      | boolean                                | `false`            | Should we display the beautiful **go-feature-flag** banner when starting the relay proxy                                                                                                                                                                                                                                                                                                                                                         |
| `enableSwagger`                   | boolean                                | `false`            | Enables Swagger for testing the APIs directly. If you are enabling Swagger you will have to provide the `host` configuration and the Swagger UI will be available at `http://<host>:<listen>/swagger/`.                                                                                                                                                                                                                                          |
| `host`                            | string                                 | `localhost`        | This is the DNS you will use to access the relay proxy. This field is used by Swagger to query the API at the right place.                                                                                                                                                                                                                                                                                                                       |
| `restApiTimeout`                  | int                                    | `5000`             | Timeout in milliseconds for API calls.                                                                                                                                                                                                                                                                                                                                                                                                           |
| `logLevel`                        | string                                 | `info`             | The log level to use for the relay proxy.<br/> Available values are `ERROR`, `WARN`, `INFO`, `DEBUG`.                                                                                                                                                                                                                                                                                                                                            |
| `fileFormat`                      | string                                 | `yaml`             | This is the format of your `go-feature-flag` configuration file. Acceptable values are `yaml`, `json`, `toml`.                                                                                                                                                                                                                                                                                                                                   |
| `startWithRetrieverError`         | boolean                                | `false`            | By default the **relay proxy** will crash if it is not able to retrieve the flags from the configuration.<br/>If you don't want your relay proxy to crash, you can set `startWithRetrieverError` to true. Until the flag is retrievable the relay proxy will only answer with default values.                                                                                                                                                    |
| `exporter`                        | [exporter](#exporter)                  | **none**           | Exporter is the configuration used to export data.                                                                                                                                                                                                                                                                                                                                                                                               |
| `notifier`                        | [notifier](#notifier)                  | **none**           | Notifiers is the configuration on where to notify a flag change.                                                                                                                                                                                                                                                                                                                                                                                 |
| `authorizedKeys`                  | [authorizedKeys](#type-authorizedkeys) | **none**           | List of authorized API keys.                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `evaluationContextEnrichment`     | object                                 | **none**           | It is a free field that will be merged with the evaluation context sent during the evaluation. It is useful to add common attributes to all the evaluations, such as a server version, environment, etc.<br/><br/>These fields will be included in the custom attributes of the evaluation context.<br/><br/>If in the evaluation context you have a field with the same name, it will be override by the `evaluationContextEnrichment`.         |
| `openTelemetryOtlpEndpoint`       | string                                 | **none**           | Endpoint of your OpenTelemetry OTLP collector, used to send traces to it and you will be able to forward them to your OpenTelemetry solution with the appropriate provider.                                                                                                                                                                                                                                                                      |
| `kafka`                           | object                                 | **none**           | Settings for the Kafka exporter. Mandatory when using the 'kafka' exporter type, and ignored otherwise.                                                                                                                                                                                                                                                                                                                                          |                     
| `projectID`                       | string                                 | **none**           | ID of GCP project. Mandatory when using PubSub exporter.                                                                                                                                                                                                                                                                                                                                                                                         |
| `topic`                           | string                                 | **none**           | Name of PubSub topic on which messages will be published. Mandatory when using PubSub exporter.                                                                                                                                                                                                                                                                                                                                                  |
| `persistentFlagConfigurationFile` | string                                 | **none**           | If set GO Feature Flag will store the flags configuration in this file to be able to serve the flags even if none of the retrievers is available during starting time.<br/>By default, the flag configuration is not persisted and stays on the retriever system. By setting a file here, you ensure that GO Feature Flag will always start with a configuration but which can be out-dated.<br/><br/>_(example: `/tmp/goff_persist_conf.yaml`)_ |
| `startAsAwsLambda`                | boolean                                | **`false`**        | If set GO Feature Flag start the relay-proxy as a AWS Lambda, it means that it will start the server to receive request in the AWS format _(see `awsLambdaAdapter` to set the request/response format you are using)_.                                                                                                                                                                                                                           |
| `awsLambdaAdapter`                | string                                 | **`APIGatewayV2`** | This param is used only if `startAsAwsLambda` is `true`.<br/>This parameter allow you to decide which type of AWS lambda handler you wan to use.<br/>Accepted values are `APIGatewayV2`, `APIGatewayV1`, `ALB`.                                                                                                                                                                                                                                  |

## type `authorizedKeys`

To be able to control who can access your relay proxy, you can set a list of authorized keys.

| Field name   | Type     | Default  | Description                                                                                                                                                                                                                                                                            |
|--------------|----------|----------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `evaluation` | []string | **none** | If set, we will check for each evaluation if an authorized key is provided.<br/>Each request will need to provide one of authorized key inside `Authorization` header with format `Bearer <api-key>`.<br /><br />_Note: there will be no authorization when this config is not set._   |
| `admin`      | []string | **none** | You need to set API keys in this field if you want to access the `/v1/admin/*` endpoints.<br/> If no api key is configured the endpoint will be unreachable.<br/>Each request will need to provide one of authorized key inside `Authorization` header with format `Bearer <api-key>`. |

<a name="retriever"></a>

## type `retriever`

`go-feature-flag` supports different kind of retriever types such as S3, Google store, etc ...
In this section we will present all the available retriever configurations available.

### S3

If you are using the S3 provider, the easiest way to provide credentials is to set environment variables.
It will be used by GO Feature Flag to identify to your S3 bucket.

```shell
export AWS_SECRET_ACCESS_KEY=xxxx
export AWS_ACCESS_KEY_ID=xxx
export AWS_DEFAULT_REGION=eu-west-1
```

| Field name | Type   | Default  | Description                                                                                                          |
|------------|--------|----------|----------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `bucket`   | string | **none** | **(mandatory)** This is the name of your S3 bucket _(ex: `my-featureflag-bucket`)_.                                  |
| `item`     | string | **none** | **(mandatory)** Path to the file inside the bucket _(ex: `config/flag/my-flags.yaml`)_.                              |

### GitHub

:::tip
GitHub has rate limits, be sure to correctly set your `PollingInterval` to avoid reaching the limit.

If the rate limit is reached, the retriever will return an error and will stop polling until GitHub allows it again.
:::

| Field name       | Type   | Default  | Description                                                                                                                                                                                                                        |
|------------------|--------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`           | string | **none** | **(mandatory)** Value should be **`github`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                           |
| `repositorySlug` | string | **none** | **(mandatory)** The repository slug of the GitHub repository where your file is located _(ex: `thomaspoignant/go-feature-flag`)_.                                                                                                  |
| `path`           | string | **none** | **(mandatory)** Path to the file inside the repository _(ex: `config/flag/my-flags.yaml`)_.                                                                                                                                        |
| `branch`         | string | `main`   | The branch we should check in the repository.                                                                                                                                                                                      |
| `token`          | string | **none** | Github token used to access a private repository, you need the repo permission ([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)). |
| `timeout`        | string | `10000`  | Timeout in millisecond used when calling GitHub.                                                                                                                                                                                   |

### GitLab

| Field name       | Type   | Default              | Description                                                                                                                                                                                              |
|------------------|--------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`           | string | **none**             | **(mandatory)** Value should be **`gitlab`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                 |
| `repositorySlug` | string | **none**             | **(mandatory)** The repository slug of the GitLab repository where your file is located _(ex: `thomaspoignant/go-feature-flag`)_.                                                                        |
| `path`           | string | **none**             | **(mandatory)** Path to the file inside the repository _(ex: `config/flag/my-flags.yaml`)_.                                                                                                              |
| `baseUrl`        | string | `https://gitlab.com` | The base URL of your GitLab instance.                                                                                                                                                                    |
| `branch`         | string | `main`               | The branch we should check in the repository.                                                                                                                                                            |
| `token`          | string | **none**             | GitLab personal access token used to access a private repository ([Create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token)). |
| `timeout`        | string | `10000`              | Timeout in millisecond used when calling GitLab.                                                                                                                                                         |

### File

| Field name | Type   | Default  | Description                                                                                                            |
|------------|--------|----------|------------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`file`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `path`     | string | **none** | **(mandatory)** Path to the file in your local computer _(ex: `/goff/my-flags.yaml`)_.                                 |

### HTTP

| Field name | Type                | Default  | Description                                                                                                            |
|------------|---------------------|----------|------------------------------------------------------------------------------------------------------------------------|
| `kind`     | string              | **none** | **(mandatory)** Value should be **`http`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `url`      | string              | **none** | **(mandatory)** Location to retrieve the file.                                                                         |
| `method`   | string              | `GET`    | The HTTP Method you are using to call the HTTP endpoint.                                                               |
| `body`     | string              | **none** | The HTTP Body you are using to call the HTTP endpoint.                                                                 |
| `headers`  | map[string][]string | **none** | The HTTP headers used when calling the HTTP endpoint (useful for authorization).                                       |
| `timeout`  | string              | `10000`  | Timeout in millisecond when calling the HTTP endpoint.                                                                 |

### Google Storage

| Field name | Type   | Default  | Description                                                                                                                     |
|------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`googleStorage`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `bucket`   | string | **none** | **(mandatory)** This is the name of your Google Storage bucket _(ex: `my-featureflag-bucket`)_.                                 |
| `object`   | string | **none** | **(mandatory)** Path to the file inside the bucket _(ex: `config/flag/my-flags.yaml`)_.                                         |

### Kubernetes ConfigMap

_Note that relay proxy is only supporting this while running inside the kubernetes cluster._

| Field name  | Type   | Default  | Description                                                                                                                 |
|-------------|--------|----------|-----------------------------------------------------------------------------------------------------------------------------|
| `kind`      | string | **none** | **(mandatory)** Value should be **`configmap`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `namespace` | string | **none** | **(mandatory)** This is the name of the namespace where your **configmap** is located _(ex: `default`)_.                    |
| `configmap` | string | **none** | **(mandatory)** Name of the **configmap** we should read  _(ex: `feature-flag`)_.                                           |
| `key`       | string | **none** | **(mandatory)** Name of the `key` in the **configmap** which contains the flag.                                             |

### MongoDB

_To understand the format in which a flag needs to be configured in MongoDB, check
the [example](https://github.com/thomaspoignant/go-feature-flag/examples/retriever_mongodb) available._

| Field name   | Type   | Default  | Description                                                                                                               |
|--------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------|
| `kind`       | string | **none** | **(mandatory)** Value should be **`mongodb`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `uri`        | string | **none** | **(mandatory)** This is the MongoDB URI used in order to connect to the MongoDB instance.                                 |
| `database`   | string | **none** | **(mandatory)** Name of the **database** where flags are stored.                                                          |
| `collection` | string | **none** | **(mandatory)** Name of the **collection** where flags are stored.                                                        |

### Redis

_To understand the format in which a flag needs to be configured in **Redis**, check
the [doc](../go_module/store_file/redis#expected-format) available._

| Field name | Type   | Default  | Description                                                                                                                                                                                                                                           |
|------------|--------|----------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`redis`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                               |
| `options`  | object | **none** | **(mandatory)** Options used to connect to your redis instance.<br/>All the options from the `go-redis` SDK are available _([check `redis.Options`](https://github.com/redis/go-redis/blob/683f4fa6a6b0615344353a10478548969b09f89c/options.go#L31))_ |
| `prefix`   | string | **none** | Prefix used before your flag name in the Redis DB.                                                                                                                                                                                                    |

### Bitbucket

| Field name       | Type   | Default              | Description                                                                                                                                                                       |
|------------------|--------|----------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`           | string | **none**             | **(mandatory)** Value should be **`bitbucket`**.<br/>_This field is mandatory and describes which retriever you are using._                                                       |
| `repositorySlug` | string | **none**             | **(mandatory)** The repository slug of the Bitbucket repository where your file is located _(ex: `thomaspoignant/go-feature-flag`)_.                                              |
| `path`           | string | **none**             | **(mandatory)** Path to the file inside the repository _(ex: `config/flag/my-flags.yaml`)_.                                                                                       |
| `baseUrl`        | string | `https://gitlab.com` | The base URL of your Bitbucket instance<br/>By default we are using the public API `https://api.bitbucket.org`.                                                                   |
| `branch`         | string | `main`               | The branch we should check in the repository.                                                                                                                                     |
| `token`          | string | **none**             | Bitbucket token used to access a private repository ([_Create a Repository Access Token_](https://support.atlassian.com/bitbucket-cloud/docs/create-a-repository-access-token/)). |
| `timeout`        | string | `10000`              | Timeout in millisecond used when calling GitLab.                                                                                                                                  |

<a name="exporter"></a>

## type `exporter`

### Webhook

| Field name         | Type                | Default  | Description                                                                                                                                                                                                                 |
|--------------------|---------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string              | **none** | **(mandatory)** Value should be **`webhook`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                   |
| `endpointUrl`      | string              | **none** | **(mandatory)** EndpointURL of your webhook.                                                                                                                                                                                |
| `flushInterval`    | int                 | `60000`  | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                               |
| `maxEventInMemory` | int                 | `100000` | If we hit that limit we will call the webhook.                                                                                                                                                                              |
| `secret`           | string              | **none** | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br/> See [signature section](https://thomaspoignant.github.io/go-feature-flag/latest/data_collection/webhook/#signature) for more details. |
| `meta`             | map[string]string   | **none** | Add all the information you want to see in your request.                                                                                                                                                                    |
| `headers`          | map[string][]string | **none** | Add all the headers you want to add while calling the endpoint                                                                                                                                                              |

### File

| Field name                | Type   | Default                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                    |
|---------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`                    | string | **none**                                                                                                                           | **(mandatory)** Value should be **`file`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                                                                                                                                                                         |
| `outputDir`               | string | **none**                                                                                                                           | **(mandatory)** OutputDir is the location of the directory where to store the exported files.                                                                                                                                                                                                                                                                                                  |
| `flushInterval`           | int    | `60000`                                                                                                                            | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                  |
| `maxEventInMemory`        | int    | `100000`                                                                                                                           | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                 |
| `format`                  | string | `JSON`                                                                                                                             | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                        |
| `filename`                | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                       | You can use a templated config to define the name of your exported files. Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                     |
| `csvTemplate`             | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using format other than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see the fields available. |
| `parquetCompressionCodec` | string | `SNAPPY`                                                                                                                           | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)                                                                                                                                                                                                                 |

### Log

| Field name         | Type   | Default                                                                             | Description                                                                                                                                                                                                                                                  |
|--------------------|--------|-------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string | **none**                                                                            | **(mandatory)** Value should be **`log`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                                        |
| `flushInterval`    | int    | `60000`                                                                             | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval, we will call the webhook before)_.                                                                                               |
| `maxEventInMemory` | int    | `100000`                                                                            | If we hit that limit we will call the webhook.                                                                                                                                                                                                               |
| `logFormat`        | string | `[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"` | LogFormat is the [template](https://golang.org/pkg/text/template/) configuration of the output format of your log.<br/>You can use all the key from the exporter.FeatureEvent + a key called FormattedDate that represent the date with the RFC 3339 Format. |

### S3

| Field name                | Type   | Default                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                             |
|---------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`                    | string | **none**                                                                                                                           | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                                                                                                                                                                                    |
| `bucket`                  | string | **none**                                                                                                                           | **(mandatory)** Name of your S3 Bucket.                                                                                                                                                                                                                                                                                                                                                                 |
| `flushInterval`           | int    | `60000`                                                                                                                            | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                           |
| `maxEventInMemory`        | int    | `100000`                                                                                                                           | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                          |
| `format`                  | string | `JSON`                                                                                                                             | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                                 |
| `filename`                | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                       | You can use a config template to define the name of your exported files. Available replacements are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                               |
| `csvTemplate`             | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using format other than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available. |`
| `path`                    | string | **bucket root level**                                                                                                              | The location of the directory in S3.                                                                                                                                                                                                                                                                                                                                                                    |
| `parquetCompressionCodec` | string | `SNAPPY`                                                                                                                           | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)                                                                                                                                                                                                                          |`

### Google Storage

| Field name                | Type   | Default                                                                                                                            | Description                                                                                                                                                                                                                                                                                                                                                                                             |
|---------------------------|--------|------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`                    | string | **none**                                                                                                                           | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                                                                                                                                                                                    |
| `bucket`                  | string | **none**                                                                                                                           | **(mandatory)** Name of your Google Cloud Storage Bucket.                                                                                                                                                                                                                                                                                                                                               |
| `flushInterval`           | int    | `60000`                                                                                                                            | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                           |
| `maxEventInMemory`        | int    | `100000`                                                                                                                           | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                          |
| `format`                  | string | `JSON`                                                                                                                             | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                                 |
| `filename`                | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                                       | You can use a templated config to define the name of your exported files. Available replacement are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                               |
| `csvTemplate`             | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using format other than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available. |`
| `path`                    | string | **bucket root level**                                                                                                              | The location of the directory in S3.                                                                                                                                                                                                                                                                                                                                                                    |
| `parquetCompressionCodec` | string | `SNAPPY`                                                                                                                           | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md)                                                                                                                                                                                                                          |`

### SQS

| Field name | Type   | Default  | Description                                                                                                           |
|------------|--------|----------|-----------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`sqs`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `queueUrl` | string | **none** | **(mandatory)** URL of your SQS queue.<br/>_You can find it in your AWS console._                                     |

### Kafka

| Field name        | Type     | Default           | Description                                                                                                                                                                                                                                                                                                |
|-------------------|----------|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`            | string   | **none**          | **(mandatory)** Value should be **`kafka`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                                                                                                    |
| `kafka.topic`     | string   | **none**          | **(mandatory)** Kafka topic to bind to.                                                                                                                                                                                                                                                                    |
| `kafka.addresses` | []string | **none**          | **(mandatory)** List of bootstrap addresses for the Kafka cluster.                                                                                                                                                                                                                                         |
| `kafka.config`    | object   | _see description_ | This field allows fine tuning of the Kafka reader. This object should contain the [Sarama configuration](https://pkg.go.dev/github.com/IBM/sarama#Config) that the reader will use. On empty, a sensible default is created using [sarama.NewConfig()](https://pkg.go.dev/github.com/IBM/sarama#NewConfig) |

### Google PubSub

| Field name  | Type   | Default  | Description                                                                                                              |
|-------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------|
| `kind`      | string | **none** | **(mandatory)** Value should be **`pubsub`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `projectID` | string | **none** | **(mandatory)** Value should be ID of GCP project you are using.                                                         |
| `topic`     | string | **none** | **(mandatory)** Topic name on which messages will be published.                                                          |

### AWS Kinesis

| Field name   | Type   | Default  | Description                                                                                                               |
|--------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------|
| `kind`       | string | **none** | **(mandatory)** Value should be **`kinesis`**.<br/>_This field is mandatory and describes which retriever you are using._ |
| `streamArn`  | string | **none** | **(mandatory)**  The ARN of your kinesis stream.                                                                          |
| `streamName` | string | **none** | The name of your kinesis stream.                                                                                          |

<a name="notifier"></a>

## type `notifier`

### Slack

| Field name   | Type   | Default  | Description                                                                                                            |
|--------------|--------|----------|------------------------------------------------------------------------------------------------------------------------|
| `kind`       | string | **none** | **(mandatory)** Value should be **`slack`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `webhookUrl` | string | **none** | **(mandatory)** The complete URL of your incoming webhook configured in Slack.                                         |

### Discord

| Field name   | Type   | Default  | Description                                                                                                              |
|--------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------|
| `kind`       | string | **none** | **(mandatory)** Value should be **`discord`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `webhookUrl` | string | **none** | **(mandatory)** The complete URL of your incoming webhook configured in Discord.                                         |

### Microsoft Teams

| Field name   | Type   | Default  | Description                                                                                                                     |
|--------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------------|
| `kind`       | string | **none** | **(mandatory)** Value should be **`microsoftteams`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `webhookUrl` | string | **none** | **(mandatory)** The complete URL of your incoming webhook configured in Discord.                                                |

### Webhook

| Field name    | Type                | Default  | Description                                                                                                                                                                                                                 |
|---------------|---------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`        | string              | **none** | **(mandatory)** Value should be **`webhook`**.<br/>_This field is mandatory and describes which retriever you are using._                                                                                                   |
| `endpointUrl` | string              | **none** | **(mandatory)** The complete URL of your API (we will send a POST request to this URL, see [format](https://thomaspoignant.github.io/go-feature-flag/latest/notifier/webhook/#format)                                       |
| `secret`      | string              | **none** | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br/> See [signature section](https://thomaspoignant.github.io/go-feature-flag/latest/data_collection/webhook/#signature) for more details. |
| `meta`        | map[string]string   | **none** | Add all the information you want to see in your request.                                                                                                                                                                    |
| `headers`     | map[string][]string | **none** | Add all the headers you want to add while calling the endpoint                                                                                                                                                              |
