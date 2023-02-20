---
sidebar_position: 91
description: Lint your config
---

# Lint your config

A faulty configuration could make **GO Feature Flag** not the way you expect.  
This is why we have introduced the `go-feature-flag-lint` a command line tool validates that a flags file can be parsed by **GO Feature Flag**.

:::tip
We recommend you to use this command line in your CI/CD pipelines to avoid any disappointment.
:::

## Install the linter

### Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-lint
```
 
### Install using Scoop (windows)
```shell
scoop bucket add org https://github.com/thomaspoignant/scoop.git
scoop install go-feature-flag-lint
```

### Install using Docker
```shell
docker pull thomaspoignant/go-feature-flag-lint:latest
```

## Use the linter

```shell
./go-feature-flag-lint \
  --input-format=yaml \
  --input-file=/input/my-go-feature-flag-config.yaml
```

The command line has 2 arguments you should specify.

| param            | description                                                                                                       |
|------------------|-------------------------------------------------------------------------------------------------------------------|
| `--input-file`   | **(mandatory)** The location of your configuration file.                                                          |
| `--input-format` | **(mandatory)** The format of your current configuration file. <br/>Available formats are `yaml`, `json`, `toml`. |