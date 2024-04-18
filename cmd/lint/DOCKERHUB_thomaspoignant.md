# GO Feature Flag Lint cli
<span style="color: red">🚨 Attention: The GO Feature Flag has now transitioned to its own organization. We recommend updating your configurations to use [`gofeatureflag/go-feature-flag-lint`](https://hub.docker.com/r/gofeatureflag/go-feature-flag-lint:). We will continue to provide support for the original organization for a certain period of time.</span>

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
