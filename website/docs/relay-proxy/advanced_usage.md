---
sidebar_position: 90
description: All the advanced usage of the relay proxy.
---
# ‼️ Advanced usage

## Manually trigger retrievers and refresh the internal flag cache
By default, the retrievers are called regularly to refresh the configuration based on the polling interval.

But there are use cases where you want to refresh the configuration explicitly _(for example, during the CI process
after you have changed your configuration file)_.

To do so you can call the `/v1/admin/retriever/refresh` endpoint with a POST request.  
It will force the retrievers to be called and update the internal cache.

```shell
curl -X 'POST' \
  'http://<your_domain>:1031/admin/v1/retriever/refresh' \
  -H 'accept: application/json' \
  -H 'X-API-Key: <your_admin_api_key>' \
  -d ''
```

:::note
This endpoint must be called with an **admin token**.
Authorized keys should be configured in the relay-proxy configuration file under the key **`authorizedKeys.admin`**.

```yaml title="goff-proxy.yaml"
# ...
authorizedKeys:
  admin:
    - "YOUR_API_KEY"
```
:::

## 🔒 FIPS 140-3 mode

GO Feature Flag publishes **FIPS 140-3 validated** builds of the relay proxy so it can be
deployed inside FedRAMP-authorized boundaries that require FIPS 140 validated cryptography
(controls **SC-13** / **SC-8**).

The FIPS builds are produced with Go's native FIPS mode (`GOFIPS140`, available since
Go 1.24) backed by the [Go Cryptographic Module](https://go.dev/doc/security/fips140),
which goes through the CMVP validation process. Building with `GOFIPS140=v1.0.0` bakes a
default `GODEBUG=fips140=on` into the binary, so the FIPS image runs in FIPS mode out of
the box — no extra runtime configuration is required.

:::info Scope
**Inbound TLS is out of scope.** The relay proxy serves plain HTTP and expects TLS to be
terminated at the ingress / service mesh, so north-south FIPS-validated TLS is handled by
your authorized infrastructure.

The FIPS build matters for the proxy's **outbound** crypto: pulling flag configuration
over HTTPS, webhook notifiers, and OTLP export.
:::

### Pulling the FIPS image

FIPS images are published to the same `gofeatureflag/go-feature-flag` repository, with a
`-fips` tag suffix.

```shell
# Latest FIPS build
docker pull gofeatureflag/go-feature-flag:fips

# Pinned to a specific version
docker pull gofeatureflag/go-feature-flag:v1.2.3-fips
```

Available platforms: `linux/amd64` and `linux/arm64`.

FIPS binaries are also attached to every
[GitHub Release](https://github.com/thomaspoignant/go-feature-flag/releases) as
`go-feature-flag-fips_<version>_Linux_x86_64` / `..._arm64` archives.

### Building from source

```shell
make build-relayproxy-fips
```

This produces `out/bin/relayproxy-fips`, built with `GOFIPS140=v1.0.0`.

### Verifying FIPS mode

**At build level** — confirm the binary was compiled against the Go Cryptographic Module:

```shell
go version -m ./out/bin/relayproxy-fips | grep GOFIPS140
# build	GOFIPS140=v1.0.0-<commit>
```

The Go toolchain appends the module's commit hash (e.g. `v1.0.0-c2097c7c`), so the
presence of a `GOFIPS140` build setting is what confirms the FIPS module was linked.

**At runtime** — on startup the relay proxy logs whether FIPS crypto is active:

```json
{"level":"info","msg":"crypto mode","fips140":true}
```

A standard (non-FIPS) build logs `"fips140":false`.

**Strict enforcement (optional)** — run with `GODEBUG=fips140=only` to make the process
fail if any non-FIPS-approved cryptographic algorithm is used:

```shell
GODEBUG=fips140=only ./out/bin/relayproxy-fips --config=/goff/config.yaml
```

### Recommended hardening for FedRAMP deployments (Helm)

If you deploy the relay proxy with our [Helm chart](./deployment/helm), the FIPS build
covers cryptography but a hardened deployment should also tighten the container's runtime
security context. The chart exposes `securityContext` options that are commented out by
default — for FedRAMP-style boundaries we recommend enabling at least:

```yaml
# values.yaml (Helm)
securityContext:
  runAsNonRoot: true
  readOnlyRootFilesystem: true
```

This runs the proxy as a non-root user with an immutable root filesystem. Combine it with
your platform's pod security standards as appropriate.

