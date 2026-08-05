# GO Feature Flag Python Provider

OpenFeature Python provider for [GO Feature Flag](https://gofeatureflag.org).

## Project Overview

This package implements the OpenFeature provider interface for GO Feature Flag.
It supports two evaluation modes, selected with `options.evaluation_type`:

- **In-process** _(default)_: the provider polls the relay proxy for the flag
  configuration and evaluates flags locally through the embedded WASM
  evaluation engine (`wasmtime`). No network call per evaluation.
- **Remote**: every evaluation is delegated to the relay proxy through the
  OpenFeature Remote Evaluation Protocol (OFREP), via the
  `openfeature-provider-ofrep` package.

There is **no evaluation-result cache** in either mode (the former `pylru`
LRU cache and WebSocket invalidation were removed with the in-process
rewrite). The cache is an optional capability in the provider specification;
this provider does not implement it.

## Architecture

```
gofeatureflag_python_provider/
├── provider.py                 # GoFeatureFlagProvider (AbstractProvider impl), hook wiring, provider events
├── options.py                  # GoFeatureFlagOptions configuration
├── metadata.py                 # Provider metadata ("GO Feature Flag Provider")
├── exceptions.py               # Provider exceptions
├── utils.py                    # contextKind helpers etc.
├── evaluator/
│   ├── abstract_evaluator.py   # AbstractEvaluator interface (sync + async resolves, is_flag_trackable)
│   ├── inprocess_evaluator.py  # Polls flag configuration, evaluates via WASM, remote fallback via OFREP
│   ├── remote_evaluator.py     # Delegates every resolve to the OFREP provider; flags are never trackable
│   └── util.py
├── hooks/
│   ├── data_collector.py       # Collects FeatureEvents for local evaluations (after/error)
│   └── enrich_evaluation_context.py  # Adds the gofeatureflag metadata to the context before evaluation
├── services/
│   ├── api.py                  # HTTP client: flag configuration (ETag aware) + data collector
│   ├── event_publisher.py      # Buffers events, bulk-flushes on interval/threshold/shutdown
│   ├── ofrep.py                # Shared OFREP client construction (auth, headers, timeout)
│   └── model/                  # Pydantic request/response models
└── wasm/
    ├── evaluate_wasm.py        # WASM evaluation entry point, instance pool, trap handling
    ├── wasi_runtime.py         # wasmtime runtime setup
    ├── errors/                 # Trap/pool/loading error types
    ├── model/                  # WASM input/output models
    └── gofeatureflag-evaluation_<version>.wasi  # Bundled engine (pinned in _wasi_version.txt)
```

## Key Components

### GoFeatureFlagProvider (`provider.py`)
- Extends `AbstractProvider`; delegates all `resolve_*_details` (sync and
  async) to the evaluator chosen from `options.evaluation_type`.
- Registers two provider hooks, in order: `EnrichEvaluationContextHook`, then
  `DataCollectorHook` (the collector must observe the enriched context).
- Emits provider events (`configuration changed`, `stale`, `ready`) from the
  in-process polling loop.

### GoFeatureFlagOptions (`options.py`)
- `endpoint` (required): URL of the relay proxy
- `evaluation_type`: `INPROCESS` (default) or `REMOTE`
- `api_key`: sent as an `X-API-Key` header when set
- `timeout`: HTTP timeout in ms (default: 10000)
- `flag_config_poll_interval_seconds`: in-process polling interval (default: 120)
- `evaluation_flag_list`: restrict fetched flag configuration (in-process only)
- `wasm_file_path`, `wasm_pool_size`: WASM binary override / instance pool size
  (default pool size: CPU core count)
- `data_flush_interval` (ms, default 60000), `max_pending_events`
  (default 10000), `disable_data_collection`, `exporter_metadata`,
  `data_collector_base_url`: data collection tuning
- `custom_headers`, `urllib3_pool_manager`, `log_level`

### DataCollectorHook (`hooks/data_collector.py`)
- Builds a `FeatureEvent` in `after()`/`error()` and enqueues it on the
  `EventPublisher`.
- Skips the event when data collection is disabled, when the evaluator reports
  the flag as not trackable, or when the result carries the
  `gofeatureflag_evaluated_remotely` metadata (the relay proxy already
  recorded it — emitting would double count).
- **Remote mode emits no feature events at all**: `RemoteEvaluator.
  is_flag_trackable()` always returns `False` because every remote evaluation
  is already recorded server-side. `source` is always `INPROCESS`.

### EventPublisher (`services/event_publisher.py`)
- Buffers events and bulk-posts them to `/v1/data/collector` on a periodic
  interval, when the buffer reaches `max_pending_events`, and on shutdown.
- Never posts an empty batch; failed batches are re-queued in order; the
  buffer is capped at twice `max_pending_events` (oldest dropped first).

## API Reference

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/v1/flag/configuration` | POST | Fetch flag configuration (in-process; ETag/If-None-Match aware) |
| `/ofrep/v1/evaluate/flags/{flag_key}` | POST | Remote evaluation via OFREP |
| `/v1/data/collector` | POST | Bulk usage/evaluation data |

## Development

### Prerequisites
- Python 3.9+
- uv (package manager)

### Setup and common commands
```bash
uv sync                  # Install dependencies
uv run pytest            # Run all tests
uv run pytest tests/test_inprocess_evaluator.py   # One file
uv run black gofeatureflag_python_provider tests  # Format
```

## Key Patterns

- **Pydantic models** for options and API payloads (`model_dump_json()` /
  `model_validate_json()`).
- **HTTP**: `urllib3.PoolManager` for flag configuration and data collection;
  the OFREP client (built in `services/ofrep.py`) handles remote evaluation.
  Both paths share auth/header/timeout construction.
- **WASM**: `wasmtime.Store` is not thread-safe — evaluation uses an instance
  pool; a trapped instance is poisoned and recycled, never reused.
- **Error handling**: evaluators raise OpenFeature exceptions
  (`FlagNotFoundError`, `TypeMismatchError`, `InvalidContextError`,
  `GeneralError`); the SDK converts them into default-value results.

## Tests

Table of the main test files (pytest):

| File | Covers |
|------|--------|
| `test_gofeatureflag_provider.py` | Provider end-to-end (remote mode) |
| `test_gofeatureflag_provider_inprocess.py` | Provider end-to-end (in-process) |
| `test_inprocess_evaluator.py` | In-process evaluator, polling, fallback |
| `test_remote_evaluator_ofrep.py` | Remote evaluator delegation to OFREP |
| `test_evaluate_wasm.py`, `test_wasm_trap_diagnosis.py` | WASM engine host, trap handling |
| `test_data_collector_hook.py`, `test_event_publisher.py` | Data collection |
| `test_options.py` | Options, including removal of the old cache options |
| `test_services_api.py` | HTTP layer (ETag, status mapping) |
