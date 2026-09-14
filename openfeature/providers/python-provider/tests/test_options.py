"""
Tests for GoFeatureFlagOptions defaults and validation.

Defaults are normative: a fleet whose providers disagree about them behaves
inconsistently across languages, which no other test would catch.
"""

from __future__ import annotations

import os

import pytest
from pydantic import ValidationError

from gofeatureflag_python_provider.options import EvaluationType, GoFeatureFlagOptions


def _options(**kwargs):
    return GoFeatureFlagOptions(endpoint="http://localhost:1031", **kwargs)


def test_endpoint_is_required_and_validated_at_construction():
    with pytest.raises(ValidationError):
        GoFeatureFlagOptions()  # type: ignore[call-arg]
    with pytest.raises(ValidationError):
        GoFeatureFlagOptions(endpoint="not-a-url")  # type: ignore[arg-type]


def test_default_evaluation_mode_is_in_process():
    assert _options().evaluation_type == EvaluationType.INPROCESS


def test_canonical_defaults():
    options = _options()

    assert options.timeout == 10_000
    assert options.flag_config_poll_interval_seconds == 120
    assert options.data_flush_interval == 60_000
    assert options.max_pending_events == 10_000
    assert options.disable_data_collection is False
    assert options.data_collector_base_url is None
    assert options.api_key is None


def test_wasm_pool_size_defaults_to_cpu_count():
    expected = os.cpu_count() or 1

    assert _options().wasm_pool_size == expected
    assert _options().get_wasm_pool_size() == expected
    # An explicit value wins, and a nonsensical one falls back rather than
    # building a zero-slot pool that could never serve an evaluation.
    assert _options(wasm_pool_size=3).get_wasm_pool_size() == 3
    assert _options(wasm_pool_size=0).get_wasm_pool_size() == expected
    assert _options(wasm_pool_size=None).get_wasm_pool_size() == expected


def test_removed_options_are_gone():
    """Vestigial options for the removed cache and WebSocket features.

    They were declared and documented but never read, so every value a user set
    was silently ignored.
    """
    for name in ("cache_size", "reconnect_interval", "disable_cache_invalidation"):
        assert name not in GoFeatureFlagOptions.model_fields


def test_exporter_metadata_accepts_scalar_values():
    options = _options(
        exporter_metadata={"name": "demo", "beta": True, "id": 7, "rate": 1.5}
    )
    assert options.exporter_metadata == {
        "name": "demo",
        "beta": True,
        "id": 7,
        "rate": 1.5,
    }


@pytest.mark.parametrize(
    "value",
    [
        {"nested": {"a": 1}},
        {"list": ["a", "b"]},
        {"none": None},
    ],
)
def test_exporter_metadata_rejects_non_scalar_values_at_construction(value):
    """Rejected here rather than inside the publisher.

    A value that only fails at publish time makes the batch fail, get re-queued
    and retried forever, with the real cause buried in a log line.
    """
    with pytest.raises(ValidationError):
        _options(exporter_metadata=value)


def test_exporter_metadata_always_includes_the_reserved_keys():
    assert _options().get_exporter_metadata() == {
        "provider": "python",
        "openfeature": True,
    }
    assert _options(exporter_metadata={"appName": "demo"}).get_exporter_metadata() == {
        "appName": "demo",
        "provider": "python",
        "openfeature": True,
    }


def test_specification_version_is_exposed():
    """Which specification version this provider targets, machine-readably."""
    import gofeatureflag_python_provider as provider

    assert provider.__specification_version__ == "1.0"


def test_evaluation_flag_list_defaults_to_all_flags():
    assert _options().evaluation_flag_list is None
    assert _options(evaluation_flag_list=["a", "b"]).evaluation_flag_list == ["a", "b"]


def test_custom_headers_default_to_none():
    assert _options().custom_headers is None
