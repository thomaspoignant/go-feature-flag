# 🛠️ Installation

## **Docker[​](#-docker "Direct link to -docker")

#### **Install from the command line[​](#-install-from-the-command-line "Direct link to -install-from-the-command-line")

```
docker pull gofeatureflag/go-feature-flag:latest
```

#### **Use as base image in Dockerfile[​](#-use-as-base-image-in-dockerfile "Direct link to -use-as-base-image-in-dockerfile")

```
FROM gofeatureflag/go-feature-flag:trixie
```

#### **GitHub Container Registry[​](#-github-container-registry "Direct link to -github-container-registry")

The same images are also published to the GitHub Container Registry, for those who cannot reach Docker Hub.

```
docker pull ghcr.io/go-feature-flag/go-feature-flag:latest
```

Mirroring starts with the first release after `v1.55.1`, so earlier versions and their floating tags (for example `v1.55`) remain available on Docker Hub only.

## **HomeBrew (macOS and Linux)[​](#-homebrew-macos-and-linux "Direct link to -homebrew-macos-and-linux")

```
brew install go-feature-flag
```

## **Scoop (Windows)[​](#-scoop-windows "Direct link to -scoop-windows")

```
scoop install go-feature-flag
```

## ******Binary[​](#---binary "Direct link to ---binary")

All binaries are available in x86/ARM in the [GitHub Release Page](https://github.com/thomaspoignant/go-feature-flag/releases).

## **FIPS 140-3[​](#-fips-140-3 "Direct link to -fips-140-3")

If you need FIPS 140 validated cryptography (e.g. FedRAMP), use the FIPS-tagged image:

```
docker pull gofeatureflag/go-feature-flag:fips

# or from the GitHub Container Registry

docker pull ghcr.io/go-feature-flag/go-feature-flag:fips
```

See [FIPS 140-3 mode](https://gofeatureflag.org/docs/relay-proxy/advanced_usage.md#-fips-140-3-mode) for details on running and verifying it.

## Summary[​](#summary "Direct link to Summary")

Once **GO Feature Flag** is installed, you can start using it within your application by connecting your OpenFeature provider to it.
