# GO Feature Flag Python Provider

GO Feature Flag provider allows you to connect to your GO Feature Flag instance.

[GO Feature Flag](https://gofeatureflag.org) believes in simplicity and offers a simple and lightweight solution to use feature flags.  
Our focus is to avoid any complex infrastructure work to use GO Feature Flag.

This is a complete feature flagging solution with the possibility to target only a group of users, use any types of flags, store your configuration in various location and advanced rollout functionality. You can also collect usage data of your flags and be notified of configuration changes.

# Python SDK usage

## Install dependencies

The first things we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

```shell
pip install gofeatureflag-python-provider
```

## Evaluation modes

The provider supports two evaluation modes:

| Mode | Description |
|------|-------------|
| **In-Process** _(default)_ | Flag configuration is fetched and cached locally; evaluation runs via a WASM module — no per-evaluation network call. |
| **Remote** | Each flag evaluation makes an HTTP request to the GO Feature Flag relay proxy using the OFREP API. |

## Initialize your Open Feature client

### In-Process evaluation (default)

In-Process evaluation fetches the flag configuration from the relay proxy at startup and on a configurable polling interval. Flags are evaluated locally using a bundled WASM module, which gives you lower latency and no per-evaluation network dependency.

```python
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.options import GoFeatureFlagOptions, EvaluationType
from openfeature import api
from openfeature.evaluation_context import EvaluationContext

goff_provider = GoFeatureFlagProvider(
    options=GoFeatureFlagOptions(
        endpoint="https://gofeatureflag.org/",
        evaluation_type=EvaluationType.INPROCESS,  # default
    )
)
api.set_provider(goff_provider)
client = api.get_client(domain="test-client")
```

### Remote evaluation

Remote evaluation sends each flag evaluation as an HTTP request to the relay proxy.

```python
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.options import GoFeatureFlagOptions, EvaluationType
from openfeature import api

goff_provider = GoFeatureFlagProvider(
    options=GoFeatureFlagOptions(
        endpoint="https://gofeatureflag.org/",
        evaluation_type=EvaluationType.REMOTE,
    )
)
api.set_provider(goff_provider)
client = api.get_client(domain="test-client")
```

## Evaluate your flag

This code block explains how you can create an `EvaluationContext` and use it to evaluate your flag.

> In this example we are evaluating a `boolean` flag, but other types are available.
>
> **Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

```python
# Context of your flag evaluation.
# With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.
evaluation_ctx = EvaluationContext(
    targeting_key="d45e303a-38c2-11ed-a261-0242ac120002",
    attributes={
        "email": "john.doe@gofeatureflag.org",
        "firstname": "john",
        "lastname": "doe",
        "anonymous": False,
        "professional": True,
        "rate": 3.14,
        "age": 30,
        "company_info": {"name": "my_company", "size": 120},
        "labels": ["pro", "beta"],
    },
)

admin_flag = client.get_boolean_value(
    flag_key="flag-only-for-admin",
    default_value=False,
    evaluation_context=evaluation_ctx,
)

if admin_flag:
    # flag "flag-only-for-admin" is true for the user
    pass
else:
    # flag "flag-only-for-admin" is false for the user
    pass
```

## Configuration options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `endpoint` | `str` | _(required)_ | URL of the GO Feature Flag relay proxy |
| `evaluation_type` | `EvaluationType` | `INPROCESS` | Evaluation mode: `INPROCESS` or `REMOTE` |
| `cache_size` | `int` | `10000` | Maximum number of flag evaluations kept in the LRU cache _(remote mode)_ |
| `data_flush_interval` | `int` | `60000` | Interval (ms) to flush usage data to the relay proxy |
| `disable_data_collection` | `bool` | `False` | Set to `True` to disable usage analytics |
| `reconnect_interval` | `int` | `60` | WebSocket reconnect interval (seconds) _(remote mode)_ |
| `disable_cache_invalidation` | `bool` | `False` | Disable WebSocket-based cache invalidation _(remote mode)_ |
| `flag_config_poll_interval_seconds` | `int` | `10` | Polling interval (seconds) for flag configuration _(in-process mode)_ |
| `api_key` | `str` | `None` | API key for authenticated relay proxy requests |
| `exporter_metadata` | `dict` | `{}` | Static metadata attached to evaluation events |
| `max_pending_events` | `int` | `10000` | Maximum buffered events before a forced flush |
| `wasm_file_path` | `str` | `None` | Path to a custom WASM/WASI evaluation binary _(in-process mode, uses bundled binary by default)_ |
| `wasm_pool_size` | `int` | `10` | Pool size for concurrent WASM evaluation instances _(in-process mode)_ |
| `log_level` | `str\|int` | `"WARNING"` | Logging level (`"DEBUG"`, `"INFO"`, `"WARNING"`, `"ERROR"`) |
| `urllib3_pool_manager` | `urllib3.PoolManager` | `None` | Custom HTTP client |
