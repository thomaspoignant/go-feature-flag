# GO Feature Flag Relay Proxy

<p align="center">
  <img width="250" height="238" src="https://github.com/thomaspoignant/go-feature-flag/raw/main/logo.png" alt="go-feature-flag logo" />
</p>

<p align="center">
  <img alt="Docker Image Version" src="https://img.shields.io/docker/v/thomaspoignant/go-feature-flag-relay-proxy?sort=semver&color=green"/>
  <img alt="Docker Image Size" src="https://img.shields.io/docker/image-size/thomaspoignant/go-feature-flag-relay-proxy?sort=semver"/>
  <img alt="Docker Hub downloads" src="https://img.shields.io/docker/pulls/thomaspoignant/go-feature-flag-relay-proxy?logo=docker"/>
  <a href="https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag" alt="License"/></a>
  <a href="https://gophers.slack.com/messages/go-feature-flag"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a> 
</p>


--- 

# What is GO Feature Flag Relay Proxy?

The GO Feature Flag Relay Proxy retrieve your feature flag configuration file using [`thomaspoignant/go-feature-flag`](https://github.com/thomaspoignant/go-feature-flag) SDK and expose APIs to get your flags variations.  
It lets a number of servers to connect to a single configuration file.

This can be useful if you want to use the same feature flags configuration file for frontend and backend, this allows to be language agnostic by using standard protocol.

For more information about GO Feature Flag Relay Proxy, please visit [github.com/thomaspoignant/go-feature-flag](https://github.com/thomaspoignant/go-feature-flag/tree/main/cmd/relayproxy).


# Quick reference

- This default distribution is the official distribution for `go-feature-flag-relay-proxy`.

- Where to file issues: 
  [https://github.com/thomaspoignant/go-feature-flag/issues/](https://github.com/thomaspoignant/go-feature-flag/issues/new?assignees=&labels=bug%2C+docker%2C+needs-triage&template=bug.md&title=(bug%20docker)).

- Source are available in [`go-feature-flag-relay-proxy` repo](https://github.com/thomaspoignant/go-feature-flag-relay-proxy).

- All versions are available in the [tags](https://hub.docker.com/r/thomaspoignant/go-feature-flag-relay-proxy/tags).

- Release notes are available [here](https://github.com/thomaspoignant/go-feature-flag-relay-proxy/releases).


# How to use this image

**`go-feature-flag-relay-proxy`** requires a configuration file to be used.

By default, we expect to have this configuration file in the `/goff` directory of the container and the file should be named `goff-proxy.yaml`.  

The default port used for the service is `1031`.

```shell
docker run \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag-relay-proxy:latest
```

## Test it locally

This is a small example on how to run `go-feature-flag-relay-proxy` locally.

```shell
# Download an example of a basic configuration file.
curl https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/cmd/relayproxy/testdata/config/valid-file.yaml -o goff-proxy.yaml

# Launch the container
docker run \
  -p 1031:1031 \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag-relay-proxy:latest
  
# Call the API
curl -X 'POST' \
  'http://localhost:1031/v1/feature/flag-only-for-admin/eval'  -H 'accept: application/json'  -H 'Content-Type: application/json' \
  -d '{ "user": { "key": "contact@gofeatureflag.org", "anonymous": true, "custom": { "admin": true, "email": "contact@gofeatureflag.org" }}, "defaultValue": "false"}'
```

# License

View [license](https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE) information for the software contained in this image.

## How can I contribute?
This project is open for contribution, see the [contributor's guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.
