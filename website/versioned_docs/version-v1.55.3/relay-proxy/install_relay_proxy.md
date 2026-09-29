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
FROM gofeatureflag/go-feature-flag:trixie
```

#### <i class="fa-brands fa-github"></i> GitHub Container Registry
The same images are also published to the GitHub Container Registry, for those who cannot
reach Docker Hub.
```shell
docker pull ghcr.io/go-feature-flag/go-feature-flag:latest
```
Mirroring starts with the first release after `v1.55.1`, so earlier versions and
their floating tags (for example `v1.55`) remain available on Docker Hub only.

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

## <i class="fa-solid fa-lock"></i> FIPS 140-3
If you need FIPS 140 validated cryptography (e.g. FedRAMP), use the FIPS-tagged image:
```shell
docker pull gofeatureflag/go-feature-flag:fips
# or from the GitHub Container Registry
docker pull ghcr.io/go-feature-flag/go-feature-flag:fips
```
See [FIPS 140-3 mode](./advanced_usage#-fips-140-3-mode) for details on running and verifying it.

## Summary
Once **GO Feature Flag** is installed, you can start using it within your application by connecting your OpenFeature provider to it.