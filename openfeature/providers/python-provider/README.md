# GoFeatureFlag Python Provider for OpenFeature

A Python provider for [OpenFeature](https://openfeature.dev/) that integrates with [GoFeatureFlag](https://gofeatureflag.org/).

## Features

- **Remote Evaluation**: Uses the OFREP protocol for remote flag evaluation
- **In-Process Evaluation**: Uses WASM for local flag evaluation (coming soon)
- **Full OpenFeature Compliance**: Implements all required OpenFeature interfaces
- **Async Support**: Full async/await support for all operations
- **Comprehensive Testing**: Extensive test coverage with pytest
- **Type Safety**: Full type hints and Pydantic models

## Installation

```bash
pip install gofeatureflag-python-provider
```

Or with Poetry:

```bash
poetry add gofeatureflag-python-provider
```

## Quick Start

```python
import asyncio
from openfeature import api
from gofeatureflag_python_provider import GoFeatureFlagProvider, GoFeatureFlagProviderOptions, EvaluationType

async def main():
    # Create provider options
    options = GoFeatureFlagProviderOptions(
        endpoint="https://your-go-feature-flag-relay.com",
        evaluation_type=EvaluationType.REMOTE,
        timeout=5000
    )

    # Create and initialize the provider
    provider = GoFeatureFlagProvider(options)
    await provider.initialize()

    # Set the provider in OpenFeature
    api.set_provider(provider)

    # Use the client
    client = api.get_client()
    result = await client.get_boolean_details("my-feature-flag", False)
    print(f"Flag value: {result.value}")

asyncio.run(main())
```

## Configuration Options

### GoFeatureFlagProviderOptions

| Option                            | Type               | Default      | Description                                   |
| --------------------------------- | ------------------ | ------------ | --------------------------------------------- |
| `endpoint`                        | `str`              | **Required** | The endpoint of the GoFeatureFlag relay-proxy |
| `evaluation_type`                 | `EvaluationType`   | `IN_PROCESS` | Type of evaluation (Remote or InProcess)      |
| `timeout`                         | `int`              | `10000`      | HTTP request timeout in milliseconds          |
| `flag_change_polling_interval_ms` | `int`              | `120000`     | Flag configuration polling interval           |
| `data_flush_interval`             | `int`              | `120000`     | Data collection flush interval                |
| `max_pending_events`              | `int`              | `10000`      | Maximum pending events before flushing        |
| `disable_data_collection`         | `bool`             | `False`      | Whether to disable data collection            |
| `api_key`                         | `str`              | `None`       | API key for authentication                    |
| `exporter_metadata`               | `ExporterMetadata` | `None`       | Metadata for the exporter                     |

### Evaluation Types

- **`EvaluationType.REMOTE`**: Uses the OFREP protocol for remote evaluation
- **`EvaluationType.IN_PROCESS`**: Uses WASM for local evaluation (coming soon)

## Usage Examples

### Remote Evaluation

```python
from gofeatureflag_python_provider import GoFeatureFlagProvider, GoFeatureFlagProviderOptions, EvaluationType

options = GoFeatureFlagProviderOptions(
    endpoint="https://your-relay.com",
    evaluation_type=EvaluationType.REMOTE,
    timeout=5000,
    api_key="your-api-key"
)

provider = GoFeatureFlagProvider(options)
```

### In-Process Evaluation (Coming Soon)

```python
options = GoFeatureFlagProviderOptions(
    endpoint="https://your-relay.com",
    evaluation_type=EvaluationType.IN_PROCESS,
    flag_change_polling_interval_ms=60000
)

provider = GoFeatureFlagProvider(options)
```

### Custom Tracking

```python
# Track custom events
provider.track("user_action", context, {"action": "button_click"})
```

## Development

### Prerequisites

- Python 3.9+
- Poetry
- GoFeatureFlag relay-proxy

### Setup

1. Clone the repository
2. Install dependencies: `poetry install`
3. Run tests: `poetry run pytest`
4. Format code: `poetry run black .`

### Testing

```bash
# Run all tests
poetry run pytest

# Run with coverage
poetry run pytest --cov=gofeatureflag_python_provider

# Run specific test file
poetry run pytest tests/test_provider.py -v
```

## Architecture

The provider follows the OpenFeature specification and includes:

- **Provider**: Main provider class implementing OpenFeature interfaces
- **Evaluators**: Strategy pattern for different evaluation types
- **Models**: Pydantic models for configuration and responses
- **WASM Integration**: WebAssembly support for local evaluation
- **OFREP Integration**: Remote evaluation using the OFREP protocol

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the test suite
6. Submit a pull request

## License

Apache 2.0

## Support

- [GoFeatureFlag Documentation](https://gofeatureflag.org/)
- [OpenFeature Documentation](https://openfeature.dev/)
- [Issues](https://github.com/thomaspoignant/go-feature-flag/issues)
