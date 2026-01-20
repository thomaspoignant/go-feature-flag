---
sidebar_position: 30
title: 🎯 Evaluate CLI
description: Evaluate feature flags directly in your terminal
---

# 🎯 Evaluate feature flags in your terminal

Sometimes for debug or testing purposes, you may want to be able to know what will be the variant used during the evaluation of your feature flag.
With the GO Feature Flag Command Line, you can evaluate a feature flag directly in your terminal.

:::tip
You can also use the `evaluate` command to use feature flags in your CI/CD pipelines.
:::

## Install the Command Line

Check the [installation guide](./cli) to install the `go-feature-flag-cli`.

## Use the evaluate command in your terminal

```shell
./go-feature-flag-cli evaluate \
  --config="<location_of_your_flag_configuration_file>" \
  --format="yaml" \
  --flag="<name_of_your_flag_to_evaluate>" \ 
  --ctx='<evaluation_ctx_as_json_string>'
```

| param      | description                                                                                                            |
|------------|------------------------------------------------------------------------------------------------------------------------|
| `--config` | **(mandatory)** The location of your configuration file.                                                               |
| `--ctx`    | **(mandatory)** The evaluation context used to evaluate the flag in json format (ex: `{"targetingKey":"123"}`).        |
| `--format` | The format of your configuration flag _(acceptable values:`yaml`, `json`, `toml`)_.<br/>Default: **`yaml`**            |
| `--flag`   | The name of the flag you want to evaluate, if omitted all flags will be evaluated                                      |

