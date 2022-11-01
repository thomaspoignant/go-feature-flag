---
sidebar_position: 20
description: Relay proxy is the component that will challenge the flags, this page explain how to install it.
---

# Install the relay proxy

## When should I use the GO Feature Flag Relay Proxy?
- If you want to access your feature flag using an API instead of the [`thomaspoignant/go-feature-flag`](https://github.com/thomaspoignant/go-feature-flag) SDK.
- If you are not using GOlang to build your application.
- If you want to reduce the number of accesses to your configuration flag by centralizing them.
<!-- - If you are using any SDKs that connect to the Relay Proxy. -->

## Install using Homebrew (mac and linux)
```shell
brew tap thomaspoignant/homebrew-tap
brew install go-feature-flag-relay-proxy
```

## Install using Scoop (windows)
```shell
scoop bucket add org https://github.com/thomaspoignant/scoop.git
scoop install go-feature-flag-relay-proxy
```

## Install using docker
```shell
docker pull thomaspoignant/go-feature-flag-relay-proxy:latest
```
:::info
More info in the [dockerhub page](https://hub.docker.com/r/thomaspoignant/go-feature-flag-relay-proxy).
:::

## Getting started

Before starting your **relay proxy** you will need to create a minimal configuration file.  

```yaml
# this is a minimal config containing only where your flag file is located 
retriever:
  kind: http
  url: https://raw.githubusercontent.com/thomaspoignant/go-feature-flag/main/examples/file/flags.yaml
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

- **current folder**
- `/goff/`
- `/etc/opt/goff/`

To learn how to configure the relay proxy, read [Configuration](../go_module/configuration).

## Exporting metrics and traces

To export the data you can use all the capabilities of `go-feature-flag` SDK.  
To configure it please refer to the [type `exporter` section](../go_module/configuration#exporter) of the configuration.

## Service endpoints
The Relay Proxy defines many HTTP/HTTPS endpoints. 
Most of these are proxies for GO Feature Flag services, to be used by SDKs that connect to the Relay Proxy.  
Others are specific to the Relay Proxy, such as for monitoring its status.

Please refer to [endpoints documentation](./relay_proxy_endpoints) to get the full details of what exists in our REST API.
