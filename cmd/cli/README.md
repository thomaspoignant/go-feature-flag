# GO Feature Flag Command Line

# What is GO Feature Flag Command Line?

The GO Feature Flag Command Line is a CLI tool to interact with GO Feature Flag in your terminal.  
For now it supports the following commands:

- `evaluate` to evaluate feature flags directly in your terminal
- `lint` to validate a configuration file format.

## How to install the cli

### Install using Homebrew (mac and linux)

```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-cli
```

### Install using docker

```shell
docker pull gofeatureflag/go-feature-flag-cli
```

More information about the usage of the container in
the [dockerhub page](https://hub.docker.com/r/gofeatureflag/go-feature-flag-cli).

# How to use the command line

**`go-feature-flag-cli`**  is a command line tool.

## How to evaluate a flag

The evaluate command allows you to evaluate a feature flag or inspect the configuration of your retriever using
`--check-mode`

```shell
go-feature-flag-cli evaluate [OPTIONS]
```

### Key Flags

| Flag           | Shorthand | Description                                                                                                                     | Default |
|----------------|-----------|---------------------------------------------------------------------------------------------------------------------------------|---------|
| `--kind`       | k         | Kind of configuration source. Determines where to read your flags from                                                          | file    |
| `--config`     | c         | Path to the local flag configuration file (⚠️ deprecated, use `--path` instead)                                                 | ""      |
| `--format`     | f         | Format of your input file (YAML, JSON or TOML)                                                                                  | yaml    |
| `--flag`       |           | Name of the flag to evaluate                                                                                                    | ""      |
| `--timeout`    |           | Timeout in seconds to access your configuration file                                                                            | 0       |
| `--ctx`        |           | Evaluation context as a json string                                                                                             | {}      |
| `--check-mode` |           | Check only mode - when set, the command will not perform any <br/>evaluation and returns the configuration of spanned retriever | false   |

Supported values for `--kind` are:

- `file`
- `http`
- `github`
- `gitlab`
- `s3`
- `googleStorage`
- `configmap` (kubernetes)
- `mongodb`
- `bitbucket`
- `azureBlobStorage`
- `postgresql`

**Caution**: We do not support `redis` retriever as for now due
to: https://github.com/thomaspoignant/go-feature-flag/issues/4023.

### Retriever specific flags

The aforementioned `--kind` parameter is used to determine the retriever to use. The semantic meaning of other flags
depends on that one, for
example `--path` parameter is used to specify local file when `--kind` being `file` but when `--kind` being `github` it
is used to specify
the path to the remote file.

#### File

| Flag     | Description                               | Default |
|----------|-------------------------------------------|---------|
| `--path` | Path to the local flag configuration file | ""      |

#### HTTP

| Flag       | Description                                                         | Default |
|------------|---------------------------------------------------------------------|---------|
| `--url`    | URL of the remote flag configuration file                           | ""      |
| `--method` | HTTP method to access your configuration file on HTTP               | GET     |
| `--body`   | Http body to access your configuration file on HTTP                 | ""      |
| `--header` | Header to add to the request. Supported formats are `k:v` and `k=v` | ""      |

#### GitHub

| Flag                | Description                                                                                                       | Default |
|---------------------|-------------------------------------------------------------------------------------------------------------------|---------|
| `--repository-slug` | Name of the repository                                                                                            | ""      |
| `--branch`          | Git branch name                                                                                                   | ""      |
| `--auth-token`      | Authentication token to access your configuration file                                                            | ""      |
| `--github-token`    | Authentication token to access your configuration file on GitHub<br/> (⚠️ deprecated, use `--auth-token` instead) | ""      |
| `--path`            | Path to the remote flag configuration file inside github repository                                               | ""      |

#### GitLab

| Flag                | Description                                                         | Default |
|---------------------|---------------------------------------------------------------------|---------|
| `--base-url`        | Base URL of your configuration file on Gitlab                       | ""      |
| `--repository-slug` | Name of the repository                                              | ""      |
| `--branch`          | Git branch name                                                     | ""      |
| `--path`            | Path to the remote flag configuration file inside gitlab repository | ""      |

#### BitBucket

| Flag                | Description                                                            | Default |
|---------------------|------------------------------------------------------------------------|---------|
| `--base-url`        | Base URL of your configuration file on BitBucket                       | ""      |
| `--repository-slug` | Name of the repository                                                 | ""      |
| `--branch`          | Git branch name                                                        | ""      |
| `--path`            | Path to the remote flag configuration file inside bitbucket repository | ""      |

#### S3

| Flag       | Description        | Default |
|------------|--------------------|---------|
| `--bucket` | Name of the bucket | ""      |
| `--item`   | Item of the bucket | ""      |

#### Google Storage

| Flag       | Description          | Default |
|------------|----------------------|---------|
| `--bucket` | Name of the bucket   | ""      |
| `--object` | Object of the bucket | ""      |

#### ConfigMap (Kubernetes)

| Flag           | Description                | Default   |
|----------------|----------------------------|-----------|
| `--namespace`  | Namespace of the ConfigMap | "default" |
| `--config-map` | Name of the ConfigMap      | ""        |
| `--key`        | Key of the ConfigMap       | ""        |

#### MongoDB

| Flag           | Description                                         | Default |
|----------------|-----------------------------------------------------|---------|
| `--uri`        | URI of your configuration file                      | ""      |
| `--database`   | Database name of your configuration file on mongodb | ""      |
| `--collection` | Collection of your configuration file on mongodb    | ""      |

#### Azure Blob Storage

| Flag             | Description                 | Default |
|------------------|-----------------------------|---------|
| `--account-name` | Name of the storage account | ""      |
| `--account-key`  | Key of the storage account  | ""      |
| `--container`    | Name of the container       | ""      |
| `--object`       | Name of the object blob     | ""      |

#### PostgreSQL

| Flag       | Description                                        | Default |
|------------|----------------------------------------------------|---------|
| `--uri`    | URI of your configuration file                     | ""      |
| `--table`  | Table of your configuration file                   | ""      |
| `--column` | Column mapping to add. Supported format is `c1:c2` | ""      |

As mentioned above the `--config` flag is deprecated and we encourage you to use the `--path` flag instead. For example
the following command:

```shell
go-feature-flag-cli evaluate --config="<location_of_your_flag_configuration_file>" --flag="<name_of_your_flag_to_evaluate>" --ctx='<evaluation_ctx_as_json_string>'
```

may be replaced by:

```shell
go-feature-flag-cli evaluate --kind="file" --path="<location_of_your_flag_configuration_file>" --flag="<name_of_your_flag_to_evaluate>" --ctx='<evaluation_ctx_as_json_string>'
```

## How to lint a configuration file

```shell
go-feature-flag-cli lint <location_of_your_flag_configuration_file> --format="<yaml or json or toml>"
```

# License

View [license](https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE) information for the software
contained in this image.

## How can I contribute?

This project is open for contribution, see
the [contributor's guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful
tips.
