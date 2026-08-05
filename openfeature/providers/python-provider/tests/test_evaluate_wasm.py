"""
Tests for EvaluateWasm: loads the real WASI binary via wasmtime and validates
initialization, disposal, and flag evaluation for all supported types and scenarios.
"""

from __future__ import annotations

import threading
from typing import Optional

import pytest
import wasmtime

import gofeatureflag_python_provider.wasm.evaluate_wasm as evaluate_wasm_module
from gofeatureflag_python_provider.wasm import EvaluateWasm, WasmFlagContext, WasmInput
from gofeatureflag_python_provider.wasm.errors import (
    WasmEvaluationTrapError,
    WasmInvalidResultError,
    WasmNotLoadedError,
    WasmPoolTimeoutError,
)
from tests.wasm_helpers import BOOL_FLAG as _BOOL_FLAG

# ---------------------------------------------------------------------------
# Shared helpers
# ---------------------------------------------------------------------------

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


def test_evaluate_concurrent_threads():
    """Multiple threads can evaluate flags concurrently using the pool; no races or errors."""
    num_threads = 8
    evals_per_thread = 25
    pool_size = 4
    e = EvaluateWasm(pool_size=pool_size)
    e.initialize()
    try:
        flag = {
            "variations": {"on": True, "off": False},
            "defaultRule": {"variation": "on"},
        }
        results_per_thread: list[list[bool]] = []
        errors: list[list[Exception]] = []

        def run_thread(thread_id: int) -> None:
            my_results: list[bool] = []
            my_errors: list[Exception] = []
            for i in range(evals_per_thread):
                try:
                    inp = _make_input(
                        f"concurrent-flag-{thread_id}",
                        flag,
                        ctx={"targetingKey": f"user-{thread_id}-{i}"},
                        default=False,
                    )
                    resp = e.evaluate(inp)
                    my_results.append(resp.value is True)
                except Exception as exc:
                    my_errors.append(exc)
            results_per_thread.append(my_results)
            errors.append(my_errors)

        threads = [
            threading.Thread(target=run_thread, args=(tid,))
            for tid in range(num_threads)
        ]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        flat_errors = [exc for err_list in errors for exc in err_list]
        assert not flat_errors, f"Concurrent evaluation raised: {flat_errors}"
        assert len(results_per_thread) == num_threads
        for tid, results in enumerate(results_per_thread):
            assert len(results) == evals_per_thread, f"Thread {tid} result count"
            assert all(results), f"Thread {tid}: expected all True, got {results}"
    finally:
        e.dispose()


# ---------------------------------------------------------------------------
# Trap handling and buffer protocol (issue #5651)
# ---------------------------------------------------------------------------


def _make_fake_slot(
    output: bytes = b'{"value": true, "variationType": "on"}',
    malloc_ret: int = 8,
    trap_on_evaluate: bool = False,
    evaluate_ret: object = None,
    evaluate_error: Optional[BaseException] = None,
):
    """Duck-typed (store, memory, malloc, free, evaluate) slot recording calls."""
    calls = {"free": 0, "writes": 0}
    store = object()
    out_ptr = 100

    class FakeMemory:
        def write(self, _store, data, ptr):
            calls["writes"] += 1

        def read(self, _store, start, end):
            assert start == out_ptr
            return output

    def malloc_fn(_store, _size):
        return malloc_ret

    def free_fn(_store, _ptr):
        calls["free"] += 1

    def evaluate_fn(_store, _ptr, _length):
        if evaluate_error is not None:
            raise evaluate_error
        if trap_on_evaluate:
            raise wasmtime.Trap("synthetic trap")
        if evaluate_ret is not None:
            return evaluate_ret
        return (out_ptr << 32) | len(output)

    return (store, FakeMemory(), malloc_fn, free_fn, evaluate_fn), calls


def test_evaluate_with_slot_frees_input_after_reading_output():
    """On success, the input buffer is freed exactly once (after the output read)."""
    slot, calls = _make_fake_slot()
    resp = EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert resp.value is True
    assert calls["writes"] == 1
    assert calls["free"] == 1


def test_evaluate_with_slot_skips_free_after_trap():
    """free must never run on a trapped store (it would fault and mask the trap)."""
    slot, calls = _make_fake_slot(trap_on_evaluate=True)
    with pytest.raises(wasmtime.Trap):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["free"] == 0


def test_evaluate_with_slot_skips_free_after_a_wasi_abort():
    """An abort exits through proc_exit, which wasmtime reports as ExitTrap.

    ExitTrap is a WasmtimeError and *not* a Trap subclass, so a handler that
    only names Trap would run free() on a store that is just as poisoned.
    """
    slot, calls = _make_fake_slot(evaluate_error=wasmtime.ExitTrap("wasi proc_exit(1)"))
    with pytest.raises(wasmtime.ExitTrap):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["free"] == 0


def test_wasi_abort_discards_the_store_like_a_trap():
    """The store is poisoned either way, so it must never go back to the pool."""
    e = EvaluateWasm(pool_size=1)
    e.initialize()
    try:
        original_store = e._pool.queue[0][0]
        real_evaluate_with_slot = e._evaluate_with_slot
        raised = []

        def abort_once(slot, wasm_input):
            if not raised:
                raised.append(True)
                raise wasmtime.ExitTrap("wasi proc_exit(1)")
            return real_evaluate_with_slot(slot, wasm_input)

        e._evaluate_with_slot = abort_once
        with pytest.raises(WasmEvaluationTrapError):
            e.evaluate(_make_input("f", _BOOL_FLAG, default=False))

        assert e._pool.qsize() == 1
        assert e._pool.queue[0][0] is not original_store
        assert e.evaluate(_make_input("f", _BOOL_FLAG, default=False)).value is True
    finally:
        e.dispose()


def test_evaluate_with_slot_rejects_null_malloc():
    """A NULL malloc result raises instead of writing to address 0."""
    slot, calls = _make_fake_slot(malloc_ret=0)
    with pytest.raises(WasmInvalidResultError, match="invalid pointer"):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["writes"] == 0
    assert calls["free"] == 0


def test_evaluate_with_slot_rejects_non_int_result():
    """A non-integer evaluate return raises instead of being unpacked."""
    slot, calls = _make_fake_slot(evaluate_ret="nope")
    with pytest.raises(WasmInvalidResultError, match="unexpected type"):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["free"] == 1  # not a trap: the input buffer is still freed


def test_evaluate_with_slot_rejects_zero_output_pointer():
    """A 0 result (no output produced) raises instead of reading address 0."""
    slot, calls = _make_fake_slot(evaluate_ret=1)  # ptr=0, len=1
    with pytest.raises(WasmInvalidResultError, match="null or zero-length"):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["free"] == 1


def test_evaluate_with_slot_wraps_malformed_output():
    """Malformed module output raises the typed error, not a raw pydantic one."""
    slot, calls = _make_fake_slot(output=b"definitely not json")
    with pytest.raises(WasmInvalidResultError, match="malformed output"):
        EvaluateWasm()._evaluate_with_slot(slot, _make_input("f", _BOOL_FLAG))
    assert calls["free"] == 1


def test_trap_recycles_slot_and_next_evaluation_succeeds():
    """A trapped store is discarded and replaced; the pool keeps working."""
    e = EvaluateWasm(pool_size=1)
    e.initialize()
    try:
        original_store = e._pool.queue[0][0]
        real_evaluate_with_slot = e._evaluate_with_slot
        raised = []

        def trap_once(slot, wasm_input):
            if not raised:
                raised.append(True)
                raise wasmtime.Trap("synthetic trap")
            return real_evaluate_with_slot(slot, wasm_input)

        e._evaluate_with_slot = trap_once
        with pytest.raises(WasmEvaluationTrapError):
            e.evaluate(_make_input("f", _BOOL_FLAG, default=False))
        assert e._pool.qsize() == 1
        assert e._pool.queue[0][0] is not original_store

        resp = e.evaluate(_make_input("f", _BOOL_FLAG, default=False))
        assert resp.value is True
    finally:
        e.dispose()


def test_pool_size_preserved_after_trap_storm():
    """Repeated traps never shrink the pool nor break later evaluations."""
    pool_size = 3
    e = EvaluateWasm(pool_size=pool_size)
    e.initialize()
    try:
        real_evaluate_with_slot = e._evaluate_with_slot

        def always_trap(slot, wasm_input):
            raise wasmtime.Trap("synthetic trap")

        e._evaluate_with_slot = always_trap
        for _ in range(10):
            with pytest.raises(WasmEvaluationTrapError):
                e.evaluate(_make_input("f", _BOOL_FLAG, default=False))

        e._evaluate_with_slot = real_evaluate_with_slot
        assert e._pool.qsize() == pool_size
        for _ in range(10):
            assert e.evaluate(_make_input("f", _BOOL_FLAG, default=False)).value is True
    finally:
        e.dispose()


def _drain_pool_by_losing_slots(evaluator, creation_fails, count=1):
    """
    Make the pool genuinely lose `count` slots: trap the evaluation while slot
    creation is broken, so the post-trap replacement fails too. This is the only
    way capacity is really lost, as opposed to merely being checked out.
    """
    real_evaluate_with_slot = evaluator._evaluate_with_slot

    def always_trap(_slot, _wasm_input):
        raise wasmtime.Trap("synthetic trap")

    evaluator._evaluate_with_slot = always_trap
    creation_fails["value"] = True
    try:
        for _ in range(count):
            with pytest.raises(WasmEvaluationTrapError):
                evaluator.evaluate(_make_input("f", _BOOL_FLAG, default=False))
    finally:
        creation_fails["value"] = False
        evaluator._evaluate_with_slot = real_evaluate_with_slot


def _patch_fallible_create_slot(monkeypatch):
    """Patch _create_slot so tests can switch creation failures on and off."""
    creation_fails = {"value": False}
    real_create_slot = evaluate_wasm_module._create_slot

    def maybe_failing_create_slot(*args):
        if creation_fails["value"]:
            raise RuntimeError("instantiation failed")
        return real_create_slot(*args)

    monkeypatch.setattr(evaluate_wasm_module, "_create_slot", maybe_failing_create_slot)
    return creation_fails


def test_lost_slot_is_rebuilt_on_next_evaluation(monkeypatch):
    """Capacity lost to a failed post-trap rebuild is restored, not left missing."""
    monkeypatch.setattr(evaluate_wasm_module, "_POOL_GET_TIMEOUT_SECONDS", 0.05)
    creation_fails = _patch_fallible_create_slot(monkeypatch)
    e = EvaluateWasm(pool_size=1)
    e.initialize()
    try:
        _drain_pool_by_losing_slots(e, creation_fails)
        assert e._pool.qsize() == 0
        assert e._live_slots == 0

        resp = e.evaluate(_make_input("f", _BOOL_FLAG, default=False))
        assert resp.value is True
        assert e._pool.qsize() == 1  # the rebuilt slot went back to the pool
        assert e._live_slots == 1
    finally:
        e.dispose()


def test_empty_pool_raises_typed_error_when_rebuild_fails(monkeypatch):
    """If rebuilding fails too, a typed error surfaces instead of a hang."""
    monkeypatch.setattr(evaluate_wasm_module, "_POOL_GET_TIMEOUT_SECONDS", 0.05)
    creation_fails = _patch_fallible_create_slot(monkeypatch)
    e = EvaluateWasm(pool_size=1)
    e.initialize()
    try:
        _drain_pool_by_losing_slots(e, creation_fails)

        creation_fails["value"] = True
        with pytest.raises(WasmPoolTimeoutError, match="no WASM evaluation slot"):
            e.evaluate(_make_input("f", _BOOL_FLAG, default=False))
        # A failed rebuild must not leave phantom capacity behind, or the pool
        # would never rebuild that slot again.
        assert e._live_slots == 0
    finally:
        e.dispose()


def test_saturated_pool_does_not_grow(monkeypatch):
    """Busy slots are not lost slots: saturation errors instead of adding stores."""
    monkeypatch.setattr(evaluate_wasm_module, "_POOL_GET_TIMEOUT_SECONDS", 0.05)
    e = EvaluateWasm(pool_size=1)
    e.initialize()
    try:
        busy_slot = e._pool.get()  # checked out by another (slow) evaluation
        for _ in range(3):
            with pytest.raises(WasmPoolTimeoutError, match="all 1 slot"):
                e.evaluate(_make_input("f", _BOOL_FLAG, default=False))

        e._pool.put(busy_slot)
        assert e._pool.qsize() == 1  # never grew past pool_size
        assert e._live_slots == 1
        assert e.evaluate(_make_input("f", _BOOL_FLAG, default=False)).value is True
    finally:
        e.dispose()


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
