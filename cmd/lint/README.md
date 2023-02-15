# GO Feature Flag Lint cli

The lint command lint tool validates that a flags file can be parsed by GO Feature Flag.

## How to install the cli

:construction:

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