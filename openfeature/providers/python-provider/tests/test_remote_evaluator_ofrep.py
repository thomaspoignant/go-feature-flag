"""
Tests for RemoteEvaluator OFREP bridge (delegation to openfeature-provider-ofrep).
"""

from __future__ import annotations

import asyncio
from unittest.mock import patch

from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagResolutionDetails, Reason

from gofeatureflag_python_provider.evaluator.remote_evaluator import RemoteEvaluator
from gofeatureflag_python_provider.options import GoFeatureFlagOptions


def _make_options(endpoint: str = "http://localhost:1031", api_key: str | None = None):
    return GoFeatureFlagOptions(endpoint=endpoint, api_key=api_key)


@patch("gofeatureflag_python_provider.evaluator.remote_evaluator.OFREPProvider")
def test_remote_evaluator_constructs_with_endpoint(mock_ofrep_class):
    """RemoteEvaluator creates OFREPProvider with relay base URL."""
    options = _make_options(endpoint="http://relay.example:1031")
    RemoteEvaluator(options)
    mock_ofrep_class.assert_called_once()
    call_kw = mock_ofrep_class.call_args[1]
    assert call_kw["base_url"] == "http://relay.example:1031"
    assert call_kw["headers_factory"] is None


@patch("gofeatureflag_python_provider.evaluator.remote_evaluator.OFREPProvider")
def test_remote_evaluator_passes_api_key_via_headers_factory(mock_ofrep_class):
    """When api_key is set, OFREPProvider receives a headers_factory that adds X-API-Key auth."""
    options = _make_options(api_key="secret-key")
    RemoteEvaluator(options)
    call_kw = mock_ofrep_class.call_args[1]
    assert call_kw["headers_factory"] is not None
    headers = call_kw["headers_factory"]()
    assert headers == {"Content-Type": "application/json", "X-API-Key": "secret-key"}


@patch("gofeatureflag_python_provider.evaluator.remote_evaluator.OFREPProvider")
def test_remote_evaluator_initialize_shutdown_no_op(mock_ofrep_class):
    """initialize and shutdown do not call OFREPProvider (it has no such methods)."""
    options = _make_options()
    evaluator = RemoteEvaluator(options)
    evaluator.initialize(None)
    evaluator.initialize(EvaluationContext(targeting_key="user-1"))
    evaluator.shutdown()
    assert mock_ofrep_class.return_value.initialize.call_count == 0
    assert mock_ofrep_class.return_value.shutdown.call_count == 0


@patch("gofeatureflag_python_provider.evaluator.remote_evaluator.OFREPProvider")
def test_remote_evaluator_resolve_boolean_delegates(mock_ofrep_class):
    """resolve_boolean_details delegates to OFREPProvider and returns its result."""
    expected = FlagResolutionDetails(value=True, reason=Reason.TARGETING_MATCH)
    mock_ofrep_class.return_value.resolve_boolean_details.return_value = expected
    options = _make_options()
    evaluator = RemoteEvaluator(options)
    ctx = EvaluationContext(targeting_key="user-1")
    result = evaluator.resolve_boolean_details("my_flag", False, ctx)
    assert result == expected
    mock_ofrep_class.return_value.resolve_boolean_details.assert_called_once_with(
        "my_flag", False, ctx
    )


@patch("gofeatureflag_python_provider.evaluator.remote_evaluator.OFREPProvider")
def test_remote_evaluator_resolve_boolean_async_delegates(mock_ofrep_class):
    """resolve_boolean_details_async runs sync OFREP resolve via asyncio.to_thread."""
    expected = FlagResolutionDetails(value=False, reason=Reason.DISABLED)
    mock_ofrep_class.return_value.resolve_boolean_details.return_value = expected
    options = _make_options()
    evaluator = RemoteEvaluator(options)
    ctx = EvaluationContext(targeting_key="user-2")

    async def run():
        return await evaluator.resolve_boolean_details_async("other_flag", True, ctx)

    result = asyncio.run(run())
    assert result == expected
    mock_ofrep_class.return_value.resolve_boolean_details.assert_called_once_with(
        "other_flag", True, ctx
    )
