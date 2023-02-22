# GO Feature Flag Lint cli

The lint command line tool validates that a flags file can be parsed by GO Feature Flag.

## How to install the cli

### Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-lint
```

### Install using docker
```shell
docker pull thomaspoignant/go-feature-flag-lint
```
More information about the usage of the container in the [dockerhub page](https://hub.docker.com/r/thomaspoignant/go-feature-flag-lint-cli).

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