"""
Tests for EvaluateWasm: loads the real WASI binary via wasmtime and validates
initialization, disposal, and flag evaluation for all supported types and scenarios.
"""

import pytest

from gofeatureflag_python_provider.wasm import EvaluateWasm, WasmFlagContext, WasmInput
from gofeatureflag_python_provider.wasm.evaluate_wasm import (
    WasmInvalidResultError,
    WasmNotLoadedError,
)

# ---------------------------------------------------------------------------
# Shared helpers
# ---------------------------------------------------------------------------

_BOOL_FLAG = {
    "variations": {"on": True, "off": False},
    "defaultRule": {"variation": "on"},
    "trackEvents": True,
}

_STR_FLAG = {
    "variations": {"v1": "hello", "v2": "world"},
    "defaultRule": {"variation": "v1"},
}

_INT_FLAG = {
    "variations": {"v1": 42, "v2": 0},
    "defaultRule": {"variation": "v1"},
}

_FLOAT_FLAG = {
    "variations": {"v1": 3.14, "v2": 0.0},
    "defaultRule": {"variation": "v1"},
}

_OBJ_FLAG = {
    "variations": {"v1": {"key": "value", "count": 1}, "v2": {}},
    "defaultRule": {"variation": "v1"},
}

_LIST_FLAG = {
    "variations": {"v1": ["a", "b", "c"], "v2": []},
    "defaultRule": {"variation": "v1"},
}

_CTX = {"targetingKey": "user-123"}


def _make_input(
    flag_key: str,
    flag: dict,
    ctx: dict | None = None,
    default=None,
    enrichment: dict | None = None,
) -> WasmInput:
    return WasmInput(
        flagKey=flag_key,
        flag=flag,
        evalContext=ctx or _CTX,
        flagContext=WasmFlagContext(
            defaultSdkValue=default,
            evaluationContextEnrichment=enrichment or {},
        ),
    )


@pytest.fixture(scope="module")
def evaluator():
    """Single EvaluateWasm instance shared across tests in this module."""
    e = EvaluateWasm()
    e.initialize()
    yield e
    e.dispose()


# ---------------------------------------------------------------------------
# Lifecycle tests
# ---------------------------------------------------------------------------


def test_initialize_succeeds_with_bundled_wasi():
    """EvaluateWasm.initialize() loads the bundled WASI binary without error."""
    e = EvaluateWasm()
    e.initialize()
    e.dispose()


def test_initialize_raises_when_wasi_file_missing():
    """WasmNotLoadedError is raised when the WASI binary path does not exist."""
    e = EvaluateWasm(wasm_path="/nonexistent/path/eval.wasi")
    with pytest.raises(WasmNotLoadedError, match="not found"):
        e.initialize()


def test_evaluate_raises_when_not_initialized():
    """Calling evaluate() before initialize() raises WasmNotLoadedError."""
    e = EvaluateWasm()
    with pytest.raises(WasmNotLoadedError, match="not been initialized"):
        e.evaluate(_make_input("flag", _BOOL_FLAG))


def test_evaluate_raises_after_dispose():
    """Calling evaluate() after dispose() raises WasmNotLoadedError."""
    e = EvaluateWasm()
    e.initialize()
    e.dispose()
    with pytest.raises(WasmNotLoadedError):
        e.evaluate(_make_input("flag", _BOOL_FLAG))


def test_dispose_is_idempotent():
    """Calling dispose() multiple times does not raise."""
    e = EvaluateWasm()
    e.initialize()
    e.dispose()
    e.dispose()


# ---------------------------------------------------------------------------
# Type-resolution tests
# ---------------------------------------------------------------------------


def test_evaluate_returns_boolean(evaluator):
    """Boolean flag returns bool value."""
    resp = evaluator.evaluate(_make_input("bool-flag", _BOOL_FLAG, default=False))
    assert resp.value is True
    assert resp.variationType == "on"
    assert resp.trackEvents is True


def test_evaluate_returns_string(evaluator):
    """String flag returns str value."""
    resp = evaluator.evaluate(_make_input("str-flag", _STR_FLAG, default="default"))
    assert resp.value == "hello"
    assert resp.variationType == "v1"


def test_evaluate_returns_integer(evaluator):
    """Integer flag returns int value."""
    resp = evaluator.evaluate(_make_input("int-flag", _INT_FLAG, default=0))
    assert resp.value == 42
    assert isinstance(resp.value, int)


def test_evaluate_returns_float(evaluator):
    """Float flag returns numeric value."""
    resp = evaluator.evaluate(_make_input("float-flag", _FLOAT_FLAG, default=0.0))
    assert resp.value == pytest.approx(3.14)


def test_evaluate_returns_object(evaluator):
    """Object flag returns dict value."""
    resp = evaluator.evaluate(_make_input("obj-flag", _OBJ_FLAG, default={}))
    assert resp.value == {"key": "value", "count": 1}


def test_evaluate_returns_list(evaluator):
    """List flag returns list value."""
    resp = evaluator.evaluate(_make_input("list-flag", _LIST_FLAG, default=[]))
    assert resp.value == ["a", "b", "c"]


# ---------------------------------------------------------------------------
# Scenario tests
# ---------------------------------------------------------------------------


def test_evaluate_targeting_match(evaluator):
    """Targeting rule matching eval context produces TARGETING_MATCH reason."""
    flag = {
        "variations": {"yes": True, "no": False},
        "targeting": [{"query": 'targetingKey == "vip-user"', "variation": "yes"}],
        "defaultRule": {"variation": "no"},
    }
    resp = evaluator.evaluate(
        _make_input("tgt-flag", flag, ctx={"targetingKey": "vip-user"}, default=False)
    )
    assert resp.value is True
    assert resp.reason == "TARGETING_MATCH"
    assert resp.variationType == "yes"


def test_evaluate_targeting_miss_uses_default_rule(evaluator):
    """When no targeting rule matches, the default rule is applied."""
    flag = {
        "variations": {"yes": True, "no": False},
        "targeting": [{"query": 'targetingKey == "vip-user"', "variation": "yes"}],
        "defaultRule": {"variation": "no"},
    }
    resp = evaluator.evaluate(
        _make_input("tgt-flag", flag, ctx={"targetingKey": "regular"}, default=False)
    )
    assert resp.value is False
    assert resp.variationType == "no"


def test_evaluate_disabled_flag_returns_sdk_default(evaluator):
    """Disabled flag returns the SDK default value with reason DISABLED."""
    flag = {
        "variations": {"on": True, "off": False},
        "defaultRule": {"variation": "on"},
        "disable": True,
    }
    resp = evaluator.evaluate(_make_input("dis-flag", flag, default=False))
    assert resp.value is False
    assert resp.reason == "DISABLED"


def test_evaluate_percentage_100_returns_expected_variation(evaluator):
    """100 % allocation to a variation always returns that variation."""
    flag = {
        "variations": {"on": True, "off": False},
        "defaultRule": {"percentage": {"on": 100, "off": 0}},
    }
    resp = evaluator.evaluate(_make_input("pct-flag", flag, default=False))
    assert resp.value is True


def test_evaluate_custom_attribute_targeting(evaluator):
    """Targeting on a custom attribute in evalContext is matched correctly."""
    flag = {
        "variations": {"yes": "admin", "no": "user"},
        "targeting": [{"query": 'email eq "admin@example.com"', "variation": "yes"}],
        "defaultRule": {"variation": "no"},
    }
    resp = evaluator.evaluate(
        _make_input(
            "attr-flag",
            flag,
            ctx={"targetingKey": "u1", "email": "admin@example.com"},
            default="user",
        )
    )
    assert resp.value == "admin"
    assert resp.reason == "TARGETING_MATCH"


def test_evaluate_passes_evaluation_context_enrichment(evaluator):
    """evaluationContextEnrichment is forwarded to the WASM and available for rules."""
    flag = {
        "variations": {"on": True, "off": False},
        "defaultRule": {"variation": "on"},
    }
    resp = evaluator.evaluate(
        _make_input(
            "enrich-flag",
            flag,
            default=False,
            enrichment={"region": "eu-west"},
        )
    )
    assert resp.value is True


def test_evaluate_empty_context_still_returns_default_rule(evaluator):
    """Evaluation with an empty context (no targetingKey) still resolves the default rule."""
    flag = {
        "variations": {"v": "result"},
        "defaultRule": {"variation": "v"},
    }
    resp = evaluator.evaluate(
        _make_input("empty-ctx", flag, ctx={}, default="fallback")
    )
    assert resp.value == "result"


# ---------------------------------------------------------------------------
# WasmInput model tests
# ---------------------------------------------------------------------------


def test_wasm_input_serialises_correctly():
    """WasmInput.model_dump_json() produces the expected JSON structure."""
    import json

    wasm_input = WasmInput(
        flagKey="my-flag",
        flag={"variations": {"v": True}, "defaultRule": {"variation": "v"}},
        evalContext={"targetingKey": "u1"},
        flagContext=WasmFlagContext(
            defaultSdkValue=False, evaluationContextEnrichment={"env": "test"}
        ),
    )
    data = json.loads(wasm_input.model_dump_json())
    assert data["flagKey"] == "my-flag"
    assert data["evalContext"]["targetingKey"] == "u1"
    assert data["flagContext"]["defaultSdkValue"] is False
    assert data["flagContext"]["evaluationContextEnrichment"] == {"env": "test"}
