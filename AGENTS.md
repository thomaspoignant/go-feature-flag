# AGENT.md - GO Feature Flag Codebase Guide

This document provides AI agents and developers with essential information to understand and work effectively with the GO Feature Flag codebase.

## 🎯 Project Overview

**GO Feature Flag** is a lightweight, open-source feature flagging solution written in Go. This repository is a **monorepo** containing multiple related projects and modules.

The project provides:
- A simple, complete feature flag implementation
- Support for multiple languages through OpenFeature standard
- Self-hosted solution with no backend server required
- Multiple storage backends (S3, HTTP, Kubernetes, MongoDB, Redis, etc.)
- Complex rollout strategies (A/B testing, progressive rollout, scheduled rollout)
- Data export capabilities
- Notification system for flag changes

## 🏗️ Architecture

### High-Level Architecture

```
┌─────────────────┐
│  OpenFeature    │
│     SDKs        │  (Multiple languages)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Relay Proxy    │  (API Server - cmd/relayproxy/)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  GO Module      │  (Core library - ffclient package)
│  (ffclient)     │
└────────┬────────┘
         │
    ┌────┴────┐
    ▼        ▼
┌──────────┐  ┌──────────┐
│Retriever │  │  Cache   │
│ Manager  │─▶│ Manager  │
└──────────┘  └────┬──────┘
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
   ┌────────┐ ┌────────┐ ┌──────────┐
   │Notifier│ │Exporter│ │Evaluation│
   │Service │ │Manager │ │  Events  │
   └────┬───┘ └────┬───┘ └───────────┘
        │          │
        ▼          ▼
   ┌────────┐ ┌──────────┐
   │Notifiers│ │Exporters │
   │(Slack,  │ │(S3, File,│
   │Webhook, │ │Kafka,    │
   │etc.)    │ │etc.)     │
   └─────────┘ └──────────┘
```

**Data Flow:**

**Configuration Loading:**
1. **Retrievers** fetch flag configurations from various sources (file, HTTP, S3, K8s, MongoDB, Redis, etc.)
2. **Retriever Manager** coordinates multiple retrievers and handles polling
3. **Cache Manager** stores flags in memory and manages flag state

**Change Detection & Notifications:**
4. **Cache Manager** compares old/new cache after each refresh
5. **Notification Service** detects differences (added, updated, deleted flags)
6. **Notifiers** are triggered asynchronously when changes are detected
   - Each notifier runs in its own goroutine
   - Supports Slack, Webhook, Discord, Microsoft Teams, and Logs

**Flag Evaluation & Export:**
7. **Evaluation** happens when `BoolVariation()`, `StringVariation()`, etc. are called
8. **Evaluation Events** (FeatureEvent/TrackingEvent) are generated
9. **Exporter Manager** collects events in an event store
10. **Exporters** receive events either immediately (non-bulk) or in batches (bulk mode)
    - Bulk exporters flush when max events reached or flush interval elapsed

### Key Components

1. **Core Library (`ffclient` package)**: Main Go module for direct integration
2. **Relay Proxy (`cmd/relayproxy/`)**: HTTP API server exposing feature flags
3. **Retrievers (`retriever/`)**: Fetch flag configurations from various sources
   - Managed by `retriever.Manager`
   - Supports polling and background updates
4. **Cache (`internal/cache/`)**: In-memory storage for flag configurations
   - Managed by `cache.Manager`
   - Detects flag changes and triggers notifications
5. **Notifiers (`notifier/`)**: Send notifications when flags change
   - Managed by `internal/notification.Service`
   - Triggered when cache detects flag additions, updates, or deletions
   - Supports Slack, Webhook, Discord, Microsoft Teams, and Logs
6. **Exporters (`exporter/`)**: Export flag evaluation data to various destinations
   - Managed by `exporter.Manager`
   - Receives evaluation events (FeatureEvent and TrackingEvent)
   - Supports S3, File, Kafka, Kinesis, Webhook, GCS, Pub/Sub, SQS, Azure
7. **OpenFeature Providers**: Language-specific providers
   - **Most providers** are in separate OpenFeature contrib repositories (e.g., `open-feature/js-sdk-contrib`, `open-feature/go-sdk-contrib`)
   - **Some providers** are maintained in this repository (`openfeature/providers/`) - Kotlin, Python
   - Providers use `modules/core` and `modules/evaluation` for flag evaluation logic

## 📁 Directory Structure

### Root Level

- **`ffclient/`**: Core feature flag client package (main entry point)
- **`cmd/`**: Command-line applications
  - `relayproxy/`: HTTP API server
  - `cli/`: CLI tools (lint, evaluate, generate)
  - `editor/`: Editor API
  - `lint/`: Linter tool
  - `wasm/`: WebAssembly evaluation library
- **`retriever/`**: Flag configuration retrievers (file, HTTP, S3, K8s, MongoDB, Redis, etc.)
- **`exporter/`**: Data exporters (file, S3, Kinesis, Kafka, webhook, etc.)
- **`notifier/`**: Notification systems (Slack, webhook, Discord, Teams)
- **`openfeature/`**: Some OpenFeature provider implementations (Kotlin, Python)
  - Note: Most OpenFeature providers are in separate OpenFeature contrib repositories
- **`modules/`**: Separate Go modules (evaluation, core)
- **`internal/`**: Internal packages (cache, flagstate, notification)
- **`testutils/`**: Testing utilities and mocks
- **`examples/`**: Example implementations
- **`website/`**: Documentation website (Docusaurus)

### Key Packages

#### Core Packages

- **`ffclient`**: Main package - initialization and flag evaluation
- **`ffcontext`**: Evaluation context creation
- **`ffuser`**: User context utilities
- **`variation.go`**: Flag variation evaluation methods
- **`feature_flag.go`**: Core feature flag logic

#### Internal Packages (`internal/`)

- **`cache/`**: In-memory cache for flag configurations
- **`flagstate/`**: Flag state management
- **`notification/`**: Notification service
- **`signer/`**: Webhook signature generation

#### Retriever Packages (`retriever/`)

Each retriever implements the `Retriever` interface:
- `fileretriever/`: Local file system
- `httpretriever/`: HTTP endpoint
- `s3retrieverv2/`: AWS S3
- `gcstorageretriever/`: Google Cloud Storage
- `k8sretriever/`: Kubernetes ConfigMaps
- `mongodbretriever/`: MongoDB
- `redisretriever/`: Redis
- `githubretriever/`: GitHub
- `gitlabretriever/`: GitLab
- `bitbucketretriever/`: Bitbucket
- `postgresqlretriever/`: PostgreSQL
- `azblobstorageretriever/`: Azure Blob Storage

#### Exporter Packages (`exporter/`)

Each exporter implements the `Exporter` interface:
- `fileexporter/`: Local file export
- `logsexporter/`: Log-based export
- `s3exporterv2/`: AWS S3 export
- `kinesisexporter/`: AWS Kinesis
- `kafkaexporter/`: Kafka
- `webhookexporter/`: HTTP webhook
- `gcstorageexporter/`: Google Cloud Storage
- `pubsubexporter/`: Google Pub/Sub
- `sqsexporter/`: AWS SQS
- `azureexporter/`: Azure export

## 🔑 Key Concepts

### Flag Evaluation Flow

1. **Initialization**: `ffclient.Init()` initializes retrievers, cache, exporters, and notifiers
2. **Retrieval**: Retrievers fetch flag configurations from configured sources
3. **Caching**: Flags are stored in memory cache (`internal/cache/`)
4. **Polling**: Periodic refresh from retriever (configurable interval)
5. **Change Detection**: Cache manager compares old/new flags and detects changes
6. **Notification**: When changes detected, notification service triggers all configured notifiers
7. **Evaluation**: `BoolVariation()`, `StringVariation()`, etc. evaluate flags from cache
8. **Event Export**: Evaluation events are sent to exporter manager, which distributes to configured exporters

### Flag Configuration Format

Flags are stored in YAML, JSON, or TOML format:

```yaml
flag-name:
  variations:
    variation-a: true
    variation-b: false
  targeting:
    - query: key eq "specific-user"
      percentage:
        variation-a: 100
        variation-b: 0
  defaultRule:
    variation: variation-a
```

### Evaluation Context

- **Targeting Key**: Unique identifier for consistent flag evaluation
- **Custom Attributes**: Additional context for targeting rules
- **Bucketing Key**: Optional key for team/group-based bucketing

### Retriever Pattern

All retrievers implement:
```go
type Retriever interface {
    Retrieve(ctx context.Context) ([]byte, error)
}
```

### Exporter Pattern

All exporters implement:
```go
type Exporter interface {
    Export(ctx context.Context, events []FeatureEvent) error
    IsBulk() bool
}
```

## 🛠️ Common Tasks

### Adding a New Retriever

1. Create new package in `retriever/yourretriever/`
2. Implement `Retriever` interface
3. Add configuration struct
4. Register in `retriever/manager.go`
5. Add tests in `retriever/yourretriever/retriever_test.go`
6. Update documentation

### Adding a New Exporter

1. Create new package in `exporter/yourexporter/`
2. Implement `Exporter` interface
3. Add configuration struct
4. Register in `exporter/manager.go`
5. Add tests
6. Update documentation

### Adding a New Notifier

1. Create new package in `notifier/yournotifier/`
2. Implement `Notifier` interface
3. Add configuration struct
4. Register in `notifier/manager.go`
5. Add tests
6. Update documentation

### Working with OpenFeature Providers

**Provider Locations:**
- **Most providers** are in OpenFeature contrib repositories:
  - Go: `github.com/open-feature/go-sdk-contrib/providers/go-feature-flag`
  - JavaScript/TypeScript: `@openfeature/go-feature-flag-provider`
  - Java/Kotlin: `dev.openfeature.contrib.providers/go-feature-flag`
  - .NET: `OpenFeature.Contrib.GOFeatureFlag`
  - Python: `gofeatureflag-python-provider`
  - Ruby: `openfeature-go-feature-flag-provider`
  - Swift: `go-feature-flag/openfeature-swift-provider`
  - PHP: `open-feature/go-feature-flag-provider`
  
- **Providers in this repo** (`openfeature/providers/`):
  - Kotlin Provider (`kotlin-provider/`)
  - Python Provider (`python-provider/`)

**Developing Providers:**
- Providers should use `modules/core` and `modules/evaluation` for flag evaluation
- Connect to relay proxy via HTTP API or use OpenFeature SDK
- Follow OpenFeature specification for provider implementation
- See existing providers in contrib repos for examples

### Modifying Flag Evaluation Logic

- Core evaluation: `variation.go`
- Flag state: `internal/flagstate/`
- Cache management: `internal/cache/`
- Context handling: `ffcontext/`

### Working with Relay Proxy

- Main entry: `cmd/relayproxy/main.go`
- API routes: `cmd/relayproxy/api/routes_*.go`
- Controllers: `cmd/relayproxy/controller/`
- Configuration: `cmd/relayproxy/config/`
- Services: `cmd/relayproxy/service/`

### Opening Pull Requests

**⚠️ Important**: When opening a pull request, you **MUST** use the PR template located at `.github/PULL_REQUEST_TEMPLATE.md`.

**PR Template Policy:**
- The PR template is automatically populated when creating a new PR
- Fill out all relevant sections of the template
- Provide clear descriptions of changes
- Link to related issues if applicable
- Ensure all checklist items are completed before requesting review
- Include test coverage for new features or bug fixes

**Before submitting a PR:**
1. Ensure your code follows the project's coding standards
2. Run `make test` to verify all tests pass
3. Run `make lint` to check for linting issues
4. Update documentation if your changes affect user-facing features
5. Use the PR template to provide context and details

## 🧪 Testing

### Test Structure

- Test files follow `*_test.go` naming convention
- Test utilities in `testutils/`
- Mock implementations in `testutils/mock/`
- Test data in `testdata/`

### Testing Style: Table-Driven Tests

**This repository uses table-driven tests (test array style)** - the standard Go testing pattern. All tests should follow this structure:

```go
func TestFunctionName(t *testing.T) {
    type args struct {
        // function arguments
    }
    tests := []struct {
        name    string
        args    args
        want    // expected return value
        wantErr assert.ErrorAssertionFunc // or bool
    }{
        {
            name: "test case description",
            args: args{
                // test input
            },
            want:    // expected output
            wantErr: assert.NoError, // or assert.Error
        },
        {
            name: "another test case",
            args: args{
                // different test input
            },
            want:    // expected output
            wantErr: assert.Error,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.args.input)
            if tt.wantErr != nil {
                tt.wantErr(t, err)
            }
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**Key points:**
- Use `tests := []struct{...}` to define test cases
- Each test case should have a descriptive `name` field
- Use `t.Run(tt.name, ...)` to run subtests
- Use `testify/assert` for assertions (`assert.NoError`, `assert.Error`, `assert.Equal`, etc.)
- Group related test cases in the same test function

### Running Tests

**Use the Makefile (recommended):**
```bash
make test              # Run all tests
make coverage          # Run tests with coverage report
make bench             # Run benchmark tests
```

**Direct Go commands (alternative):**
```bash
go test ./...          # Run tests in current directory
go test -v ./pkg/...   # Verbose output for specific package
```

### Test Coverage

- Aim for 90%+ coverage
- Use `testify` for assertions (`assert` and `require` packages)
- Mock external dependencies using interfaces
- Use `testutils/mock/` for common mocks

## 🔍 Code Navigation Tips

### Finding Flag Evaluation Logic

1. Start at `variation.go` for evaluation methods
2. Check `feature_flag.go` for core logic
3. Look at `internal/cache/` for caching
4. See `internal/flagstate/` for state management

### Finding API Endpoints

1. Check `cmd/relayproxy/api/routes_*.go` for route definitions
2. Look at `cmd/relayproxy/controller/` for handlers
3. See `cmd/relayproxy/model/` for request/response models

### Finding Configuration Options

1. Core config: `config.go` in `ffclient` package
2. Relay proxy config: `cmd/relayproxy/config/config.go`
3. Retriever configs: `retriever/*/retriever.go`
4. Exporter configs: `exporter/*/exporter.go`

### Understanding Flag Format

- Schema: `.schema/flag-schema.json`
- Examples: `testdata/flag-config.*`
- Documentation: https://gofeatureflag.org/docs (source in `website/docs/`)

## 📝 Code Patterns

### Initialization Pattern

```go
err := ffclient.Init(ffclient.Config{
    PollingInterval: 3 * time.Second,
    Retriever: &httpretriever.Retriever{
        URL: "http://example.com/flags.yaml",
    },
})
defer ffclient.Close()
```

### Evaluation Pattern

```go
user := ffcontext.NewEvaluationContext("user-key")
value, _ := ffclient.BoolVariation("flag-name", user, false)
```

### Error Handling

- Always return errors, never panic
- Use default values when evaluation fails
- Log errors but don't break execution

### Logging Patterns

**Different logging libraries are used in different parts of the codebase:**

1. **Go Module (`ffclient` package)**: Uses **`slog`** (structured logging)
   - Located in `utils/fflog/`
   - Uses `slog.Logger` for structured logging
   - Supports both structured logging (slog) and legacy logging (log.Logger)
   - Log levels: Error, Info, Debug, Warn
   - Example: `logger.Info("message", slog.String("key", "value"))`

2. **Relay Proxy (`cmd/relayproxy/`)**: Uses **Echo with Zap**
   - Echo framework integrated with Zap logger via middleware
   - Located in `cmd/relayproxy/log/` and `cmd/relayproxy/api/middleware/zap.go`
   - Uses `go.uber.org/zap` for high-performance structured logging
   - Zap middleware (`ZapLogger`) is applied to Echo routes for request logging
   - Supports JSON and logfmt formats
   - Example: `logger.ZapLogger.Info("message", zap.String("key", "value"))`

**Key Points:**
- Use structured logging with key-value pairs
- Include context in log messages (request IDs, user IDs, etc.)
- Use appropriate log levels (Error for errors, Info for important events, Debug for detailed debugging)
- Never log sensitive information (passwords, tokens, etc.)

## 🚀 Development Workflow

> **💡 Important**: The **Makefile is the easiest way to interact with this repository**. All common tasks are available as Makefile targets. Use `make help` to see all available commands.

### Setup

Since this is a monorepo, you need to initialize the Go workspace first:

```bash
make workspace-init  # Initialize Go workspace (creates go.work file)
make vendor          # Vendor dependencies
pre-commit install   # Install pre-commit hooks
```

**Note**: The Go workspace (`go.work`) is required for the monorepo structure and allows Go to work with multiple modules in the same repository.

### Common Makefile Commands

**Build Commands:**
```bash
make build                    # Build all binaries (relayproxy, cli, lint, editor-api, jsonschema-generator)
make build-relayproxy         # Build relay proxy only
make build-cli                # Build CLI tools
make build-lint              # Build linter tool
make build-wasm              # Build WebAssembly library
make build-wasi              # Build WASI library
make build-modules           # Build all Go modules in workspace
make build-doc               # Build documentation website
```

**Development Commands:**
```bash
make watch-relayproxy        # Launch relay proxy in watch mode (auto-reload)
make watch-doc              # Start documentation server (auto-reload)
make serve-doc               # Serve built documentation
```

**Testing Commands:**
```bash
make test                    # Run all tests
make coverage                # Run tests with coverage report
make bench                   # Run benchmark tests
make provider-tests          # Run OpenFeature provider integration tests
```

**Code Quality:**
```bash
make lint                    # Run golangci-lint
make tidy                    # Run go mod tidy for all modules
make vendor                  # Vendor dependencies
```

**Utilities:**
```bash
make clean                   # Remove build artifacts and vendor directory
make help                    # Show all available Makefile commands
make swagger                 # Generate Swagger documentation
make generate-helm-docs      # Generate Helm chart documentation
```

### Documentation

```bash
make watch-doc               # Start documentation server (recommended for development)
make build-doc               # Build documentation for production
make serve-doc               # Serve the built documentation
```

### Code Quality

- Use `make lint` to run `golangci-lint` (configured in `.golangci.yml`)
- Follow Go best practices
- Write tests for new features (use table-driven test style)
- Update documentation

## 🔗 Important Files

- **`Makefile`**: **Primary interface for repository interaction** - Use `make help` to see all commands
- **`go.mod`**: Go module dependencies
- **`.golangci.yml`**: Linter configuration
- **`CONTRIBUTING.md`**: Contribution guidelines
- **`README.md`**: Project overview and quick start
- **`config.go`**: Core configuration structure
- **`feature_flag.go`**: Main feature flag implementation

## 🌐 External Dependencies

### Key Libraries

- **Echo**: HTTP framework (relay proxy)
- **Koanf**: Configuration management
- **OpenTelemetry**: Observability
- **Prometheus**: Metrics
- **Testcontainers**: Integration testing

### Monorepo Structure

This repository is organized as a **monorepo** using Go workspaces. The monorepo contains:

- **Main module** (root directory): Core GO Feature Flag library
- **`modules/evaluation/`**: Separate Go module for flag evaluation logic
- **`modules/core/`**: Separate Go module for core flag data structures
- **`cmd/wasm/`**: WebAssembly module for browser-side evaluation
- **`openfeature/providers/`**: Some OpenFeature providers maintained here (Kotlin, Python)
  - Most providers are in separate OpenFeature contrib repositories
- **`cmd/relayproxy/`**: Relay proxy server (part of main module)
- **`cmd/cli/`**: CLI tools (part of main module)
- **`cmd/lint/`**: Linter tool (part of main module)
- **`cmd/editor/`**: Editor API (part of main module)

**Go Workspace**: The project uses Go workspaces (`go.work`) to manage multiple modules. Run `make workspace-init` to set up the workspace.

### Module Details: `modules/core` and `modules/evaluation`

**`modules/core`** and **`modules/evaluation`** contain the core logic of GO Feature Flag. They are separated into different modules because they are reused by:

1. **OpenFeature Providers**: Most providers (in OpenFeature contrib repositories) use these modules for flag evaluation
   - Go Provider: `github.com/open-feature/go-sdk-contrib/providers/go-feature-flag`
   - Other providers connect via HTTP API to relay proxy
2. **WASM Module** (`cmd/wasm/`): Uses these modules for browser-side flag evaluation

**Module Structure:**
- **`modules/core/`**: Contains core flag data structures, context handling, flag models, and utilities
  - Flag definitions (`flag/`)
  - Evaluation context (`ffcontext/`)
  - Data models (`model/`)
  - Utility functions (`utils/`)
  - Minimal dependencies for reusability

- **`modules/evaluation/`**: Contains flag evaluation logic
  - Depends on `modules/core`
  - Provides type-safe evaluation functions (`Evaluate[T]()`)
  - Handles variation resolution and type conversion

**Why Separate Modules?**
- **Reusability**: Core logic can be imported independently without pulling in the entire main module
- **Dependency Management**: Allows different consumers (WASM, OpenFeature providers) to use only what they need
- **Versioning**: Modules can be versioned independently (see `modules/*/go.mod`)
- **Size Optimization**: WASM builds benefit from smaller dependency trees

**Usage Pattern:**
```go
// In WASM or OpenFeature provider
import (
    "github.com/thomaspoignant/go-feature-flag/modules/core/flag"
    "github.com/thomaspoignant/go-feature-flag/modules/core/ffcontext"
    "github.com/thomaspoignant/go-feature-flag/modules/evaluation"
)

// Use core types
ctx := ffcontext.NewEvaluationContext("user-key")
// Use evaluation logic
result, err := evaluation.Evaluate[bool](flag, "flag-key", ctx, ...)
```

## 📚 Additional Resources

- **Documentation**: https://gofeatureflag.org/docs/docs
- **Examples**: `examples/` directory
- **API Docs**: `cmd/relayproxy/docs/`
- **OpenFeature Spec**: https://openfeature.dev

## 🎯 Quick Reference

### Main Entry Points

- **Go Module**: `ffclient.Init()`
- **Relay Proxy**: `cmd/relayproxy/main.go`
- **CLI**: `cmd/cli/main.go`
- **Linter**: `cmd/lint/main.go`

### Key Interfaces

- `Retriever`: Flag configuration retrieval
- `Exporter`: Data export
- `Notifier`: Flag change notifications
- `Cache`: Flag caching

### Configuration Files

- Flag config: YAML/JSON/TOML format
- Relay proxy: `cmd/relayproxy/config/config.go`
- Go module: `ffclient.Config`

---

**Note**: This is a living document. Update it as the codebase evolves!
