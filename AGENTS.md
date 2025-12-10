# AGENTS.md - GO Feature Flag Codebase Guide

Essential information for AI agents and developers working with the GO Feature Flag codebase.

## 🎯 Project Overview

**GO Feature Flag** is a lightweight, open-source feature flagging solution written in Go. This repository is a **monorepo** using Go workspaces.

**Key Features:**
- Feature flag implementation with OpenFeature standard support
- Multiple storage backends (S3, HTTP, Kubernetes, MongoDB, Redis, etc.)
- Complex rollout strategies (A/B testing, progressive, scheduled)
- Data export and notification systems

## 🏗️ Architecture

```
OpenFeature SDKs → Relay Proxy (cmd/relayproxy/) → GO Module (ffclient)
                                                         ↓
                    Retriever Manager → Cache Manager → Notifier Service / Exporter Manager
```

**Data Flow:** Retrievers fetch configs → Cache stores flags → Change detection triggers Notifiers → Evaluation generates Events → Exporters send data

**Key Components:**
- **`ffclient`**: Core Go module for direct integration
- **`cmd/relayproxy/`**: HTTP API server (uses Echo + Zap logging)
- **`retriever/`**: Flag configuration sources (file, HTTP, S3, K8s, MongoDB, Redis, GitHub, GitLab, Bitbucket, PostgreSQL, Azure)
- **`exporter/`**: Data export destinations (S3, File, Kafka, Kinesis, Webhook, GCS, Pub/Sub, SQS, Azure)
- **`notifier/`**: Change notifications (Slack, Webhook, Discord, Teams, Logs)
- **`modules/core`** & **`modules/evaluation`**: Core logic modules used by OpenFeature providers and WASM

## 📁 Directory Structure

**Root Level:**
- `ffclient/`: Core client package
- `cmd/`: Applications (relayproxy, cli, lint, editor, wasm)
- `retriever/`, `exporter/`, `notifier/`: Integration packages
- `modules/`: Separate Go modules (core, evaluation)
- `internal/`: Internal packages (cache, flagstate, notification, signer)
- `openfeature/providers/`: Some providers (Kotlin, Python) - most are in OpenFeature contrib repos
- `testutils/`, `testdata/`, `examples/`, `website/`

**Key Files:**
- `variation.go`: Flag evaluation methods
- `feature_flag.go`: Core logic
- `config.go`: Configuration structure
- `Makefile`: Primary interface (use `make help`)

## 🔑 Key Concepts

**Flag Evaluation Flow:**
1. Init → Retrievers fetch configs → Cache stores → Polling refreshes
2. Change detection → Notifiers triggered → Evaluation → Events exported

**Flag Format:** YAML/JSON/TOML with variations, targeting rules, and defaultRule

**Evaluation Context:** Targeting key (required), custom attributes, optional bucketing key

**Interfaces:**
- `Retriever`: `Retrieve(ctx context.Context) ([]byte, error)`
- `Exporter`: `Export(ctx context.Context, events []FeatureEvent) error`, `IsBulk() bool`
- `Notifier`: `Notify(cache DiffCache) error`

## 🛠️ Common Tasks

**Adding Retriever/Exporter/Notifier:**
1. Create package in respective directory
2. Implement interface
3. Add config struct
4. Register in manager
5. Add table-driven tests
6. Update docs

**OpenFeature Providers:**
- Most providers in OpenFeature contrib repos (Go, JS, Java, .NET, Ruby, Swift, PHP)
- Some in this repo: Kotlin (`kotlin-provider/`), Python (`python-provider/`)
- Providers use `modules/core` and `modules/evaluation` for evaluation logic

**Modules (`modules/core` & `modules/evaluation`):**
- Core logic separated for reuse by OpenFeature providers and WASM module
- `modules/core`: Flag structures, context, models, utilities
- `modules/evaluation`: Evaluation logic (depends on core)
- Allows independent versioning and smaller dependency trees

**PR Policy:** Must use `.github/PULL_REQUEST_TEMPLATE.md`. Fill all sections, link issues, complete checklist, include tests.

## 🧪 Testing

**Style:** Table-driven tests (test array style) - standard Go pattern

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        args    args
        want    expectedType
        wantErr assert.ErrorAssertionFunc
    }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test logic with assert.Equal, assert.NoError, etc.
        })
    }
}
```

**Commands:** `make test`, `make coverage`, `make bench`  
**Coverage:** Aim for 90%+, use `testify/assert`, mock external deps

## 📝 Code Patterns

**Initialization:**
```go
ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &httpretriever.Retriever{URL: "..."},
})
defer ffclient.Close()
```

**Evaluation:**
```go
user := ffcontext.NewEvaluationContext("user-key")
value, _ := ffclient.BoolVariation("flag-name", user, false)
```

**Logging:**
- **Go Module**: Uses `slog` (`utils/fflog/`) - structured logging
- **Relay Proxy**: Uses Echo + Zap (`cmd/relayproxy/log/`, `api/middleware/zap.go`) - request logging

**Error Handling:** Always return errors, never panic. Use default values on failure.

## 🚀 Development Workflow

> **💡 Use Makefile for all tasks** - `make help` shows all commands

**Setup:**
```bash
make workspace-init  # Initialize Go workspace (monorepo requirement)
make vendor          # Vendor dependencies
pre-commit install   # Install pre-commit hooks
```

**Common Commands:**
- **Build:** `make build`, `make build-relayproxy`, `make build-cli`, `make build-wasm`
- **Dev:** `make watch-relayproxy`, `make watch-doc`
- **Test:** `make test`, `make coverage`, `make bench`
- **Quality:** `make lint`, `make tidy`, `make vendor`
- **Utils:** `make clean`, `make swagger`, `make generate-helm-docs`

**Code Quality:**
- Use `make lint` (golangci-lint)
- Write table-driven tests
- Update docs for user-facing changes

## 🔍 Code Navigation

**Flag Evaluation:** `variation.go` → `feature_flag.go` → `internal/cache/` → `internal/flagstate/`  
**API Endpoints:** `cmd/relayproxy/api/routes_*.go` → `controller/` → `model/`  
**Configuration:** `config.go` (ffclient), `cmd/relayproxy/config/config.go` (relay proxy)  
**Flag Format:** `.schema/flag-schema.json`, `testdata/flag-config.*`, https://gofeatureflag.org/docs

## 🔗 Important Files

- **`Makefile`**: Primary interface - `make help` for commands
- **`go.mod`**: Dependencies (monorepo with `modules/core`, `modules/evaluation`, `cmd/wasm`)
- **`.golangci.yml`**: Linter config
- **`CONTRIBUTING.md`**: Contribution guidelines
- **`.github/PULL_REQUEST_TEMPLATE.md`**: PR template (required)

## 🌐 External Dependencies

**Key Libraries:** Echo (HTTP), Koanf (config), OpenTelemetry (observability), Prometheus (metrics), Testcontainers (integration tests)

**Monorepo Modules:**
- Main module (root): Core library
- `modules/core`: Core data structures (used by providers/WASM)
- `modules/evaluation`: Evaluation logic (used by providers/WASM)
- `cmd/wasm`: WebAssembly evaluation
- `openfeature/providers/`: Some providers (most in contrib repos)

## 📚 Resources

- **Documentation**: https://gofeatureflag.org/docs
- **Examples**: `examples/` directory
- **API Docs**: `cmd/relayproxy/docs/`
- **OpenFeature**: https://openfeature.dev

## 🎯 Quick Reference

**Entry Points:** `ffclient.Init()`, `cmd/relayproxy/main.go`, `cmd/cli/main.go`  
**Interfaces:** `Retriever`, `Exporter`, `Notifier`, `Cache`  
**Config:** YAML/JSON/TOML flags, `ffclient.Config`, `cmd/relayproxy/config/config.go`

---

**Note**: This is a living document. Update as the codebase evolves!
