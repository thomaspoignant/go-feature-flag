---
sidebar_position: 90
description: How to migrate from v0.x.x to v1.x.x
---

# 🔄 Migrate from v0.x to v1.x
:::warning
**⚠️ Version `v1.35.0` will be the last version of the cli.**.  
**Why? Because it is feature complete and because it has been decided to stop supporting `v0.x.x` format.**
:::

:::info
Version `v1.0.0` has introduced a new flag format that push the limits of **GO Feature Flag** even further.  
**NOTE:** The flag format from all the versions `v0.x.x` are still compatible and supported by the `v1.0.0`.
:::

A command line is available to help you to convert your actual configuration file to the version `v1.x.x`.


## Install the migration command line

### Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-migration-cli
```
 
### Install using Scoop (windows)
```shell
scoop bucket add org https://github.com/go-feature-flag/scoop.git
scoop install go-feature-flag-migration-cli
```

### Install using Docker
```shell
docker pull gofeatureflag/go-feature-flag-migration-cli:latest
```

## Use the migration command line

```shell
./go-feature-flag-migration-cli \
  --input-format=yaml \
  --input-file=/config/my-go-feature-flag-config-v0.x.x.yaml \
  --output-format=yaml \
  --output-file=/config/my-go-feature-flag-config-v1.x.x.yaml
```

The command line has 4 arguments you should specify.

- `input-format`: Format of your input file (`YAML`, `JSON` or `TOML`).
- `input-file`: Location of the flag file you want to convert.
- `output-format`: Format of your output file (`YAML`, `JSON` or `TOML`).
- `output-file`: Location of the converted flag file.


## Update your flag file
When your file is ready, you just have to replace your file in the location where GO Feature Flag is retrieves it.

:::tip
If for any reason your file is not readable by GO Feature Flag, it will not break anything, we will keep the latest version we have in memory. 
:::
