# Threat model

## What this project does and where untrusted input enters
GO Feature Flag is a feature flag solution: a Go library (root module, `modules/core`) that evaluates flags, and a
relay proxy (`cmd/relayproxy`) that exposes the same evaluation over HTTP for OpenFeature SDKs in other languages.

Untrusted input:
- **Relay proxy HTTP API** — the main attack surface. Callers are SDKs/clients holding (at most) an evaluation API key:
  - `POST /v1/allflags`, `POST /v1/feature/:flagKey/eval`, `POST /v1/data/collector`, `GET /v1/flag/change`,
    `POST /v1/flag/configuration`
  - OFREP: `POST /ofrep/v1/evaluate/flags`, `POST /ofrep/v1/evaluate/flags/:flagKey`
  - streams: `GET /stream/v1/ws/flag/change` (websocket), `GET /stream/v1/sse/flag/change`, legacy
    `GET /ws/v1/flag/change`
  - `GET /openfeature/v0/manifest`
  - admin: `POST /admin/v1/retriever/refresh` (admin key only)
  - unauthenticated: `/health`, `/info`, `/metrics`, `/swagger/*` (when enabled)
  - API keys (`authorizedKeys`, `X-API-Key` / `Authorization: Bearer`) and flag sets are enforced in
    `cmd/relayproxy/api/middleware`.
- **Evaluation context** (targeting key + arbitrary attributes) sent by callers; it is matched against flag rules
  (`internal/` rule/query evaluation, `nikunjy/rules` fork, JsonLogic).
- **Editor API** (`cmd/editor`) — evaluates a flag config posted by the caller.

Trusted (operator-controlled): the relay proxy config file, environment variables, and the flag configuration files the
retrievers fetch (file, HTTP, S3, GCS, GitHub, GitLab, Bitbucket, K8s, Redis, MongoDB, PostgreSQL, Azure Blob).

## Components that matter most / least
- Most: relay proxy authentication/authorization (API keys, admin vs evaluation keys, flag-set isolation), request
  parsing and evaluation endpoints, rule/targeting evaluation, websocket/SSE handling.
- Medium: flag config parsing (YAML/JSON/TOML) and retrievers, exporters and notifiers (they send data to
  operator-configured destinations), CLI (`cmd/cli`: lint/evaluate local files).
- Out of scope: `website/`, `examples/`, `.github/`, the non-Go OpenFeature providers under `openfeature/`,
  `cmd/wasm` (needs TinyGo, not built here).

## How to exercise it
- Binaries are built in `/src/out/bin/`: `relayproxy`, `cli`, `editor-api`, `jsonschema-generator`.
- Start the relay proxy offline: `out/bin/relayproxy --config .oss-scanner/scanner-proxy.yaml` (file retriever, evaluation
  key `eval-key`, admin key `admin-key`), then e.g.
  `curl -X POST -H 'X-API-Key: eval-key' -H 'Content-Type: application/json' -d '{"context":{"targetingKey":"u1"}}' localhost:1031/ofrep/v1/evaluate/flags`
- More configs in `cmd/relayproxy/testdata/config/`; flag files in `testdata/`.
- Tests: `go test -short ./...` (root) and `cd modules/core && go test ./...`. Tests tagged `docker` need a Docker daemon
  and cannot run here.

## How you rate severity
- Critical: authentication bypass on the relay proxy (evaluating or refreshing without a valid key, admin endpoints
  with an evaluation key), reading another flag set's flags with a key scoped to a different flag set, RCE, SSRF or
  file read reachable from an API caller.
- High: crash / panic of the relay proxy triggered by an unauthenticated or evaluation-key caller, leaking API keys or
  secrets from the config through any endpoint.
- Medium: resource exhaustion DoS from an authenticated caller, incorrect flag evaluation that a caller can force
  (e.g. bypassing a targeting rule).
- Low: anything that requires control of the relay proxy config file, environment, or the flag configuration source
  (those are trusted); issues in the CLI on local files.

## Anything to leave alone
- Do not report that `/health`, `/info` and `/metrics` are unauthenticated, or that auth is disabled when no
  `authorizedKeys` are configured: both are by design.
- Do not report vulnerable dependencies without a reachable path in this code.
