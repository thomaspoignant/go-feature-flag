"""
Diagnosis and regression tests for issue #5651: WASM stack-overflow traps and
poisoned-store reuse.

Binaries up to 0.2.3 have a 64KB shadow stack placed at the bottom of linear
memory (wasm-ld --stack-first, TinyGo default stack size). Deeply nested JSON
overflows it during decoding: the stack pointer wraps below address 0 and the
call traps out-of-bounds at a ~0xffffXXXX address. Because a trap never unwinds
the module's __stack_pointer global, a store that trapped once is permanently
poisoned — every later call faults inside `malloc` — which is the production
signature reported in the issue.

These tests document that failure mode against the vulnerable bundled binary
and pin the invariant that the host can never be poisoned, whatever binary is
in use.

Environment switches:
  GOFF_TEST_WASI_PATH  path to a candidate .wasi binary to test instead of the
                       bundled one (used to validate new wasm releases).
  GOFF_SOAK=1          enable the long-running soak test.
"""

from __future__ import annotations

import json
import os
import re
from pathlib import Path

import pytest
import wasmtime

from gofeatureflag_python_provider.wasm import (
    EvaluateWasm,
    WasmEvaluationTrapError,
    WasmFlagContext,
    WasmInput,
    WasmInputTooDeepError,
)
from gofeatureflag_python_provider.wasm.evaluate_wasm import (
    _create_slot,
    _default_wasi_path,
)
from tests.wasm_helpers import BOOL_FLAG as _BOOL_FLAG
from tests.wasm_helpers import int_in_list_query as _int_in_list_query
from tests.wasm_helpers import nested_ctx as _nested_ctx
from tests.wasm_helpers import or_chain_query as _or_chain_query
from tests.wasm_helpers import query_flag as _query_flag
from tests.wasm_helpers import split_int_in_list_query as _split_int_in_list_query

_WASI_OVERRIDE = os.environ.get("GOFF_TEST_WASI_PATH")

_FAULT_RE = re.compile(r"memory fault at wasm address (0x[0-9a-f]+)")


def _wasi_path() -> Path:
    return Path(_WASI_OVERRIDE) if _WASI_OVERRIDE else _default_wasi_path()


def _is_known_vulnerable_binary() -> bool:
    """True when testing a bundled binary <= 0.2.3 (64KB stack, no guards)."""
    if _WASI_OVERRIDE:
        return False
    match = re.search(r"_(\d+)\.(\d+)\.(\d+)\.wasi$", _wasi_path().name)
    if match is None:
        return False
    return tuple(int(part) for part in match.groups()) <= (0, 2, 3)


requires_vulnerable_binary = pytest.mark.skipif(
    not _is_known_vulnerable_binary(),
    reason="documents the failure mode of bundled binaries <= 0.2.3",
)


def _payload(ctx: dict) -> bytes:
    return json.dumps(
        {
            "flagKey": "diag-flag",
            "flag": _BOOL_FLAG,
            "evalContext": ctx,
            "flagContext": {"defaultSdkValue": False},
        }
    ).encode("utf-8")


def _wasm_input(ctx: dict) -> WasmInput:
    return WasmInput(
        flagKey="diag-flag",
        flag=_BOOL_FLAG,
        evalContext=ctx,
        flagContext=WasmFlagContext(defaultSdkValue=False),
    )


@pytest.fixture(scope="module")
def wasm_module() -> tuple[wasmtime.Engine, wasmtime.Module]:
    engine = wasmtime.Engine()
    return engine, wasmtime.Module.from_file(engine, str(_wasi_path()))


def _new_slot(wasm_module: tuple) -> tuple:
    engine, module = wasm_module
    return _create_slot(engine, module)


def _raw_evaluate(slot: tuple, payload: bytes) -> bytes:
    """Bare host protocol with no trap protection (deliberately unhardened)."""
    store, memory, malloc_fn, free_fn, evaluate_fn = slot
    ptr = malloc_fn(store, len(payload) + 1)
    memory.write(store, payload + b"\x00", ptr)
    result = evaluate_fn(store, ptr, len(payload))
    output_ptr = (result >> 32) & 0xFFFFFFFF
    output_len = result & 0xFFFFFFFF
    output = bytes(memory.read(store, output_ptr, output_ptr + output_len))
    free_fn(store, ptr)
    return output


def _fault_address(exc: BaseException) -> int | None:
    match = _FAULT_RE.search(str(exc))
    return int(match.group(1), 16) if match else None


def _query_payload(query: str, ctx: dict) -> bytes:
    return json.dumps(
        {
            "flagKey": "diag-flag",
            "flag": _query_flag(query),
            "evalContext": ctx,
            "flagContext": {"defaultSdkValue": False},
        }
    ).encode("utf-8")


# ---------------------------------------------------------------------------
# Failure-mode documentation (vulnerable binaries only)
# ---------------------------------------------------------------------------


@requires_vulnerable_binary
def test_deep_nested_context_overflows_the_shadow_stack(wasm_module):
    """A ~3KB payload with a 400-deep context traps the module (stack overflow)."""
    slot = _new_slot(wasm_module)
    payload = _payload(_nested_ctx(400))
    assert len(payload) < 5_000
    with pytest.raises(wasmtime.Trap):
        _raw_evaluate(slot, payload)


@requires_vulnerable_binary
def test_moderate_nesting_evaluates_fine(wasm_module):
    """Depth 200 fits in the 64KB stack: the trap threshold sits above it."""
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(slot, _payload(_nested_ctx(200)))
    assert b"variationType" in output


@requires_vulnerable_binary
def test_trapped_store_is_poisoned_and_faults_ratchet_downward(wasm_module):
    """
    After one trap, every valid evaluation on the same store faults inside
    malloc at a wrapped ~0xffffXXXX address that decreases call after call —
    the exact signature reported in issue #5651.
    """
    slot = _new_slot(wasm_module)
    baseline = _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
    assert b"variationType" in baseline

    with pytest.raises(wasmtime.Trap):
        _raw_evaluate(slot, _payload(_nested_ctx(400)))

    addresses = []
    for _ in range(5):
        with pytest.raises(wasmtime.Trap) as excinfo:
            _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
        assert "malloc" in str(excinfo.value)
        address = _fault_address(excinfo.value)
        assert address is not None and address >= 0xFFFF0000
        addresses.append(address)
    assert addresses == sorted(addresses, reverse=True)
    assert len(set(addresses)) == len(addresses)


# ---------------------------------------------------------------------------
# Host invariant: no payload may poison the evaluator (any binary version)
# ---------------------------------------------------------------------------


def test_hostile_payload_never_poisons_the_evaluator():
    """
    The invariant under hostile inputs is recovery, not any particular
    outcome of the hostile call itself: depending on the binary it may trap
    (converted to a typed error, store recycled), be rejected by a guard
    (host-side typed error or a structured error result), or even evaluate
    cleanly (a future binary with a bigger stack). Whatever happens, later
    evaluations must succeed and the pool must keep its size.

    The inputs cover every known recursion driver: deeply nested query
    parentheses (~30 levels trap 0.2.3), huge flat `in` lists (breadth, the
    issue #5651 production shape), flat `and`/`or` chains (condition count),
    and a deeply nested evaluation context (input JSON depth).
    """
    deep_parens = "(" * 300 + 'targetingKey eq "u"' + ")" * 300
    huge_in_list = "targetingKey in [" + ",".join(f'"u{i}"' for i in range(2500)) + "]"
    huge_int_list = _int_in_list_query(20_000)
    huge_or_chain = _or_chain_query(2_000)
    hostile_inputs = [
        *(
            WasmInput(
                flagKey="hostile-flag",
                flag=_query_flag(query),
                evalContext={"targetingKey": "user-1"},
                flagContext=WasmFlagContext(defaultSdkValue=False),
            )
            for query in (deep_parens, huge_in_list, huge_int_list, huge_or_chain)
        ),
        WasmInput(
            flagKey="hostile-flag",
            flag=_BOOL_FLAG,
            evalContext=_nested_ctx(400),
            flagContext=WasmFlagContext(defaultSdkValue=False),
        ),
    ]
    pool_size = 2
    e = EvaluateWasm(wasm_path=_WASI_OVERRIDE, pool_size=pool_size)
    e.initialize()
    try:
        for hostile in hostile_inputs:
            for _ in range(2 * pool_size):  # hit every pool slot at least twice
                try:
                    e.evaluate(hostile)
                except (WasmEvaluationTrapError, WasmInputTooDeepError):
                    pass  # typed error; a trapped store has been recycled

        for _ in range(20):
            response = e.evaluate(_wasm_input({"targetingKey": "user-1"}))
            assert response.value is True
        assert e._pool.qsize() == pool_size
    finally:
        e.dispose()


def test_invalid_json_input_returns_structured_error(wasm_module):
    """
    Literally invalid JSON must come back as a structured PARSE_ERROR from the
    built binary, not a trap. encoding/json (and the module's own recovery
    path) rely on Go's panic/recover machinery internally, so this doubles as
    proof that TinyGo's recover works in the shipped binary — if it did not,
    this input would trap and poison the store.
    """
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(slot, b'{"flagKey": not-valid-json')
    result = json.loads(output)
    assert result["errorCode"] == "PARSE_ERROR"

    # The store must still be healthy afterwards.
    ok = _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
    assert b"variationType" in ok


def test_moderate_paren_query_needs_more_than_64kb_stack():
    """
    A 45-paren nikunjy query passes the nesting guard (limit 64) and must
    evaluate cleanly on hardened binaries — it needs more than a 64KB stack,
    so this test fails functionally if the stack-size linker flag is ever
    lost (see .github/ci-scripts/verify-wasm-stack.py for the build-time
    check). On the vulnerable bundled 0.2.3 it traps instead, which the
    evaluator must survive via the recycle path.
    """
    query = "(" * 45 + 'targetingKey eq "user-1"' + ")" * 45
    e = EvaluateWasm(wasm_path=_WASI_OVERRIDE, pool_size=1)
    e.initialize()
    try:
        wasm_input = WasmInput(
            flagKey="paren-flag",
            flag=_query_flag(query),
            evalContext={"targetingKey": "user-1"},
            flagContext=WasmFlagContext(defaultSdkValue=False),
        )
        if _is_known_vulnerable_binary():
            with pytest.raises(WasmEvaluationTrapError):
                e.evaluate(wasm_input)
        else:
            response = e.evaluate(wasm_input)
            assert response.errorCode in (None, "")
            assert response.value is True
        # Either way the evaluator keeps working.
        follow_up = e.evaluate(_wasm_input({"targetingKey": "user-1"}))
        assert follow_up.value is True
    finally:
        e.dispose()


# ---------------------------------------------------------------------------
# Issue #5651 reporter shape: flat `in` lists (breadth, not nesting)
# ---------------------------------------------------------------------------


@requires_vulnerable_binary
def test_reporter_int_list_traps_on_vulnerable_binary(wasm_module):
    """
    The reporters' production trigger: a flat `in` list of ~200 integers.
    Bracket depth is 1, so no nesting guard can see it — the right-recursive
    list parser (one SubListOfInts frame per item) overflows the 64KB stack
    at 154 items.
    """
    slot = _new_slot(wasm_module)
    with pytest.raises(wasmtime.Trap):
        _raw_evaluate(
            slot,
            _query_payload(
                _int_in_list_query(200), {"targetingKey": "user-1", "age": 150}
            ),
        )


def test_reporter_int_list_evaluates_cleanly_on_hardened_binary():
    """
    Acceptance test for the production case of issue #5651: the exact flag
    shape that trapped (flat 200-integer `in` list) must evaluate cleanly on
    hardened binaries (1MB stack: first trap moves to 2,947 items, and the
    breadth guard rejects lists above 1,000 with a structured error first).
    """
    e = EvaluateWasm(wasm_path=_WASI_OVERRIDE, pool_size=1)
    e.initialize()
    try:
        wasm_input = WasmInput(
            flagKey="reporter-flag",
            flag=_query_flag(_int_in_list_query(200)),
            evalContext={"targetingKey": "user-1", "age": 150},
            flagContext=WasmFlagContext(defaultSdkValue=False),
        )
        if _is_known_vulnerable_binary():
            with pytest.raises(WasmEvaluationTrapError):
                e.evaluate(wasm_input)
        else:
            response = e.evaluate(wasm_input)
            assert response.errorCode in (None, "")
            assert response.value is True
        # Either way the evaluator keeps working.
        follow_up = e.evaluate(_wasm_input({"targetingKey": "user-1"}))
        assert follow_up.value is True
    finally:
        e.dispose()


@requires_vulnerable_binary
def test_workaround_split_in_list_avoids_trap_on_vulnerable_binary(wasm_module):
    """
    The flag-config workaround recommended on the issue: splitting the same
    200-integer allow-list into or-joined `in` chunks of 50 keeps the parser
    recursion at 50 frames per list and evaluates fine even on the shipped
    0.2.3 binary (64KB stack).
    """
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(
        slot,
        _query_payload(
            _split_int_in_list_query(200, chunk=50),
            {"targetingKey": "user-1", "age": 150},
        ),
    )
    assert b"variationType" in output
    result = json.loads(output)
    assert result["value"] is True


@pytest.mark.skipif(
    _is_known_vulnerable_binary(),
    reason="the breadth guard only exists in binaries > 0.2.3",
)
def test_over_breadth_list_returns_structured_error(wasm_module):
    """
    A list too large for any stack must be rejected by the breadth guard as a
    structured PARSE_ERROR before the parser recurses — the store stays
    healthy (5,000 items would trap even the 1MB stack at ~2,947).
    """
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(
        slot,
        _query_payload(_int_in_list_query(5_000), {"targetingKey": "user-1"}),
    )
    result = json.loads(output)
    assert result["errorCode"] == "PARSE_ERROR"
    assert "maximum item count" in result["errorDetails"]

    # The guard fired before any recursion: the same store keeps working.
    ok = _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
    assert b"variationType" in ok


@pytest.mark.skipif(
    _is_known_vulnerable_binary(),
    reason="the input nesting guard only exists in binaries > 0.2.3",
)
def test_over_deep_input_returns_structured_error(wasm_module):
    """
    Input JSON nested beyond the module's depth budget must be rejected by
    the module itself as a structured PARSE_ERROR before the recursive JSON
    decode — hosts need no pre-flight depth guard of their own.
    """
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(slot, _payload(_nested_ctx(400)))
    result = json.loads(output)
    assert result["errorCode"] == "PARSE_ERROR"
    assert "maximum nesting depth" in result["errorDetails"]

    # The guard fired before any recursion: the same store keeps working.
    ok = _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
    assert b"variationType" in ok


@pytest.mark.skipif(
    _is_known_vulnerable_binary(),
    reason="the condition-count guard only exists in binaries > 0.2.3",
)
def test_over_condition_chain_returns_structured_error(wasm_module):
    """
    A flat and/or chain too long for any stack must be rejected by the
    condition-count guard as a structured PARSE_ERROR before the parser
    recurses (5,000 conditions would trap even the 1MB stack at ~3,266).
    """
    slot = _new_slot(wasm_module)
    output = _raw_evaluate(
        slot,
        _query_payload(_or_chain_query(5_000), {"targetingKey": "user-1", "age": 1}),
    )
    result = json.loads(output)
    assert result["errorCode"] == "PARSE_ERROR"
    assert "maximum condition count" in result["errorDetails"]

    # The guard fired before any recursion: the same store keeps working.
    ok = _raw_evaluate(slot, _payload({"targetingKey": "user-1"}))
    assert b"variationType" in ok


# ---------------------------------------------------------------------------
# Soak test (opt-in): linear memory must plateau, not grow unbounded
# ---------------------------------------------------------------------------


@pytest.mark.skipif(
    os.environ.get("GOFF_SOAK") != "1",
    reason="long-running soak test; set GOFF_SOAK=1 to enable",
)
def test_soak_linear_memory_plateaus(wasm_module):
    """50k evaluations on one store: memory may grow early but must plateau."""
    slot = _new_slot(wasm_module)
    store, memory = slot[0], slot[1]
    samples = []
    for i in range(50_000):
        output = _raw_evaluate(
            slot,
            _payload({"targetingKey": f"user-{i}", "email": f"user-{i}@example.com"}),
        )
        assert b"variationType" in output
        if i % 5_000 == 0:
            samples.append(memory.data_len(store))
    samples.append(memory.data_len(store))
    midpoint = samples[len(samples) // 2]
    assert samples[-1] == midpoint, f"linear memory still growing: {samples}"
