"""
Tests for InProcessEvaluator: flag config fetch, storage, polling, and flag resolution.
The EvaluateWasm instance is mocked so tests run without the real WASM binary.
"""

from __future__ import annotations

import time
from unittest.mock import Mock, patch

import pytest

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import (
    ErrorCode,
    FlagNotFoundError,
    GeneralError,
    ParseError,
    ProviderFatalError,
    ProviderNotReadyError,
    TypeMismatchError,
)
from openfeature.flag_evaluation import FlagResolutionDetails

from gofeatureflag_python_provider.evaluator.inprocess_evaluator import (
    InProcessEvaluator,
)
from gofeatureflag_python_provider.exceptions import UnauthorizedError
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model import FlagConfigResponse
from gofeatureflag_python_provider.wasm import (
    WasmEvaluationTrapError,
    WasmNotLoadedError,
)
from gofeatureflag_python_provider.wasm.model import WasmEvaluationResponse

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _make_options(
    endpoint: str = "http://localhost:1031",
    flag_config_poll_interval_seconds: int | None = 10,
):
    return GoFeatureFlagOptions(
        endpoint=endpoint,
        flag_config_poll_interval_seconds=flag_config_poll_interval_seconds,
    )


def _make_evaluator_with_mock_wasm(options=None, mock_api=None):
    """
    Build an InProcessEvaluator whose EvaluateWasm is replaced by a Mock,
    so that WASM initialization and evaluation are fully controlled.
    Returns (evaluator, mock_wasm).
    """
    if options is None:
        options = _make_options()
    if mock_api is None:
        mock_api = Mock()
        mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
            etag="etag-1",
            flags={},
            evaluation_context_enrichment={},
        )

    evaluator = InProcessEvaluator(options, mock_api)
    mock_wasm = Mock()
    evaluator._wasm = mock_wasm
    return evaluator, mock_wasm


_FLAG_KEY = "my-flag"
_BOOL_FLAG_DICT = {
    "variations": {"on": True, "off": False},
    "defaultRule": {"variation": "on"},
    "trackEvents": True,
}
_DEFAULT_CTX = EvaluationContext(targeting_key="user-123", attributes={"role": "admin"})


# ---------------------------------------------------------------------------
# Lifecycle: initialize / shutdown / polling
# ---------------------------------------------------------------------------


def test_initialize_calls_wasm_initialize_and_fetches_flags():
    """initialize() calls wasm.initialize() and retrieve_flag_configuration."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
        etag="etag-1",
        flags={"flag_a": {"defaultValue": True}, "flag_b": {}},
        evaluation_context_enrichment={"env": "test"},
    )
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    evaluator.initialize()

    mock_wasm.initialize.assert_called_once()
    mock_api.retrieve_flag_configuration.assert_called_once()
    with evaluator._lock:
        assert evaluator._flags == {"flag_a": {"defaultValue": True}, "flag_b": {}}
        assert evaluator._etag == "etag-1"
        assert evaluator._evaluation_context_enrichment == {"env": "test"}
    evaluator.shutdown()


def test_polling_calls_retrieve_with_etag():
    """After initialize, _refresh_flag_configuration calls retrieve with stored etag."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag="first-etag", flags={"f": {}}),
        FlagConfigResponse(etag="second-etag", flags={"f": {"updated": True}}),
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()

    assert mock_api.retrieve_flag_configuration.call_count == 1
    evaluator._refresh_flag_configuration()
    assert mock_api.retrieve_flag_configuration.call_count == 2
    second_call = mock_api.retrieve_flag_configuration.call_args_list[1]
    assert second_call.kwargs.get("etag") == "first-etag"
    with evaluator._lock:
        assert evaluator._flags == {"f": {"updated": True}}
        assert evaluator._etag == "second-etag"
    evaluator.shutdown()


def test_shutdown_stops_polling_and_disposes_wasm():
    """shutdown() stops the poll thread and calls wasm.dispose()."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
        etag="e", flags={}
    )
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(
        options=_make_options(flag_config_poll_interval_seconds=1),
        mock_api=mock_api,
    )
    evaluator.initialize()
    call_count_after_init = mock_api.retrieve_flag_configuration.call_count
    dispose_count_after_init = mock_wasm.dispose.call_count
    evaluator.shutdown()
    time.sleep(0.1)
    assert mock_api.retrieve_flag_configuration.call_count == call_count_after_init
    assert evaluator._poll_thread is None
    assert evaluator._poll_stopper is None
    # initialize() also disposes, to release any pool a previous call left behind.
    assert mock_wasm.dispose.call_count == dispose_count_after_init + 1


def test_304_keeps_existing_flags():
    """A 'not modified' answer leaves the stored configuration untouched."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(
            etag="v1",
            flags={"my_flag": {"defaultValue": False}},
            evaluation_context_enrichment={},
        ),
        None,  # 304 Not Modified
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()
    with evaluator._lock:
        assert evaluator._flags == {"my_flag": {"defaultValue": False}}
    evaluator._refresh_flag_configuration()
    with evaluator._lock:
        assert evaluator._flags == {"my_flag": {"defaultValue": False}}
    evaluator.shutdown()


def test_an_emptied_flag_map_is_applied_and_flags_report_as_not_found():
    """Every flag being removed is a configuration change like any other.

    The provider is ready and its configuration is current, so the answer to a
    lookup is FLAG_NOT_FOUND — not PROVIDER_NOT_READY, which would blame the
    infrastructure for an empty config.
    """
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={_FLAG_KEY: _BOOL_FLAG_DICT}),
        FlagConfigResponse(etag='"v2"', flags={}),
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()

    with evaluator._lock:
        assert evaluator._flags == {}
        assert evaluator._etag == '"v2"'
    with pytest.raises(FlagNotFoundError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_poll_error_keeps_previous_state():
    """When a refresh raises, flags and etag remain unchanged."""
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag="e1", flags={"x": {}}),
        FlagConfigurationUnavailableError("network error"),
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()
    with evaluator._lock:
        assert evaluator._flags == {"x": {}}
        assert evaluator._etag == "e1"
    evaluator._refresh_flag_configuration()
    with evaluator._lock:
        assert evaluator._flags == {"x": {}}
        assert evaluator._etag == "e1"
    evaluator.shutdown()


def test_initialize_raises_on_first_fetch_failure():
    """When first retrieve_flag_configuration raises, initialize reports the error.

    The service-level exception is translated into the SDK's taxonomy so the
    provider registry can read an error code off it, but it stays reachable as
    the cause.
    """
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    cause = FlagConfigurationUnavailableError("endpoint not found")
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = cause
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(GeneralError, match="endpoint not found") as exc_info:
        evaluator.initialize()
    assert exc_info.value.__cause__ is cause

    with evaluator._lock:
        # Never loaded, which is distinct from "loaded and empty".
        assert evaluator._flags is None
        assert evaluator._etag is None


def test_initialize_releases_the_engine_when_the_first_fetch_fails():
    """A failed initialize() must not leave a WASM pool behind.

    The engine is loaded before the configuration is fetched, and the SDK never
    calls shutdown() on a provider whose initialize() raised, so a pool left
    loaded here would outlive the provider that owns it.
    """
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = (
        FlagConfigurationUnavailableError("relay proxy unreachable")
    )
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(GeneralError):
        evaluator.initialize()

    # Once on entry to clear any previous pool, once on the failure path.
    assert mock_wasm.dispose.call_count == 2
    assert evaluator._poll_thread is None


def test_initialize_releases_the_engine_on_a_fatal_failure():
    """The fatal path releases the engine too, not just the recoverable one."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = UnauthorizedError("bad api key")
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(ProviderFatalError):
        evaluator.initialize()

    assert mock_wasm.dispose.call_count == 2


def test_initialize_releases_the_engine_when_the_engine_itself_fails():
    """A pool that cannot be built is reported through the SDK taxonomy."""
    mock_api = Mock()
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    mock_wasm.initialize.side_effect = WasmNotLoadedError("WASI binary not found")

    with pytest.raises(ProviderFatalError, match="WASI binary not found"):
        evaluator.initialize()

    assert mock_wasm.dispose.call_count == 2
    mock_api.retrieve_flag_configuration.assert_not_called()


def test_initialize_304_is_recoverable_and_releases_the_engine():
    """An unconditional request answered 304 is a protocol violation, not a fatal one."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = None
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(GeneralError, match="304 Not Modified") as exc_info:
        evaluator.initialize()
    # Recoverable: the registry must land on ERROR, not FATAL.
    assert not isinstance(exc_info.value, ProviderFatalError)

    assert mock_wasm.dispose.call_count == 2
    with evaluator._lock:
        assert evaluator._flags is None


# ---------------------------------------------------------------------------
# Flag resolution: helper
# ---------------------------------------------------------------------------


def _setup_evaluator_with_flag(flag_dict: dict, enrichment: dict | None = None):
    """
    Build a fully-initialized evaluator with one flag in its local store
    and a mock WASM evaluator.
    """
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm()
    evaluator.initialize()
    with evaluator._lock:
        evaluator._flags = {_FLAG_KEY: flag_dict}
        evaluator._evaluation_context_enrichment = enrichment or {}
    return evaluator, mock_wasm


def _wasm_ok_response(**kwargs) -> WasmEvaluationResponse:
    """Build a successful WasmEvaluationResponse."""
    defaults = {
        "errorCode": "",
        "trackEvents": True,
        "reason": "TARGETING_MATCH",
        "variationType": "on",
        "value": True,
    }
    defaults.update(kwargs)
    return WasmEvaluationResponse(**defaults)


# ---------------------------------------------------------------------------
# Flag not found
# ---------------------------------------------------------------------------


def test_resolve_boolean_flag_not_in_store_raises():
    """FlagNotFoundError is raised when the flag key is absent from _flags."""
    evaluator, _ = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    evaluator._flags = {}  # empty
    with pytest.raises(FlagNotFoundError):
        evaluator.resolve_boolean_details("missing-flag", False, _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Boolean resolution
# ---------------------------------------------------------------------------


def test_resolve_boolean_details_returns_true(capfd):
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=True, variationType="on")

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert result.value is True
    assert result.variant == "on"
    assert result.reason == "TARGETING_MATCH"
    evaluator.shutdown()


def test_resolve_boolean_passes_wasm_input_correctly():
    """evaluate() is called with a WasmInput whose flagKey and evalContext match."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=False, variationType="off"
    )

    evaluator.resolve_boolean_details(_FLAG_KEY, True, _DEFAULT_CTX)

    call_args = mock_wasm.evaluate.call_args[0][0]
    assert call_args.flagKey == _FLAG_KEY
    assert call_args.evalContext["targetingKey"] == "user-123"
    assert call_args.evalContext["role"] == "admin"
    assert call_args.flagContext.defaultSdkValue is True
    evaluator.shutdown()


def test_resolve_boolean_type_mismatch_raises():
    """TypeMismatchError is raised when WASM returns a non-bool value for a bool flag."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value="not-a-bool", variationType="x"
    )

    with pytest.raises(TypeMismatchError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# String resolution
# ---------------------------------------------------------------------------


def test_resolve_string_details_returns_value():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v1": "red", "v2": "blue"}, "defaultRule": {"variation": "v1"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value="red", variationType="v1")

    result = evaluator.resolve_string_details(_FLAG_KEY, "default", _DEFAULT_CTX)

    assert result.value == "red"
    assert result.variant == "v1"
    evaluator.shutdown()


def test_resolve_string_type_mismatch_raises():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v": 42}, "defaultRule": {"variation": "v"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=42, variationType="v")

    with pytest.raises(TypeMismatchError):
        evaluator.resolve_string_details(_FLAG_KEY, "default", _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Integer resolution
# ---------------------------------------------------------------------------


def test_resolve_integer_details_returns_value():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v1": 100}, "defaultRule": {"variation": "v1"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=100, variationType="v1")

    result = evaluator.resolve_integer_details(_FLAG_KEY, 0, _DEFAULT_CTX)

    assert result.value == 100
    evaluator.shutdown()


def test_resolve_integer_type_mismatch_raises():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v": "oops"}, "defaultRule": {"variation": "v"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value="oops", variationType="v")

    with pytest.raises(TypeMismatchError):
        evaluator.resolve_integer_details(_FLAG_KEY, 0, _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Float resolution
# ---------------------------------------------------------------------------


def test_resolve_float_details_returns_float():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v1": 1.5}, "defaultRule": {"variation": "v1"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=1.5, variationType="v1")

    result = evaluator.resolve_float_details(_FLAG_KEY, 0.0, _DEFAULT_CTX)

    assert result.value == pytest.approx(1.5)
    evaluator.shutdown()


def test_resolve_float_accepts_integer_value():
    """A WASM int value is acceptable when resolving a float flag."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v": 2}, "defaultRule": {"variation": "v"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=2, variationType="v")

    result = evaluator.resolve_float_details(_FLAG_KEY, 0.0, _DEFAULT_CTX)
    assert result.value == 2
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Object resolution
# ---------------------------------------------------------------------------


def test_resolve_object_details_returns_dict():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v1": {"a": 1}}, "defaultRule": {"variation": "v1"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value={"a": 1}, variationType="v1"
    )

    result = evaluator.resolve_object_details(_FLAG_KEY, {}, _DEFAULT_CTX)

    assert result.value == {"a": 1}
    evaluator.shutdown()


def test_resolve_object_details_returns_list():
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        {"variations": {"v1": [1, 2, 3]}, "defaultRule": {"variation": "v1"}}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=[1, 2, 3], variationType="v1"
    )

    result = evaluator.resolve_object_details(_FLAG_KEY, [], _DEFAULT_CTX)

    assert result.value == [1, 2, 3]
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# WASM error-code mapping
# ---------------------------------------------------------------------------


def test_wasm_flag_not_found_error_raises_flag_not_found_error():
    """An errorCode of FLAG_NOT_FOUND from WASM maps to FlagNotFoundError."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = WasmEvaluationResponse(
        errorCode="FLAG_NOT_FOUND",
        errorDetails="flag 'my-flag' not found",
        reason="ERROR",
        trackEvents=False,
    )

    with pytest.raises(FlagNotFoundError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_wasm_type_mismatch_error_raises_type_mismatch_error():
    """An errorCode of TYPE_MISMATCH from WASM maps to TypeMismatchError."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = WasmEvaluationResponse(
        errorCode="TYPE_MISMATCH",
        errorDetails="type mismatch",
        reason="ERROR",
        trackEvents=False,
    )

    with pytest.raises(TypeMismatchError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_wasm_general_error_raises_general_error():
    """An unknown errorCode from WASM maps to GeneralError."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = WasmEvaluationResponse(
        errorCode="GENERAL",
        errorDetails="something went wrong",
        reason="ERROR",
        trackEvents=False,
    )

    with pytest.raises(GeneralError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Default value fallback
# ---------------------------------------------------------------------------


def test_resolve_returns_default_when_wasm_value_is_none():
    """If WASM returns None value (and no error), the SDK default is used."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = WasmEvaluationResponse(
        value=None,
        errorCode="",
        reason="DEFAULT",
        variationType="SdkDefault",
        trackEvents=False,
    )

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    assert result.value is False
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Metadata forwarding
# ---------------------------------------------------------------------------


def test_resolve_includes_flag_metadata():
    """Metadata returned by WASM is forwarded in FlagResolutionDetails."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = WasmEvaluationResponse(
        value=True,
        variationType="on",
        reason="TARGETING_MATCH",
        errorCode="",
        trackEvents=True,
        metadata={"experiment": "group-a", "version": "2"},
    )

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    assert result.flag_metadata == {"experiment": "group-a", "version": "2"}
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Evaluation context enrichment forwarded to WASM
# ---------------------------------------------------------------------------


def test_evaluation_context_enrichment_is_passed_to_wasm():
    """enrichment stored on the evaluator is forwarded to the WASM flagContext."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(
        _BOOL_FLAG_DICT, enrichment={"region": "eu-west"}
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=True)

    evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    call_args = mock_wasm.evaluate.call_args[0][0]
    assert call_args.flagContext.evaluationContextEnrichment == {"region": "eu-west"}
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Async wrappers delegate to sync counterparts
# ---------------------------------------------------------------------------


def test_resolve_boolean_details_async_delegates_to_sync():
    """Async resolve method runs sync evaluation via asyncio.to_thread."""
    import asyncio

    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=True)

    result = asyncio.run(
        evaluator.resolve_boolean_details_async(_FLAG_KEY, False, _DEFAULT_CTX)
    )

    assert result.value is True
    mock_wasm.evaluate.assert_called_once()
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# is_flag_trackable
# ---------------------------------------------------------------------------


def test_is_flag_trackable_true_when_flag_has_track_events_true():
    evaluator, _ = _setup_evaluator_with_flag(
        {"trackEvents": True, "defaultRule": {"variation": "v"}}
    )
    assert evaluator.is_flag_trackable(_FLAG_KEY) is True
    evaluator.shutdown()


def test_is_flag_trackable_false_when_flag_has_track_events_false():
    evaluator, _ = _setup_evaluator_with_flag(
        {"trackEvents": False, "defaultRule": {"variation": "v"}}
    )
    assert evaluator.is_flag_trackable(_FLAG_KEY) is False
    evaluator.shutdown()


def test_is_flag_trackable_returns_true_for_unknown_flag():
    """When the flag is not in _flags, trackable defaults to True."""
    evaluator, _ = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    assert evaluator.is_flag_trackable("unknown-flag") is True
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# WASM runtime failures degrade to GeneralError (issue #5651)
# ---------------------------------------------------------------------------


def test_wasm_trap_degrades_to_general_error():
    """
    A WASM trap surfaces as GeneralError so the OpenFeature SDK returns the
    default value with reason ERROR instead of a raw wasmtime exception.
    """
    from gofeatureflag_python_provider.wasm import WasmEvaluationTrapError

    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.side_effect = WasmEvaluationTrapError("stack overflow trap")

    with pytest.raises(GeneralError, match="WASM evaluation failed"):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_wasm_serialization_failure_degrades_to_general_error():
    """
    An input the host cannot serialize for the module (pydantic's ~255-level
    nesting cap, circular references) surfaces as GeneralError, not a raw
    pydantic exception.
    """
    from pydantic_core import PydanticSerializationError

    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.side_effect = PydanticSerializationError("depth exceeded")

    with pytest.raises(GeneralError, match="WASM evaluation failed"):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_wasm_pool_timeout_degrades_to_general_error():
    """An exhausted, unhealable pool surfaces as GeneralError, not a hang/raw error."""
    from gofeatureflag_python_provider.wasm import WasmPoolTimeoutError

    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.side_effect = WasmPoolTimeoutError("no slot available")

    with pytest.raises(GeneralError, match="WASM evaluation failed"):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# A boolean must not satisfy the integer or float resolver
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "resolve_attr,default_value",
    [
        ("resolve_integer_details", 0),
        ("resolve_float_details", 0.0),
    ],
)
def test_boolean_value_does_not_satisfy_numeric_resolver(resolve_attr, default_value):
    """Python makes bool a subclass of int, so isinstance(True, int) is True.

    Without an explicit guard a boolean flag silently satisfies the integer and
    float resolvers and is returned as a number.
    """
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=True, variationType="on")

    with pytest.raises(TypeMismatchError):
        getattr(evaluator, resolve_attr)(_FLAG_KEY, default_value, _DEFAULT_CTX)
    evaluator.shutdown()


def test_integral_number_still_satisfies_the_float_resolver():
    """JSON does not distinguish 100 from 100.0, so an integral number is a float."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=101, variationType="medium"
    )

    result = evaluator.resolve_float_details(_FLAG_KEY, 0.0, _DEFAULT_CTX)

    assert result.value == 101
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Readiness, re-initialization and refresh safety
# ---------------------------------------------------------------------------


def test_resolve_before_configuration_loaded_reports_not_ready():
    """Before the first successful load, evaluation must not blame the flag key.

    Reporting FLAG_NOT_FOUND here misattributes an infrastructure failure to the
    caller.
    """
    evaluator, _ = _make_evaluator_with_mock_wasm()

    with pytest.raises(ProviderNotReadyError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)


def test_resolve_after_shutdown_reports_not_ready():
    """Shutdown returns the evaluator to the not-loaded state, not to 'empty'."""
    evaluator, _ = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    evaluator.shutdown()

    with pytest.raises(ProviderNotReadyError):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)


def test_304_does_not_advance_the_stored_etag():
    """A 'not modified' answer must leave flags, enrichment and the ETag alone."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(
            etag='"v1"',
            flags={"f": {"defaultValue": True}},
            evaluation_context_enrichment={"env": "prod"},
        ),
        None,  # 304 Not Modified
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()

    with evaluator._lock:
        assert evaluator._flags == {"f": {"defaultValue": True}}
        assert evaluator._etag == '"v1"'
        assert evaluator._evaluation_context_enrichment == {"env": "prod"}
    evaluator.shutdown()


def test_a_rejected_response_does_not_advance_the_stored_etag():
    """A refresh the API rejected (e.g. a body with no flag map) changes nothing.

    Advancing the ETag here pins the provider to a configuration it never
    received, after which the server answers 304 forever and the stale
    configuration becomes permanent.
    """
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {"defaultValue": True}}),
        FlagConfigurationUnavailableError(
            "flag configuration response does not contain a flag map"
        ),
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()

    with evaluator._lock:
        assert evaluator._flags == {"f": {"defaultValue": True}}
        assert evaluator._etag == '"v1"'
    evaluator.shutdown()


def test_initialize_twice_leaves_a_single_live_poller():
    """Re-initialization must cancel the previous poller rather than orphan it."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
        etag='"v1"', flags={"f": {}}
    )
    evaluator, mock_wasm = _make_evaluator_with_mock_wasm(
        options=_make_options(flag_config_poll_interval_seconds=1),
        mock_api=mock_api,
    )

    evaluator.initialize()
    first_thread = evaluator._poll_thread
    evaluator.initialize()
    second_thread = evaluator._poll_thread

    assert first_thread is not second_thread
    assert not first_thread.is_alive()
    assert second_thread.is_alive()
    # The previous pool is released before a new one is built.
    assert mock_wasm.dispose.call_count == 2
    assert mock_wasm.initialize.call_count == 2
    evaluator.shutdown()


def test_initialize_unauthorized_is_fatal():
    """401/403 cannot be repaired by retrying, so it must not be a retryable error."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = UnauthorizedError("bad api key")
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(ProviderFatalError):
        evaluator.initialize()


def test_initialize_other_failure_is_not_fatal():
    """Every non-auth initialization failure must stay recoverable.

    Recoverable means GeneralError rather than ProviderFatalError: the registry
    reads the error code off the exception to choose between ERROR and FATAL,
    and only FATAL stops the SDK from evaluating against this provider again.
    """
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = (
        FlagConfigurationUnavailableError("relay proxy unreachable")
    )
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(GeneralError) as exc_info:
        evaluator.initialize()
    assert not isinstance(exc_info.value, ProviderFatalError)
    assert exc_info.value.error_code == ErrorCode.GENERAL


# ---------------------------------------------------------------------------
# Trackability
# ---------------------------------------------------------------------------


def test_flag_omitting_track_events_is_trackable():
    """The engine's own default is true, so an omitted trackEvents means trackable.

    Defaulting to false silently drops every event for such a flag.
    """
    evaluator, _ = _setup_evaluator_with_flag({"variations": {"on": True}})

    assert evaluator.is_flag_trackable(_FLAG_KEY) is True
    evaluator.shutdown()


def test_flag_disabling_track_events_is_not_trackable():
    evaluator, _ = _setup_evaluator_with_flag(
        {"variations": {"on": True}, "trackEvents": False}
    )

    assert evaluator.is_flag_trackable(_FLAG_KEY) is False
    evaluator.shutdown()


def test_flag_absent_from_configuration_is_trackable():
    """So that a flag added between polls still produces data."""
    evaluator, _ = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)

    assert evaluator.is_flag_trackable("never-heard-of-it") is True
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Provider events
# ---------------------------------------------------------------------------


def _evaluator_with_event_spies(mock_api, **options_kwargs):
    """Build an evaluator whose event callbacks record their calls."""
    events = {"changed": [], "stale": [], "ready": []}
    options = _make_options(**options_kwargs)
    evaluator = InProcessEvaluator(
        options,
        mock_api,
        on_configuration_changed=lambda flags: events["changed"].append(flags),
        on_stale=lambda message: events["stale"].append(message),
        on_ready=lambda: events["ready"].append(True),
    )
    evaluator._wasm = Mock()
    return evaluator, events


def test_initial_load_does_not_emit_configuration_changed():
    """Consumers must not observe a change event before the provider is ready."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
        etag='"v1"', flags={"f": {"defaultValue": True}}
    )
    evaluator, events = _evaluator_with_event_spies(mock_api)

    evaluator.initialize()

    assert events["changed"] == []
    evaluator.shutdown()


def test_configuration_changed_is_emitted_only_when_content_differs():
    """A fresh response is not the same thing as a different one."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {"defaultValue": True}}),
        # Same content, new ETag: fetched, but not changed.
        FlagConfigResponse(etag='"v2"', flags={"f": {"defaultValue": True}}),
        # Genuinely different.
        FlagConfigResponse(etag='"v3"', flags={"f": {"defaultValue": False}}),
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()
    assert events["changed"] == []

    evaluator._refresh_flag_configuration()
    assert events["changed"] == [["f"]]
    evaluator.shutdown()


def test_304_does_not_emit_configuration_changed():
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {}}),
        None,
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()

    assert events["changed"] == []
    evaluator.shutdown()


def test_enrichment_change_alone_is_a_configuration_change():
    """Enrichment feeds every evaluation, so changing it changes behaviour."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(
            etag='"v1"', flags={"f": {}}, evaluation_context_enrichment={"env": "dev"}
        ),
        FlagConfigResponse(
            etag='"v2"', flags={"f": {}}, evaluation_context_enrichment={"env": "prod"}
        ),
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()

    assert events["changed"] == [["evaluationContextEnrichment"]]
    evaluator.shutdown()


def test_stale_after_three_consecutive_failures_then_ready_on_recovery():
    """Stale is reported at the third failure, once, and cleared on recovery."""
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {"defaultValue": True}}),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigResponse(etag='"v2"', flags={"f": {"defaultValue": False}}),
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()

    evaluator._refresh_flag_configuration()
    evaluator._refresh_flag_configuration()
    assert events["stale"] == []

    evaluator._refresh_flag_configuration()
    assert len(events["stale"]) == 1

    # A fourth failure must not emit a second stale event.
    evaluator._refresh_flag_configuration()
    assert len(events["stale"]) == 1
    # The last known-good configuration keeps serving throughout.
    with evaluator._lock:
        assert evaluator._flags == {"f": {"defaultValue": True}}

    evaluator._refresh_flag_configuration()
    assert events["ready"] == [True]
    evaluator.shutdown()


def test_an_empty_flag_map_is_a_successful_refresh():
    """A relay proxy serving no flags is a valid configuration, not a failure.

    Counting it as one would drive a correctly-configured provider to STALE and
    keep it there: the ETag never advances, so no later poll can clear it.
    """
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {}}),
        FlagConfigResponse(etag='"v2"', flags={}),
        FlagConfigResponse(etag='"v3"', flags={}),
        FlagConfigResponse(etag='"v4"', flags={}),
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()

    for _ in range(3):
        evaluator._refresh_flag_configuration()

    assert events["stale"] == []
    with evaluator._lock:
        assert evaluator._flags == {}
        assert evaluator._etag == '"v4"'
    evaluator.shutdown()


def test_304_clears_a_stale_state():
    """The server answered, so the configuration in hand is current."""
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"f": {}}),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigurationUnavailableError("boom"),
        FlagConfigurationUnavailableError("boom"),
        None,
    ]
    evaluator, events = _evaluator_with_event_spies(mock_api)
    evaluator.initialize()
    for _ in range(3):
        evaluator._refresh_flag_configuration()
    assert len(events["stale"]) == 1

    evaluator._refresh_flag_configuration()

    assert events["ready"] == [True]
    assert events["changed"] == []
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Remote fallback
# ---------------------------------------------------------------------------


def _evaluator_with_fallback(flag_dict=None, remote_result=None, remote_error=None):
    """Evaluator with a stubbed OFREP client standing in for the relay proxy."""
    evaluator, mock_wasm = _setup_evaluator_with_flag(flag_dict or _BOOL_FLAG_DICT)
    fallback = Mock()
    if remote_error is not None:
        fallback.resolve_boolean_details.side_effect = remote_error
    else:
        fallback.resolve_boolean_details.return_value = (
            remote_result
            if remote_result is not None
            else FlagResolutionDetails(
                value=True,
                reason="TARGETING_MATCH",
                variant="on",
                flag_metadata={"description": "from the relay proxy"},
            )
        )
    evaluator._ofrep_fallback = fallback
    return evaluator, mock_wasm, fallback


@pytest.mark.parametrize("engine_error_code", ["PARSE_ERROR", "GENERAL"])
def test_engine_failure_falls_back_to_remote_evaluation(engine_error_code):
    """The relay proxy is authoritative, so a provider-side failure re-evaluates there."""
    evaluator, mock_wasm, fallback = _evaluator_with_fallback()
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode=engine_error_code, errorDetails="boom"
    )

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    fallback.resolve_boolean_details.assert_called_once_with(
        _FLAG_KEY, False, _DEFAULT_CTX
    )
    assert result.value is True
    assert result.variant == "on"
    # Marked so the failure is diagnosable rather than invisible, and so the
    # data collector knows not to count it twice.
    assert result.flag_metadata["gofeatureflag_evaluated_remotely"] is True
    assert result.flag_metadata["description"] == "from the relay proxy"
    evaluator.shutdown()


@pytest.mark.parametrize(
    "engine_error_code,expected_error",
    [
        # A deterministic misconfiguration the relay proxy would reproduce
        # identically, so a fallback would only add latency.
        ("FLAG_CONFIG", GeneralError),
        ("FLAG_NOT_FOUND", FlagNotFoundError),
        ("TYPE_MISMATCH", TypeMismatchError),
    ],
)
def test_non_qualifying_engine_errors_do_not_fall_back(
    engine_error_code, expected_error
):
    evaluator, mock_wasm, fallback = _evaluator_with_fallback()
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode=engine_error_code, errorDetails="nope"
    )

    with pytest.raises(expected_error):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    fallback.resolve_boolean_details.assert_not_called()
    evaluator.shutdown()


def test_wasm_trap_falls_back_without_retrying_locally():
    """The host already discarded and rebuilt the trapped instance.

    Retrying locally on the fresh instance before falling back would just burn
    another evaluation on the same bad input.
    """
    evaluator, mock_wasm, fallback = _evaluator_with_fallback()
    mock_wasm.evaluate.side_effect = WasmEvaluationTrapError("stack overflow")

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert mock_wasm.evaluate.call_count == 1
    fallback.resolve_boolean_details.assert_called_once()
    assert result.flag_metadata["gofeatureflag_evaluated_remotely"] is True
    evaluator.shutdown()


def test_fallback_happens_on_every_qualifying_occurrence():
    evaluator, mock_wasm, fallback = _evaluator_with_fallback()
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode="PARSE_ERROR", errorDetails="boom"
    )

    for _ in range(3):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert fallback.resolve_boolean_details.call_count == 3
    evaluator.shutdown()


def test_remote_failure_surfaces_the_original_in_process_error():
    """The in-process error is the root cause, so it is what the caller sees."""
    evaluator, mock_wasm, _ = _evaluator_with_fallback(
        remote_error=RuntimeError("relay proxy unreachable")
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode="PARSE_ERROR", errorDetails="malformed flag"
    )

    with pytest.raises(ParseError, match="malformed flag"):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)
    evaluator.shutdown()


def test_engine_parse_error_maps_to_parse_error():
    """PARSE_ERROR has its own SDK exception; collapsing it into GeneralError
    would report the wrong error code to the caller."""
    evaluator, mock_wasm, _ = _evaluator_with_fallback(
        remote_error=RuntimeError("relay proxy unreachable")
    )
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode="PARSE_ERROR", errorDetails="boom"
    )

    with pytest.raises(ParseError) as raised:
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert raised.value.error_code == ErrorCode.PARSE_ERROR
    evaluator.shutdown()


def test_each_fallback_is_logged_at_warning_level(caplog):
    evaluator, mock_wasm, _ = _evaluator_with_fallback()
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=None, errorCode="PARSE_ERROR", errorDetails="boom"
    )

    with caplog.at_level("WARNING"):
        evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert any(
        record.levelname == "WARNING" and "falling back to remote" in record.message
        for record in caplog.records
    )
    evaluator.shutdown()


def test_fallback_client_is_built_from_the_same_options():
    """Authentication and timeout must apply to the fallback identically."""
    evaluator, _ = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)

    with patch(
        "gofeatureflag_python_provider.evaluator.inprocess_evaluator.build_ofrep_provider"
    ) as mock_build:
        first = evaluator._fallback_provider()
        second = evaluator._fallback_provider()

    mock_build.assert_called_once_with(evaluator._options)
    assert first is second  # built once, then reused
    evaluator.shutdown()


# ---------------------------------------------------------------------------
# Flag version, flag list and poll jitter
# ---------------------------------------------------------------------------


def test_engine_version_is_surfaced_as_flag_metadata():
    """The engine reports version top-level, not inside its metadata block.

    A scheduled rollout step can override it at evaluation time, so the response
    field is authoritative -- reading it back out of the stored flag config would
    be wrong whenever such a step is active.
    """
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=True, variationType="on", version="1.7", metadata={"description": "x"}
    )

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert result.flag_metadata["version"] == "1.7"
    assert result.flag_metadata["description"] == "x"
    evaluator.shutdown()


def test_engine_version_never_overwrites_flag_defined_metadata():
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(
        value=True, version="1.7", metadata={"version": "set-by-the-flag"}
    )

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert result.flag_metadata["version"] == "set-by-the-flag"
    evaluator.shutdown()


def test_absent_engine_version_adds_no_metadata_key():
    evaluator, mock_wasm = _setup_evaluator_with_flag(_BOOL_FLAG_DICT)
    mock_wasm.evaluate.return_value = _wasm_ok_response(value=True, metadata={})

    result = evaluator.resolve_boolean_details(_FLAG_KEY, False, _DEFAULT_CTX)

    assert "version" not in result.flag_metadata
    evaluator.shutdown()


def test_evaluation_flag_list_is_sent_on_every_fetch():
    """A list applied only at startup would silently widen on the first poll."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(etag='"v1"', flags={"a": {}}),
        FlagConfigResponse(etag='"v2"', flags={"a": {"updated": True}}),
    ]
    options = GoFeatureFlagOptions(
        endpoint="http://localhost:1031",
        evaluation_flag_list=["a", "b"],
    )
    evaluator = InProcessEvaluator(options, mock_api)
    evaluator._wasm = Mock()

    evaluator.initialize()
    assert mock_api.retrieve_flag_configuration.call_args.kwargs["flags"] == ["a", "b"]

    evaluator._refresh_flag_configuration()
    assert mock_api.retrieve_flag_configuration.call_args.kwargs["flags"] == ["a", "b"]
    evaluator.shutdown()


def test_evaluation_flag_list_defaults_to_all_flags():
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.return_value = FlagConfigResponse(
        etag='"v1"', flags={"a": {}}
    )
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    evaluator.initialize()

    assert mock_api.retrieve_flag_configuration.call_args.kwargs["flags"] is None
    evaluator.shutdown()


def test_poll_delay_is_jittered_around_the_configured_interval():
    """Without jitter a fleet restarted together polls in lockstep indefinitely."""
    evaluator, _ = _make_evaluator_with_mock_wasm(
        options=_make_options(flag_config_poll_interval_seconds=100)
    )

    delays = {evaluator._next_poll_delay() for _ in range(200)}

    assert all(90.0 <= d <= 110.0 for d in delays)
    # Actually varying, not a constant dressed up as one.
    assert len(delays) > 100
    assert 95.0 < (sum(delays) / len(delays)) < 105.0
