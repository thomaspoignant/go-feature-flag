# GO Feature Flag Server

<p align="center">
  <img width="250" height="238" src="https://github.com/thomaspoignant/go-feature-flag/raw/main/logo.png" alt="go-feature-flag logo" />
</p>

<p align="center">
  <img alt="Docker Image Version" src="https://img.shields.io/docker/v/thomaspoignant/go-feature-flag?sort=semver&color=green"/>
  <img alt="Docker Image Size" src="https://img.shields.io/docker/image-size/thomaspoignant/go-feature-flag?sort=semver"/>
  <img alt="Docker Hub downloads" src="https://img.shields.io/docker/pulls/thomaspoignant/go-feature-flag?logo=docker"/>
  <a href="https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
  <a href="https://gophers.slack.com/messages/go-feature-flag"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a> 
</p>


--- 

# What is GO Feature Flag?

GO Feature Flag is a lightweight and open-source solution that provides a simple and complete feature flag implementation.

The solution has been built to facilitate the usage of feature flags in your code without having to contact any vendor.

## What can I do with GO Feature Flag?

- Storing your configuration flags file on various locations (HTTP, S3, Kubernetes, see full list).
- Configuring your flags in various format (JSON, TOML and YAML).
- Adding complex rules to target your users.
- Use a complex rollout strategy for your flags :
  - Run A/B testing experimentation.
  - Progressively rollout a feature.
  - Schedule your flag updates.
- Exporting your flags usage data to various destinations such as (S3, Google cloud storage, file, see the full list).
- Getting notified when a flag has been changed (webhook and slack).
- Use GO Feature Flag in several languages with Open Feature SDKs.


# Quick reference

- This default distribution is the official distribution for `go-feature-flag`.

- Where to file issues: 
  [https://github.com/thomaspoignant/go-feature-flag/issues/](https://github.com/thomaspoignant/go-feature-flag/issues/new?assignees=&labels=bug%2C+server%2C+docker%2C+needs-triage&template=bug.md&title=(bug%20docker)).

- Source are available in [`go-feature-flag` repo](https://github.com/thomaspoignant/go-feature-flag/tree/main/cmd/server).

- All versions are available in the [tags](https://hub.docker.com/r/thomaspoignant/go-feature-flag/tags).

- Release notes are available [here](https://github.com/thomaspoignant/go-feature-flag/releases).


# How to use this image

**`go-feature-flag`** requires a configuration file to be used.

By default, we expect to have this configuration file in the `/goff` directory of the container and the file should be named `goff-proxy.yaml`.  

The default port used for the service is `1031`.

```shell
docker run \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag:latest
```

## Test it locally

This is a small example on how to run `go-feature-flag` locally.

```shell
# Download an example of a basic configuration file.
curl https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/cmd/relayproxy/testdata/config/valid-file.yaml -o goff-proxy.yaml

# Launch the container
docker run \
  -p 1031:1031 \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag:latest
  
# Call the API
curl -X 'POST' \
  'http://localhost:1031/v1/feature/flag-only-for-admin/eval'  -H 'accept: application/json'  -H 'Content-Type: application/json' \
  -d '{ "user": { "key": "contact@gofeatureflag.org", "anonymous": true, "custom": { "admin": true, "email": "contact@gofeatureflag.org" }}, "defaultValue": "false"}'
```

# License

View [license](https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE) information for the software contained in this image.

## How can I contribute?
This project is open for contribution, see the [contributor's guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.
