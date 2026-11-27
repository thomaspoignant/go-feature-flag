# 💻 Command Line

The GO Feature Flag Command Line is a CLI tool to interact with GO Feature Flag in your terminal.

It offers a variety of commands to interact with GO Feature Flag, for now the supported commands are:

* [`evaluate`](https://gofeatureflag.org/docs/tooling/evaluate.md) to evaluate feature flags directly in your terminal
* [`lint`](https://gofeatureflag.org/docs/tooling/linter.md) to validate a configuration file format.
* [`generate`](https://gofeatureflag.org/docs/tooling/generate.md) to generate the OpenFeature flag manifest for OpenFeature cli **(experimental)**.

## Install the GO Feature Flag Command Line[​](#install-the-go-feature-flag-command-line "Direct link to Install the GO Feature Flag Command Line")

### Install using Homebrew (mac and linux)[​](#install-using-homebrew-mac-and-linux "Direct link to Install using Homebrew (mac and linux)")

```
brew tap thomaspoignant/homebrew-tap

brew install go-feature-flag-cli
```

### Install using Scoop (windows)[​](#install-using-scoop-windows "Direct link to Install using Scoop (windows)")

```
scoop bucket add org https://github.com/go-feature-flag/scoop.git

scoop install go-feature-flag-cli
```

### Install using Docker[​](#install-using-docker "Direct link to Install using Docker")

```
docker pull gofeatureflag/go-feature-flag-cli:latest



# the same image is also available on the GitHub Container Registry

# (for the releases after v1.55.1)

docker pull ghcr.io/go-feature-flag/go-feature-flag-cli:latest
```
