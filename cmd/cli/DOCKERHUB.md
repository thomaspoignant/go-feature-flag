# GO Feature Flag Command Line

<p align="center">
  <img width="250" height="238" src="https://github.com/thomaspoignant/go-feature-flag/raw/main/logo.png" alt="go-feature-flag logo" />
</p>

<p align="center">
  <img alt="Docker Image Version" src="https://img.shields.io/docker/v/gofeatureflag/go-feature-flag-cli?sort=semver&color=green"/>
  <img alt="Docker Image Size" src="https://img.shields.io/docker/image-size/gofeatureflag/go-feature-flag-cli?sort=semver"/>
  <img alt="Docker Hub downloads" src="https://img.shields.io/docker/pulls/gofeatureflag/go-feature-flag-cli?logo=docker"/>
  <a href="https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
  <a href="https://gofeatureflag.org/slack"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a> 
</p>


--- 

# What is GO Feature Flag Command Line?

The GO Feature Flag Command Line is a CLI tool to interact with GO Feature Flag in your terminal.  
For now it supports the following commands:
- `evaluate` to evaluate feature flags directly in your terminal
- `lint` to validate a configuration file format.

# Quick reference

- This default distribution is the official distribution for `go-feature-flag-cli`.

- Where to file issues:
  [https://github.com/thomaspoignant/go-feature-flag/issues/](https://github.com/thomaspoignant/go-feature-flag/issues/new?assignees=&labels=bug%2C+relay-proxy%2C+docker%2C+needs-triage&template=bug.md&title=(bug%20docker)).

- Source are available in [`go-feature-flag` repo](https://github.com/thomaspoignant/go-feature-flag/tree/main/cmd/cli).

- All versions are available in the [tags](https://hub.docker.com/r/gofeatureflag/go-feature-flag-cli/tags).

- Release notes are available [here](https://github.com/thomaspoignant/go-feature-flag/releases).


# How to use this image

**`go-feature-flag-cli`**  is a command line tool.

## How to evaluate a flag
```shell
docker run -v <location_of_your_local_file>:/config.yaml \
  gofeatureflag/go-feature-flag-cli evaluate \
  --config=/config.yaml \
  --flag="<name_of_your_flag_to_evaluate>" \
  --ctx='<evaluation_ctx_as_json_string>'
```

## How to lint a configuration file
```shell
docker run -v <location_of_your_local_file>:/config.yaml \
  gofeatureflag/go-feature-flag-cli lint \
  --config=/config.yaml \
  --format="yaml"
```

# Supported tags and respective `Dockerfile` links
GO Feature Flag Command lind is publishing the following tags:

The numbered version _(ex: `v1`, `1.29`, etc ...)_ are using a **`distroless`** base image,
ensuring a minimal image size and high security.

# License

View [license](https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE) information for the software contained in this image.

## How can I contribute?
This project is open for contribution, see the [contributor's guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.
