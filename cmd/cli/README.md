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
More information about the usage of the container in the [dockerhub page](https://hub.docker.com/r/gofeatureflag/go-feature-flag-cli).

# How to use the command line

**`go-feature-flag-cli`**  is a command line tool.

## How to evaluate a flag
```shell
go-feature-flag-cli evaluate --config="<location_of_your_flag_configuration_file>" --flag="<name_of_your_flag_to_evaluate>" --ctx='<evaluation_ctx_as_json_string>'
```

## How to lint a configuration file
```shell
go-feature-flag-cli lint <location_of_your_flag_configuration_file> --format="<yaml or json or toml>"
```

# License

View [license](https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE) information for the software contained in this image.

## How can I contribute?
This project is open for contribution, see the [contributor's guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.
