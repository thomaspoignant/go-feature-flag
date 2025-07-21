---
sidebar_position: 30
description: Relay proxy is the component that will evaluate the flags, this page explain how to install it.
---

# 🛠️ Installation

## <i class="fa-brands fa-docker"></i> Docker

#### <i class="fa-solid fa-terminal"></i> Install from the command line
```shell
docker pull gofeatureflag/go-feature-flag:latest
```

####  <i class="fa-brands fa-docker"></i> Use as base image in Dockerfile
```dockerfile
FROM gofeatureflag/go-feature-flag:bookworm
```

## <i class="fa-solid fa-beer-mug-empty"></i> HomeBrew (macOS and Linux)
```shell
brew install go-feature-flag
```

## <i class="fa-solid fa-ice-cream"></i> Scoop (Windows)
```shell
scoop install go-feature-flag
```
## <i class="fa-brands fa-windows"></i> <i class="fa-brands fa-apple"></i> <i class="fa-brands fa-linux"></i> Binary
All binaries are available in x86/ARM in the [GitHub Release Page](https://github.com/thomaspoignant/go-feature-flag/releases).

## Summary
Once **GO Feature Flag** is installed, you can start using it within your application by connecting your OpenFeature provider to it.