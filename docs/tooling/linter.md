# 🚧 Linter

Ensuring the accuracy of your configuration is vital for the **GO Feature Flag** to function as expected. This is why you can lint your configuration file using the [`go-feature-flag-cli`](https://gofeatureflag.org/docs/tooling/cli.md), a command line tool that allow to interact with GO Feature Flag in your terminal. The `lint` command validates whether a flag file can be parsed by **GO Feature Flag**.

tip

We recommend you to use this command line in your CI/CD pipelines to avoid any unforeseen issues.

## Install the Command Line[​](#install-the-command-line "Direct link to Install the Command Line")

Check the [installation guide](https://gofeatureflag.org/docs/tooling/cli.md) to install the `go-feature-flag-cli`.

## Use the linter[​](#use-the-linter "Direct link to Use the linter")

```
./go-feature-flag-cli lint /input/my-go-feature-flag-config.goff.yaml --format=yaml
```

You have to pass the location of your configuration file and the format of your current configuration file *(available formats are `yaml`, `json`, `toml`)*.

## Use the linter in your CI (continuous integration)[​](#use-the-linter-in-your-ci-continuous-integration "Direct link to Use the linter in your CI (continuous integration)")

You can run `go-feature-flag-cli` directly in your CI:

* Github Action
* CircleCi
* Gitlab

```
name: "Build"

on:

  push:

    branches:

      - main

  pull_request:

    types: [ opened, synchronize, reopened ]



jobs:

  lint:

    runs-on: ubuntu-latest

    steps:

      - uses: actions/checkout@v4

      - name: Lint the config file

        uses: go-feature-flag/gofeatureflag-lint-action@v1

        with:

          flag-file: ./path/to/your/config.yaml

          format: yaml
```

```
version: 2.1

jobs:

  build:

    docker:

     - image: cimg/base:2022.05



    steps:

      - checkout

      - run: curl -L $(curl -s https://api.github.com/repos/thomaspoignant/go-feature-flag/releases/latest | jq -r '.assets[] | select(.name|match("Linux_x86_64.tar.gz$")) | .browser_download_url' | grep 'go-feature-flag-cli') --output release.tar.gz && tar -zxvf release.tar.gz

      - run: ./go-feature-flag-cli lint flag-config.goff.yaml --format=yaml # please put the right file name
```

```
image: ubuntu

lint-job:

  stage: build



  before_script:

    - apt-get -qq update

    - apt-get install -y jq curl



  script:

    - curl -L $(curl -s https://api.github.com/repos/thomaspoignant/go-feature-flag/releases/latest | jq -r '.assets[] | select(.name|match("Linux_x86_64.tar.gz$")) | .browser_download_url' | grep 'go-feature-flag-cli') --output release.tar.gz && tar -zxvf release.tar.gz

    - ./go-feature-flag-cli lint flag-config.goff.yaml --format=yaml # please put the right file name
```

## Lint Feature Flags Programmatically[​](#lint-feature-flags-programmatically "Direct link to Lint Feature Flags Programmatically")

You can also lint/validate a GO Feature Flag configuration programmatically without using the CLI. Check the [example repository](https://github.com/thomaspoignant/go-feature-flag/tree/main/examples/lint_flag_programmatically/README.md) to see how to do it in Go.
