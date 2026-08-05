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


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_remote_evaluator_constructs_with_endpoint(mock_ofrep_class):
    """RemoteEvaluator creates OFREPProvider with relay base URL."""
    options = _make_options(endpoint="http://relay.example:1031")
    RemoteEvaluator(options)
    mock_ofrep_class.assert_called_once()
    call_kw = mock_ofrep_class.call_args[1]
    assert call_kw["base_url"] == "http://relay.example:1031"
    assert call_kw["headers_factory"] is None


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_remote_evaluator_passes_api_key_via_headers_factory(mock_ofrep_class):
    """When api_key is set, OFREPProvider receives a headers_factory with Bearer auth."""
    options = _make_options(api_key="secret-key")
    RemoteEvaluator(options)
    call_kw = mock_ofrep_class.call_args[1]
    assert call_kw["headers_factory"] is not None
    headers = call_kw["headers_factory"]()
    assert headers == {
        "Content-Type": "application/json",
        "X-API-Key": "secret-key",
    }


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_remote_evaluator_applies_the_configured_timeout(mock_ofrep_class):
    """The configured timeout must reach OFREP rather than its own default."""
    RemoteEvaluator(GoFeatureFlagOptions(endpoint="http://localhost:1031"))
    assert mock_ofrep_class.call_args[1]["timeout"] == 10.0

    mock_ofrep_class.reset_mock()
    RemoteEvaluator(
        GoFeatureFlagOptions(endpoint="http://localhost:1031", timeout=2_500)
    )
    assert mock_ofrep_class.call_args[1]["timeout"] == 2.5


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_remote_evaluator_initialize_shutdown_no_op(mock_ofrep_class):
    """initialize and shutdown do not call OFREPProvider (it has no such methods)."""
    options = _make_options()
    evaluator = RemoteEvaluator(options)
    evaluator.initialize(None)
    evaluator.initialize(EvaluationContext(targeting_key="user-1"))
    evaluator.shutdown()
    assert mock_ofrep_class.return_value.initialize.call_count == 0
    assert mock_ofrep_class.return_value.shutdown.call_count == 0


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_remote_evaluator_flags_are_never_trackable(mock_ofrep_class):
    """Remote evaluations are recorded by the relay proxy; emitting a feature
    event for them would count the same evaluation twice, so every flag is
    untrackable in remote mode."""
    evaluator = RemoteEvaluator(_make_options())
    assert evaluator.is_flag_trackable("any-flag") is False
    assert evaluator.is_flag_trackable("") is False


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
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


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
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


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_custom_headers_reach_the_ofrep_client(mock_ofrep_class):
    """Remote evaluation and the in-process fallback share this client."""
    RemoteEvaluator(
        GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            api_key="real-key",
            custom_headers={"X-Gateway-Token": "gw-secret"},
        )
    )

    headers = mock_ofrep_class.call_args[1]["headers_factory"]()

    assert headers["X-Gateway-Token"] == "gw-secret"
    # A configured api_key still wins, as it does on the other HTTP path.
    assert headers["X-API-Key"] == "real-key"


@patch("gofeatureflag_python_provider.services.ofrep.OFREPProvider")
def test_custom_headers_alone_still_build_a_headers_factory(mock_ofrep_class):
    """Without an api_key there would otherwise be no factory to carry them."""
    RemoteEvaluator(
        GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            custom_headers={"X-Gateway-Token": "gw-secret"},
        )
    )

    factory = mock_ofrep_class.call_args[1]["headers_factory"]

    assert factory is not None
    assert factory()["X-Gateway-Token"] == "gw-secret"
