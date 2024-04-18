# GO Feature Flag Migration cli

<span style="color: red">🚨 Attention: The GO Feature Flag has now transitioned to its own organization. We recommend updating your configurations to use [`gofeatureflag/go-feature-flag-migration-cli`](https://hub.docker.com/r/gofeatureflag/go-feature-flag-migration-cli:). We will continue to provide support for the original organization for a certain period of time.</span>

The migration command line purpose is to migrate your configuration file from the flag format in version `v0.x.x` to the format `v1.x.x`.

We changed the format to extend the capabilities of **GO Feature Flag**, despite that the old format will still be
supported, it is better to update your configuration file.

## How to use this image

```shell
docker run \
  -v $(pwd)/your/configuration_folder:/config \
  gofeatureflag/go-feature-flag-migration-cli:latest \
  --input-format=yaml \
  --input-file=/config/my-go-feature-flag-config-v0.x.x.yaml \
  --output-format=yaml \
  --output-file=/config/my-go-feature-flag-config-v1.x.x.yaml
```

### Params description

The command line has 4 parameters:

| param             | description                                                                                                                                                      |
|-------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--input-file`    | **(mandatory)** The location of your configuration file in version `v0.x.x`.                                                                                     |
| `--input-format`  | **(mandatory)** The format of your current configuration file.<br/>Available formats are `yaml`, `json`, `toml`.                                                 |
| `--output-file`   | _(optional)_ The location where the new configuration file will be stored in version `v1.x.x`.<br/>_If omitted the configuration will be output in the console._ |
| `--output-format` | _(optional)_ The target format of the configuration.<br/>Available formats are `yaml`, `json`, `toml`.   <br/>_If omitted `yaml` will be used._                  |
