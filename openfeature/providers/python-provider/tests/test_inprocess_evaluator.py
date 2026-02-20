"""
Tests for InProcessEvaluator: flag config fetch, storage, polling, and flag resolution.
The EvaluateWasm instance is mocked so tests run without the real WASM binary.
"""

from __future__ import annotations

import time
from unittest.mock import Mock, patch

import pytest

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import FlagNotFoundError, GeneralError, TypeMismatchError

from gofeatureflag_python_provider.evaluator.inprocess_evaluator import (
    InProcessEvaluator,
)
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.models import FlagConfigResponse
from gofeatureflag_python_provider.wasm.models import WasmEvaluationResponse


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
    evaluator.shutdown()
    time.sleep(0.1)
    assert mock_api.retrieve_flag_configuration.call_count == call_count_after_init
    assert evaluator._poll_thread is None
    assert evaluator._poll_stopper is None
    mock_wasm.dispose.assert_called_once()


def test_304_keeps_existing_flags():
    """When refresh returns a response with empty flags (304-style), stored flags are unchanged."""
    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = [
        FlagConfigResponse(
            etag="v1",
            flags={"my_flag": {"defaultValue": False}},
            evaluation_context_enrichment={},
        ),
        FlagConfigResponse(etag="v1", flags={}, evaluation_context_enrichment={}),
    ]
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)
    evaluator.initialize()
    with evaluator._lock:
        assert evaluator._flags == {"my_flag": {"defaultValue": False}}
    evaluator._refresh_flag_configuration()
    with evaluator._lock:
        assert evaluator._flags == {"my_flag": {"defaultValue": False}}
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
    """When first retrieve_flag_configuration raises, initialize propagates the error."""
    from gofeatureflag_python_provider.exceptions import (
        FlagConfigurationUnavailableError,
    )

    mock_api = Mock()
    mock_api.retrieve_flag_configuration.side_effect = (
        FlagConfigurationUnavailableError("endpoint not found")
    )
    evaluator, _ = _make_evaluator_with_mock_wasm(mock_api=mock_api)

    with pytest.raises(FlagConfigurationUnavailableError):
        evaluator.initialize()

    with evaluator._lock:
        assert evaluator._flags == {}
        assert evaluator._etag is None


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
    """Async resolve method delegates to the sync counterpart."""
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
