"""Tests for EnrichEvaluationContextHook."""

import pytest

from gofeatureflag_python_provider.hooks import EnrichEvaluationContextHook
from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagType
from openfeature.hook import HookContext


def _make_hook_context(
    targeting_key: str = "user-1",
    attributes: dict | None = None,
) -> HookContext:
    ctx = EvaluationContext(
        targeting_key=targeting_key,
        attributes=attributes or {},
    )
    return HookContext(
        flag_key="test_flag",
        flag_type=FlagType.BOOLEAN,
        default_value=False,
        evaluation_context=ctx,
    )


def test_before_adds_gofeatureflag_when_metadata_non_empty():
    """When metadata is non-empty, returned context includes gofeatureflag attribute."""
    metadata = {"provider": "python", "version": "1.0"}
    hook = EnrichEvaluationContextHook(metadata=metadata)
    hc = _make_hook_context(attributes={"email": "a@b.com"})

    result = hook.before(hc, {})

    assert result is not None
    assert result.targeting_key == "user-1"
    assert result.attributes.get("email") == "a@b.com"
    assert result.attributes.get("gofeatureflag") == {"exporterMetadata": metadata}


def test_before_does_not_add_gofeatureflag_when_metadata_empty():
    """When metadata is empty dict, returned context has no gofeatureflag key."""
    hook = EnrichEvaluationContextHook(metadata={})
    hc = _make_hook_context(attributes={"key": "value"})

    result = hook.before(hc, {})

    assert result is not None
    assert result.targeting_key == "user-1"
    assert result.attributes.get("key") == "value"
    assert "gofeatureflag" not in result.attributes


def test_before_does_not_add_gofeatureflag_when_metadata_none():
    """When metadata is None (default), returned context has no gofeatureflag key."""
    hook = EnrichEvaluationContextHook()
    hc = _make_hook_context()

    result = hook.before(hc, {})

    assert result is not None
    assert "gofeatureflag" not in result.attributes


def test_before_preserves_targeting_key_and_attributes():
    """Returned context preserves original targeting_key and all attributes."""
    metadata = {"env": "test"}
    hook = EnrichEvaluationContextHook(metadata=metadata)
    hc = _make_hook_context(
        targeting_key="custom-key",
        attributes={"a": 1, "b": "two"},
    )

    result = hook.before(hc, {})

    assert result is not None
    assert result.targeting_key == "custom-key"
    assert result.attributes["a"] == 1
    assert result.attributes["b"] == "two"
    assert result.attributes["gofeatureflag"] == {"exporterMetadata": metadata}
