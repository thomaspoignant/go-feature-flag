# GO Feature Flag Python Provider

OpenFeature Python provider for [GO Feature Flag](https://gofeatureflag.org).

## Project Overview

This is a Python package that implements the OpenFeature provider interface to connect to a GO Feature Flag relay proxy. It enables Python applications to evaluate feature flags using the OpenFeature SDK.

## Architecture

```
gofeatureflag_python_provider/
├── __init__.py                 # Package exports
├── provider.py                 # Main GoFeatureFlagProvider class (AbstractProvider implementation)
├── options.py                  # GoFeatureFlagOptions configuration class
├── hooks/                      # OpenFeature hooks
│   ├── __init__.py
│   ├── data_collector.py       # Hook for collecting flag evaluation usage data
│   └── enrich_evaluation_context.py  # Hook that adds gofeatureflag metadata to context before evaluation
├── metadata.py                 # Provider metadata
├── request_data_collector.py   # Data models for usage collection
├── request_flag_evaluation.py  # Request models for flag evaluation API calls
└── response_flag_evaluation.py # Response models for flag evaluation API calls

tests/
├── test_gofeatureflag_python_provider.py  # Main provider tests
├── test_enrich_evaluation_context_hook.py # EnrichEvaluationContextHook tests
├── test_provider_graceful_exit.py         # Shutdown/cleanup tests
├── test_websocket_cache_invalidation.py   # WebSocket cache invalidation tests
├── mock_responses/                        # JSON mock responses for testing
├── config.goff.yaml                       # Test flag configuration
└── docker-compose.yml                     # Test infrastructure
```

## Key Components

### GoFeatureFlagProvider (`provider.py`)
- Extends `AbstractProvider` from OpenFeature SDK
- Implements all resolve methods: `resolve_boolean_details`, `resolve_string_details`, `resolve_integer_details`, `resolve_float_details`, `resolve_object_details`
- Uses `generic_go_feature_flag_resolver` for all flag types
- Features:
  - LRU cache for flag evaluations (`pylru`)
  - WebSocket connection for cache invalidation
  - Data collection for usage analytics

### GoFeatureFlagOptions (`options.py`)
Configuration options:
- `endpoint` (required): URL of the GO Feature Flag relay proxy
- `cache_size`: Max cached flags (default: 10000)
- `data_flush_interval`: Interval to flush usage data in ms (default: 60000)
- `disable_data_collection`: Turn off usage tracking (default: false)
- `reconnect_interval`: WebSocket reconnect interval in seconds (default: 60)
- `disable_cache_invalidation`: Disable WebSocket cache invalidation (default: false)
- `api_key`: API key for authenticated requests
- `exporter_metadata`: Custom metadata for evaluation events
- `debug`: Enable debug logging (default: false)
- `urllib3_pool_manager`: Custom HTTP client

### DataCollectorHook (`hooks/data_collector.py`)
- OpenFeature Hook implementation for collecting usage data
- Tracks flag evaluations via `after()` and `error()` hooks
- Flushes data to `/v1/data/collector` endpoint periodically

### EnrichEvaluationContextHook (`hooks/enrich_evaluation_context.py`)
- Enriches the evaluation context with a `gofeatureflag` attribute (from `exporter_metadata`) before flag resolution
- Used by the relay proxy for analytics or filtering; registered automatically by the provider

## Development

### Prerequisites
- Python 3.9+
- uv (package manager)

### Setup
```bash
# Install dependencies
uv sync

# Run a command in the virtual environment
uv run <command>
```

### Running Tests
```bash
# Run all tests
uv run pytest

# Run specific test file
uv run pytest tests/test_gofeatureflag_python_provider.py

# Run with verbose output
uv run pytest -v
```

### Code Style
```bash
# Format code with black
uv run black gofeatureflag_python_provider tests
```

## Key Patterns

### Pydantic Models
- All data classes extend Pydantic `BaseModel` for validation
- Request/response models use `model_dump_json()` for serialization
- Use `model_validate_json()` for deserialization

### HTTP Communication
- Uses `urllib3.PoolManager` for HTTP requests
- POST to `/v1/feature/{flag_key}/eval` for flag evaluation
- POST to `/v1/data/collector` for usage data
- WebSocket at `/ws/v1/flag/change` for cache invalidation

### Caching Strategy
- LRU cache keyed by `{flag_key}:{evaluation_context_hash()}`
- Cache cleared on WebSocket message (flag config changed)
- Set `cacheable` field in response determines if result is cached

### Error Handling
- `FlagNotFoundError`: Flag doesn't exist (404)
- `InvalidContextError`: Invalid evaluation context (400)
- `TypeMismatchError`: Response type doesn't match expected type
- `GeneralError`: Other errors (500+)

## API Reference

The provider communicates with the GO Feature Flag relay proxy:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/feature/{flag}/eval` | POST | Evaluate a flag |
| `/v1/data/collector` | POST | Send usage data |
| `/ws/v1/flag/change` | WebSocket | Cache invalidation notifications |

## Dependencies

Core:
- `openfeature-sdk`: OpenFeature Python SDK
- `pydantic`: Data validation
- `urllib3`: HTTP client
- `pylru`: LRU cache implementation
- `websocket-client`: WebSocket support
- `rel`: WebSocket reconnection handling

Dev:
- `pytest`: Testing framework
- `black`: Code formatter
- `pytest-docker`: Docker-based integration tests
