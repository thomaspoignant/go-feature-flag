# GO Feature Flag Lint cli

![Status](https://img.shields.io/badge/status-deprecated-red)

> [!WARNING]  
> The linter command has been deprecated in favor of the GO Feature Flag Command Line.
> Check the documentation for the [GO Feature Flag Command Line](http://gofeatureflag.org/docs/tooling/linter) for more information.

The lint command lint tool validates that a flags file can be parsed by GO Feature Flag.

## How to use this image

```shell
docker run \
  -v $(pwd)/your/configuration_folder:/config \
  gofeatureflag/go-feature-flag-lint:latest \
  --input-file=/config/my-go-feature-flag-config.yaml \
  --input-format=yaml
```

### Params description

The command line has 2 parameters:

| param            | description                                                                                                       |
|------------------|-------------------------------------------------------------------------------------------------------------------|
| `--input-file`   | **(mandatory)** The location of your configuration file.                                                          |
| `--input-format` | **(mandatory)** The format of your current configuration file. <br/>Available formats are `yaml`, `json`, `toml`. |
