# GO Feature Flag Lint cli

![Status](https://img.shields.io/badge/status-deprecated-red)

> [!WARNING]  
> The linter command has been deprecated in favor of the GO Feature Flag Command Line.
> Check the documentation for the [GO Feature Flag Command Line](http://gofeatureflag.org/docs/tooling/linter) for more information.

The lint command line tool validates that a flags file can be parsed by GO Feature Flag.

## How to install the cli

### Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-lint
```

### Install using docker
```shell
docker pull gofeatureflag/go-feature-flag-lint
```
More information about the usage of the container in the [dockerhub page](https://hub.docker.com/r/gofeatureflag/go-feature-flag-lint).

## How to use the cli

```shell
# example:
go-feature-flag-lint --input-format=yaml --input-file=/input/my-go-feature-flag-config.yaml
```

The command line has 2 parameters:

| param            | description                                                                                                       |
|------------------|-------------------------------------------------------------------------------------------------------------------|
| `--input-file`   | **(mandatory)** The location of your configuration file.                                                          |
| `--input-format` | **(mandatory)** The format of your current configuration file. <br/>Available formats are `yaml`, `json`, `toml`. |
