# GO Feature Flag Migration cli

The migration command line purpose is to migrate your configuration file from the flag format in version `v0.x.x` to the format `v1.x.x`.

We changed the format to extend the capabilities of **GO Feature Flag**, despite that the old format will still be
supported, it is better to update your configuration file.

## How to install the cli

### Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-migration-cli
```

### Install using Scoop (windows)
```shell
scoop bucket add org https://github.com/thomaspoignant/scoop.git
scoop install go-feature-flag-migration-cli
```

### Install using docker
```shell
docker pull thomaspoignant/go-feature-flag-migration-cli
```
More information about the usage of the container in the [dockerhub page](https://hub.docker.com/r/thomaspoignant/go-feature-flag-migration-cli).


## How to use the cli

```shell
# example:
go-feature-flag-migration-cli --input-format=yaml --input-file=/input/my-go-feature-flag-config-v0.x.x.yaml --output-format=yaml --output-file=input/my-go-feature-flag-config-v1.x.x.yaml
```

The command line has 4 parameters:

| param             | description                                                                                                                                                      |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--input-file`    | **(mandatory)** The location of your configuration file in version `v0.x.x`.                                                                                     |
| `--input-format`  | **(mandatory)** The format of your current configuration file.<br/>Available formats are `yaml`, `json`, `toml`.                                                 |
| `--output-file`   | _(optional)_ The location where the new configuration file will be stored in version `v1.x.x`.<br/>_If omitted the configuration will be output in the console._ |
| `--output-format` | _(optional)_ The target format of the configuration.<br/>Available formats are `yaml`, `json`, `toml`.   <br/>_If omitted `yaml` will be used._                  |
