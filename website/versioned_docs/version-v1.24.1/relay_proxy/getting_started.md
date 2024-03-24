---
sidebar_position: 19
title: Getting started
description: Getting started with the relay proxy.
---

# Getting started

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
An image is available on Docker Hub [`thomaspoignant/go-feature-flag`](https://hub.docker.com/r/thomaspoignant/go-feature-flag).

You can also run it as a service in your application following the [**Installation**](install_relay_proxy) page.

```shell
docker run  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml thomaspoignant/go-feature-flag-relay-proxy:latest
```

## Specifying a configuration

To configure the relay proxy you should provide a configuration file when launching the instance.

The easiest way to provide the file is to use the option `--config=/path_to_your_file.yaml`.  
But if you don't provide this option, the relay proxy will look in these folders if a file named `goff-proxy.yaml` is available.

- **current folder**
- `/goff/`
- `/etc/opt/goff/`

To learn how to configure the relay proxy, read [Configuration](./configure_relay_proxy).

## Exporting metrics and traces

To export the data you can use all the capabilities of `go-feature-flag` SDK.  
To configure it please refer to the [type `exporter` section](./configure_relay_proxy#exporter) of the configuration documentation.

## Service endpoints
The Relay Proxy defines many HTTP/HTTPS endpoints. 
Most of these are proxies for GO Feature Flag services, to be used by SDKs that connect to the Relay Proxy.  
Others are specific to the Relay Proxy, such as for monitoring its status.

Please refer to [endpoints documentation](./relay_proxy_endpoints) to get the full details on available REST API endpoints.
