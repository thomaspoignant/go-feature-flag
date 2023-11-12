---
sidebar_position: 30
title: Configuration
description: How to configure the relay proxy to serve your feature flags.
---

# Configure the relay proxy

## Getting Started
The configuration of the **relay proxy** is based on a configuration file that you have to provide.

The only mandatory information you need to start the server is to provide where to retrieve your feature flags configuration.

```yaml
retriever:
  kind: file
  path: /goff/flags.yaml # Location of your feature flag files
```

## Global configuration

:::tip Use environment variables.
You can also override these file configuration by using environment variables.

Note that all environment variables should be uppercase.  
If you want to replace a nested fields, please use `_` to separate each field _(ex: `RETRIEVER_KIND`)_.
:::


| Field name                | Type                      | Default     | Description                                                                                                                                                                                                                                                                                                                            |
|---------------------------|---------------------------|-------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `retriever`               | [retriever](#retriever)   | **none**    | **(mandatory)** This is the configuration on how to retrieve the configuration of the files.<br /><br />_Note: this field is mandatory only if `retrievers` is not set._                                                                                                                                                               |                                                                                                                                                                                                  |
| `retrievers`              | [[]retriever](#retriever) | **none**    | **(mandatory)** Exactly the same things as `retriever` except that you can provide more than 1 retriever.<br /><br />_Note: this field is mandatory only if `retriever` is not set._                                                                                                                                                   |
| `listen`                  | int                       | `1031`      | This is the port used by the relay proxy when starting the HTTP server.                                                                                                                                                                                                                                                                |
| `pollingInterval`         | int                       | `60000`     | This is the time interval **in millisecond** when the relay proxy is reloading the configuration file.<br/>The minimum time accepted is 1000 millisecond.                                                                                                                                                                              |
| `enablePollingJitter`     | boolean                   | false       | Set to true if you want to avoid having true periodicity when retrieving your flags. It is useful to avoid having spike on your flag configuration storage in case your application is starting multiple instance at the same time.<br/>We ensure a deviation that is maximum + or - 10% of your polling interval.<br />Default: false |
| `hideBanner`              | boolean                   | `false`     | Should we display the beautiful **go-feature-flag** banner when starting the relay proxy                                                                                                                                                                                                                                               |
| `enableSwagger`           | boolean                   | `false`     | Do you want to enable swagger to test the APIs directly. If you are enabling Swagger you will have to provide the `host` configuration and the Swagger UI will be available at `http://<host>:<listen>/swagger/`.                                                                                                                      |
| `host`                    | string                    | `localhost` | This is the DNS you will use to access the relay proxy. This field is used by Swagger to query the API at the right place.                                                                                                                                                                                                             |
| `restApiTimeout`          | int                       | `5000`      | Timeout in millisecond we are accepting to wait in our APIs.                                                                                                                                                                                                                                                                           |
| `debug`                   | boolean                   | `false`     | If `true` you will have more logs in the output that will help you to better understand what happen. If an error happen in the API the error will be also shown in the body.                                                                                                                                                           |
| `fileFormat`              | string                    | `yaml`      | This is the format of your `go-feature-flag` configuration file. Acceptable values are `yaml`, `json`, `toml`.                                                                                                                                                                                                                         |
| `startWithRetrieverError` | boolean                   | `false`     | By default the **relay proxy** will crash if he is not able to retrieve the flags from the configuration.<br/>If you don't want your relay proxy to crash, you can set `startWithRetrieverError` to true. Until the flag is retrievable the relay proxy will only answer with default values.                                          |
| `exporter`                | [exporter](#exporter)     | **none**    | Exporter is the configuration on how to export data.                                                                                                                                                                                                                                                                                   |
| `notifier`                | [notifier](#notifier)     | **none**    | Notifiers is the configuration on where to notify a flag change.                                                                                                                                                                                                                                                                       |
| `apiKeys`                 | []string                  | **none**    | List of authorized API keys. Each request will need to provide one of authorized key inside `Authorization` header with format `Bearer <api-key>`.<br /><br />_Note: there will be no authorization when this config is not set._                                                                                                      |


<a name="retriever"></a>

## type `retriever`

`go-feature-flag` is supporting different kind of retriever such as S3, Google store, etc ...  
In this section we will present all the available retriever configuration available.

### S3

| Field name | Type   | Default  | Description                                                                                                        |
|------------|--------|----------|--------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `bucket`   | string | **none** | **(mandatory)** This is the name of your S3 bucket _(ex: `my-featureflag-bucket`)_.                                |
| `item`     | string | **none** | **(mandatory)** Path to the file inside the bucket _(ex: `config/flag/my-flags.yaml`)_.                            |


### GitHub

| Field name       | Type   | Default  | Description                                                                                                                                                                                                                          |
|------------------|--------|----------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`           | string | **none** | **(mandatory)** Value should be **`github`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                               |
| `repositorySlug` | string | **none** | **(mandatory)** The repository slug of the GitHub repository where your file is located _(ex: `thomaspoignant/go-feature-flag`)_.                                                                                                    |
| `path`           | string | **none** | **(mandatory)** Path to the file inside the repository _(ex: `config/flag/my-flags.yaml`)_.                                                                                                                                          |
| `branch`         | string | `main`   | The branch we should check in the repository.                                                                                                                                                                                        |
| `token`          | string | **none** | Github token used to access a private repository, you need the repo permission ([how to create a GitHub token](https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token)). |
| `timeout`        | string | `10000`  | Timeout in millisecond used when calling GitHub.                                                                                                                                                                                     |

### GitLab

| Field name       | Type   | Default              | Description                                                                                                                                                                                              |
|------------------|--------|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`           | string | **none**             | **(mandatory)** Value should be **`gitlab`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                  |
| `repositorySlug` | string | **none**             | **(mandatory)** The repository slug of the Gitlab repository where your file is located _(ex: `thomaspoignant/go-feature-flag`)_.                                                                        |
| `path`           | string | **none**             | **(mandatory)** Path to the file inside the repository _(ex: `config/flag/my-flags.yaml`)_.                                                                                                              |
| `baseUrl`        | string | `https://gitlab.com` | The base URL of your Gitlab instance.                                                                                                                                                                    |
| `branch`         | string | `main`               | The branch we should check in the repository.                                                                                                                                                            |
| `token`          | string | **none**             | Gitlab personal access token used to access a private repository ([Create a personal access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token)). |
| `timeout`        | string | `10000`              | Timeout in millisecond used when calling Gitlab.                                                                                                                                                         |


### File

| Field name | Type   | Default  | Description                                                                                                          |
|------------|--------|----------|----------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`file`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `path`     | string | **none** | **(mandatory)** Path to the file in your local computer _(ex: `/goff/my-flags.yaml`)_.                               |


### HTTP

| Field name | Type                | Default  | Description                                                                                                          |
|------------|---------------------|----------|----------------------------------------------------------------------------------------------------------------------|
| `kind`     | string              | **none** | **(mandatory)** Value should be **`http`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `url`      | string              | **none** | **(mandatory)** Location where to retrieve the file.                                                                 |
| `method`   | string              | `GET`    | The HTTP Method you are using to call the HTTP endpoint.                                                             |
| `body`     | string              | **none** | The HTTP Body you are using to call the HTTP endpoint.                                                               |
| `headers`  | map[string][]string | **none** | The HTTP headers used to call when calling the HTTP endpoint (useful for authorization).                             |
| `timeout`  | string              | `10000`  | Timeout in millisecond used when calling the HTTP endpoint.                                                          |


### Google Storage

| Field name | Type   | Default  | Description                                                                                                                   |
|------------|--------|----------|-------------------------------------------------------------------------------------------------------------------------------|
| `kind`     | string | **none** | **(mandatory)** Value should be **`googleStorage`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `bucket`   | string | **none** | **(mandatory)** This is the name of your Google Storage bucket _(ex: `my-featureflag-bucket`)_.                               |
| `object`   | string | **none** | **(mandatory)** Path to the file inside the bucket _(ex: `config/flag/my-flags.yaml`)_.                                       |


### Kubernetes ConfigMap

_Note that relay proxy is only supporting this while running inside the kubernetes cluster._

| Field name  | Type   | Default  | Description                                                                                                               |
|-------------|--------|----------|---------------------------------------------------------------------------------------------------------------------------|
| `kind`      | string | **none** | **(mandatory)** Value should be **`configmap`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `namespace` | string | **none** | **(mandatory)** This is the name of the namespace where your **configmap** is located _(ex: `default`)_.                  |
| `configmap` | string | **none** | **(mandatory)** Name of the **configmap** we should read  _(ex: `feature-flag`)_.                                         |
| `key`       | string | **none** | **(mandatory)** Name of the `key` in the **configmap** which contains the flag.                                           |

<a name="exporter"></a>

## type `exporter`

### Webhook

| Field name         | Type                | Default  | Description                                                                                                                                                                                                                 |
|--------------------|---------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string              | **none** | **(mandatory)** Value should be **`webhook`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                    |
| `endpointUrl`      | string              | **none** | **(mandatory)** EndpointURL of your webhook.                                                                                                                                                                                |
| `flushInterval`    | int                 | `60000`  | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                               |
| `maxEventInMemory` | int                 | `100000` | If we hit that limit we will call the webhook.                                                                                                                                                                              |
| `secret`           | string              | **none** | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br/> See [signature section](https://thomaspoignant.github.io/go-feature-flag/latest/data_collection/webhook/#signature) for more details. |
| `meta`             | map[string]string   | **none** | Add all the information you want to see in your request.                                                                                                                                                                    |
| `headers`          | map[string][]string | **none** | Add all the headers you want to add while calling the endpoint                                                                                                                                                              |
	

### File

| Field name         | Type   | Default                                                                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                             |
|--------------------|--------|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string | **none**                                                                                                              | **(mandatory)** Value should be **`file`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                                                                                                                                                                                                    |
| `outputDir`        | string | **none**                                                                                                              | **(mandatory)** OutputDir is the location of the directory where to store the exported files. It should finish with a `/`.                                                                                                                                                                                                                                                                              |
| `flushInterval`    | int    | `60000`                                                                                                               | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                           |
| `maxEventInMemory` | int    | `100000`                                                                                                              | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                          |
| `format`           | string | `JSON`                                                                                                                | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                                            |
| `filename`         | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                          | You can use a templated config to define the name of your exported files. Available replacement are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                               |
| `csvTemplate`      | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using another format than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available. |`
| `parquetCompressionCodec` | string | `SNAPPY` | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md) |`


### Log

| Field name         | Type   | Default                                                                             | Description                                                                                                                                                                                                                                                 |
|--------------------|--------|-------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string | **none**                                                                            | **(mandatory)** Value should be **`log`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                                                         |
| `flushInterval`    | int    | `60000`                                                                             | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                               |
| `maxEventInMemory` | int    | `100000`                                                                            | If we hit that limit we will call the webhook.                                                                                                                                                                                                              |
| `logFormat`        | string | `[{{ .FormattedDate}}] user="{{ .UserKey}}", flag="{{ .Key}}", value="{{ .Value}}"` | LogFormat is the [template](https://golang.org/pkg/text/template/) configuration of the output format of your log.<br/>You can use all the key from the exporter.FeatureEvent + a key called FormattedDate that represent the date with the RFC 3339 Format. |

### S3

| Field name         | Type   | Default                                                                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                             |
|--------------------|--------|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string | **none**                                                                                                              | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                                                                                                                                                                                                      |
| `bucket`           | string | **none**                                                                                                              | **(mandatory)** Name of your S3 Bucket.                                                                                                                                                                                                                                                                                                                                                                 |
| `flushInterval`    | int    | `60000`                                                                                                               | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                           |
| `maxEventInMemory` | int    | `100000`                                                                                                              | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                          |
| `format`           | string | `JSON`                                                                                                                | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                                            |
| `filename`         | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                          | You can use a templated config to define the name of your exported files. Available replacement are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                               |
| `csvTemplate`      | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using another format than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available. |`
| `path`             | string | **bucket root level**                                                                                                 | The location of the directory in S3.                                                                                                                                                                                                                                                                                                                                                                    |
| `parquetCompressionCodec` | string | `SNAPPY` | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md) |`

### Google Storage

| Field name         | Type   | Default                                                                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                             |
|--------------------|--------|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`             | string | **none**                                                                                                              | **(mandatory)** Value should be **`s3`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                                                                                                                                                                                                      |
| `bucket`           | string | **none**                                                                                                              | **(mandatory)** Name of your Google Cloud Storage Bucket.                                                                                                                                                                                                                                                                                                                                               |
| `flushInterval`    | int    | `60000`                                                                                                               | The interval in millisecond between 2 calls to the webhook _(if the `maxEventInMemory` is reached before the flushInterval we will call the webhook before)_.                                                                                                                                                                                                                                           |
| `maxEventInMemory` | int    | `100000`                                                                                                              | If we hit that limit we will call the webhook.                                                                                                                                                                                                                                                                                                                                                          |
| `format`           | string | `JSON`                                                                                                                | Format is the output format you want in your exported file. Available format: `JSON`, `CSV`, `Parquet`.                                                                                                                                                                                                                                                                                                            |
| `filename`         | string | `flag-variation-{{ .Hostname}}-{{ .Timestamp}}.{{ .Format}}`                                                          | You can use a templated config to define the name of your exported files. Available replacement are `{{ .Hostname}}`, `{{ .Timestamp}}` and `{{ .Format}`                                                                                                                                                                                                                                               |
| `csvTemplate`      | string | `{{ .Kind}};{{ .ContextKind}};{{ .UserKey}};{{ .CreationDate}};{{ .Key}};{{ .Variation}};{{ .Value}};{{ .Default}};{{ .Source}}\n` | CsvTemplate is used if your output format is CSV.<br/>This field will be ignored if you are using another format than CSV.<br/>You can decide which fields you want in your CSV line with a go-template syntax, please check [`internal/exporter/feature_event.go`](https://github.com/thomaspoignant/go-feature-flag/blob/main/internal/exporter/feature_event.go) to see what are the fields available. |`
| `path`             | string | **bucket root level**                                                                                                 | The location of the directory in S3.                                                                                                                                                                                                                                                                                                                                                                    |
| `parquetCompressionCodec` | string | `SNAPPY` | ParquetCompressionCodec is the parquet compression codec for better space efficiency. [Available options](https://github.com/apache/parquet-format/blob/master/Compression.md) |`

### SQS

| Field name                | Type   | Default                                                                                                               | Description                                                                                                                                                                                                                                                                                                                                                                                               |
|---------------------------|--------|-----------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`                    | string | **none**                                                                                                              | **(mandatory)** Value should be **`sqs`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                                                                                                                                                                                                      |
| `queueUrl`                | string | **none**                                                                                                              | **(mandatory)** URL of your SQS queue.<br/>_You can find it in your AWS console._                                                                                                                                                                                                                                                                                                                         |

<a name="notifier"></a>

## type `notifier`

### Slack

| Field name        | Type   | Default  | Description                                                                                                           |
|-------------------|--------|----------|-----------------------------------------------------------------------------------------------------------------------|
| `kind`            | string | **none** | **(mandatory)** Value should be **`slack`**.<br/>_This field is mandatory and describe which retriever you are using._ |
| `slackWebhookUrl` | string | **none** | **(mandatory)** The complete URL of your incoming webhook configured in Slack.                                        |

### Webhook

| Field name    | Type                | Default    | Description                                                                                                                                                                                                                   |
|---------------|---------------------|------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `kind`        | string              | **none**   | **(mandatory)** Value should be **`webhook`**.<br/>_This field is mandatory and describe which retriever you are using._                                                                                                        |
| `endpointUrl` | string              | **none**   | **(mandatory)** The complete URL of your API (we will send a POST request to this URL, see [format](https://thomaspoignant.github.io/go-feature-flag/latest/notifier/webhook/#format)                                         |
| `secret`      | string              | **none**   | Secret used to sign your request body and fill the `X-Hub-Signature-256` header.<br/> See [signature section](https://thomaspoignant.github.io/go-feature-flag/latest/data_collection/webhook/#signature) for more details.   |
| `meta`        | map[string]string   | **none**   | Add all the information you want to see in your request.                                                                                                                                                                      |
| `headers`     | map[string][]string | **none**   | Add all the headers you want to add while calling the endpoint                                                                                                                                                                |
