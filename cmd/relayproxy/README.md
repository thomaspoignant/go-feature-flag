# GO feature flag relay proxy
<p align="center">
  <img width="250" height="238" src="../../logo.png" alt="go-feature-flag logo" />
</p>

<p align="center">
    <a href="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml"><img src="https://github.com/thomaspoignant/go-feature-flag/actions/workflows/ci.yml/badge.svg" alt="Build Status" /></a>
    <a href="https://coveralls.io/github/thomaspoignant/go-feature-flag"><img src="https://coveralls.io/repos/github/thomaspoignant/go-feature-flag/badge.svg" alt="Coverage Status" /></a>
    <a href="https://sonarcloud.io/dashboard?id=thomaspoignant_go-feature-flag-relay-proxy"><img src="https://sonarcloud.io/api/project_badges/measure?project=thomaspoignant_go-feature-flag-relay-proxy&metric=alert_status" alt="Sonarcloud Status" /></a>
    <a href="https://github.com/thomaspoignant/go-feature-flag-relay-proxy/releases"><img src="https://img.shields.io/github/v/release/thomaspoignant/go-feature-flag-relay-proxy" alt="Release version" /></a>
    <a href="https://pkg.go.dev/github.com/thomaspoignant/go-feature-flag-relay-proxy"><img src="https://godoc.org/github.com/thomaspoignant/go-feature-flag-relay-proxy?status.svg" alt="GoDoc" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/thomaspoignant/go-feature-flag-relay-proxy?logo=go%20version" alt="Go version"/>
    <a href="https://github.com/thomaspoignant/go-feature-flag/blob/main/LICENSE"><img src="https://img.shields.io/github/license/thomaspoignant/go-feature-flag-relay-proxy" alt="License"/></a>
    <a href="https://gophers.slack.com/messages/go-feature-flag"><img src="https://img.shields.io/badge/join-us%20on%20slack-gray.svg?longCache=true&logo=slack&colorB=green" alt="Join us on slack"></a> 
</p>

## What does the GO Feature Flag Relay Proxy do?
The GO Feature Flag Relay Proxy retrieve your feature flag configuration file using [`thomaspoignant/go-feature-flag`](https://github.com/thomaspoignant/go-feature-flag) SDK and expose APIs to get your flags variations.  
It lets a number of servers to connect to a single configuration file.

This can be useful if you want to use the same feature flags configuration file for frontend and backend, this allows to be language agnostic by using standard protocol.


## When should I use the GO Feature Flag Relay Proxy?
- If you want to access your feature flag using an API instead of the [`thomaspoignant/go-feature-flag`](https://github.com/thomaspoignant/go-feature-flag) SDK.
- If you are not using GOlang to build your application.
- If you want to reduce the number of accesses to your configuration flag by centralizing them.
<!-- - If you are using any SDKs that connect to the Relay Proxy. -->

## Installation
### Install using Homebrew (mac and linux)
```shell
brew install go-feature-flag-relay-proxy
```

### Install using Scoop (windows)
```shell
scoop bucket add org https://github.com/thomaspoignant/scoop.git
scoop install go-feature-flag-relay-proxy
```

## Getting started

Before starting your **relay proxy** you will need to create a minimal configuration file.  

```yaml
# this is a minimal config containing only where your flag file is located 
retriever:
  kind: http
  url: https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/retriever_file/flags.yaml
```

After that you can launch the **relay proxy** by using this command:
```shell
go-feature-flag-relay-proxy --config=/path/to/your/configfile
```

The **relay proxy** will read the configuration file and retrieve all the flags.    
After that you can use all the available endpoints _(see **Service endpoints** section)_ and get the variations for your users.


## Deployment options

A common way to run **go-feature-flag relay proxy** is to use the Docker Container.  
An image is available on docker Hub [`thomaspoignant/go-feature-flag-relay-proxy`](https://hub.docker.com/r/thomaspoignant/go-feature-flag-relay-proxy).

You can also run it as a service in your application following the **Installation** section.

## Specifying a configuration

To configure the relay proxy you should provide a configuration file when launching the instance.

The easiest way to provide the file is to use the option `--config=/path_to_your_file.yaml`.  
But if you don't provide this option, the relay proxy will look in these folders if a file named `goff-proxy.yaml` is available.

- current folder
- `/goff/`
- `/etc/opt/goff/`

To learn how to configure the relay proxy, read [Configuration](docs/configuration.md).

## Exporting metrics and traces

To export the data you can use all the capabilities of `go-feature-flag` SDK.  
To configure it please refer to the [type `exporter` section](docs/configuration.md#exporter) of the configuration.


## Service endpoints
The Relay Proxy defines many HTTP/HTTPS endpoints. 
Most of these are proxies for GO Feature Flag services, to be used by SDKs that connect to the Relay Proxy.  
Others are specific to the Relay Proxy, such as for monitoring its status.

Please refer to [endpoints documentation](docs/endpoints.md) to get the full details of what exists in our REST API.

## How can I contribute?
This project is open for contribution, see the [contributor's guide](CONTRIBUTING.md) for some helpful tips.
