---
sidebar_position: 20
description: Relay proxy is the component that will evaluate the flags, this page explain how to install it.
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
