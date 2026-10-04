"""
WASM evaluator: wraps the GO Feature Flag WASI evaluation binary
using wasmtime as the host runtime.

Memory protocol:
  1. JSON-serialize WasmInput → UTF-8 bytes
  2. malloc(len + 1) → input_ptr
  3. Write bytes + null terminator to WASM linear memory
  4. evaluate(input_ptr, len) → result (i64)
  5. output_ptr = result >> 32  (high 32 bits)
     output_len = result & 0xFFFFFFFF  (low 32 bits)
  6. Read output_len bytes from output_ptr, JSON-parse → WasmEvaluationResponse
  7. free(input_ptr)

The output buffer is owned by the module's garbage collector and is only
guaranteed to stay valid until the next call into the instance, so it must be
read before `free` (or anything else) runs on that store.

Trap contract: a WASM trap does not unwind the module's shadow stack — the
instance's `__stack_pointer` global keeps its mid-call value forever. A store
that trapped once is therefore permanently poisoned (every later call faults
inside `malloc` at a wrapped ~0xffffXXXX address) and must be discarded and
replaced, never reused, and `free` must never be called on it.
"""

import logging
import threading
from pathlib import Path
from queue import Empty, Queue
from typing import Any, Optional

import wasmtime
from pydantic import ValidationError

from gofeatureflag_python_provider.wasm.errors import (
    WasmEvaluationTrapError,
    WasmInvalidResultError,
    WasmNotLoadedError,
    WasmPoolTimeoutError,
)
from gofeatureflag_python_provider.wasm.model import WasmEvaluationResponse, WasmInput
from gofeatureflag_python_provider.wasm.wasi_runtime import (
    _create_slot,
    _default_wasi_path,
)

logger = logging.getLogger(__name__)

# How long evaluate() waits for a free slot before giving up on it. Deliberately
# short: a slot is held only for one in-process evaluation (milliseconds), and
# the wait ends the moment any slot is returned, so reaching this ceiling means
# either sustained saturation or a pool with no live slots left. In both cases a
# relay proxy round trip is the faster answer, and the lost-slot rebuild below
# cannot even start until this elapses.
_POOL_GET_TIMEOUT_SECONDS = 1.0

# Errors that mean the call into the module died mid-flight, leaving the store
# poisoned. wasmtime raises Trap for a genuine trap, and ExitTrap — which is a
# WasmtimeError, *not* a Trap subclass — when the module calls proc_exit, as
# TinyGo's abort() does. Catching only Trap would let an abort escape with the
# store still in the pool.
_STORE_POISONING_ERRORS = (wasmtime.Trap, wasmtime.WasmtimeError)


class EvaluateWasm:
    """
    Loads and executes the GO Feature Flag evaluation WASI module.

    When pool_size > 1, maintains a pool of wasmtime.Store instances so
    evaluations can run concurrently without blocking each other (Store is
    not thread-safe). When pool_size is 1 or None, uses a single Store.

    Usage::

        evaluator = EvaluateWasm()          # or EvaluateWasm(wasm_path="...", pool_size=4)
        evaluator.initialize()
        response = evaluator.evaluate(wasm_input)
        evaluator.dispose()
    """

    def __init__(
        self,
        wasm_path: Optional[str] = None,
        pool_size: Optional[int] = None,
    ) -> None:
        if wasm_path:
            self._wasm_path = Path(wasm_path)
        else:
            self._wasm_path = _default_wasi_path()
        self._pool_size = 4 if pool_size is None or pool_size < 1 else pool_size
        self._engine: Optional[wasmtime.Engine] = None
        self._module: Optional[wasmtime.Module] = None
        self._pool: Optional[Queue[tuple[Any, ...]]] = None
        # How many slots exist right now, queued or checked out. Only this
        # tells a lost slot apart from a busy one, so rebuilding can be limited
        # to capacity actually lost and never grows the pool past _pool_size.
        self._live_slots = 0
        self._live_slots_lock = threading.Lock()

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def initialize(self) -> None:
        """
        Load the WASI binary, create the engine and module, and either a single
        Store (pool_size=1) or a pool of Stores for concurrent evaluation.
        """
        if not self._wasm_path.exists():
            raise WasmNotLoadedError(f"WASI binary not found at: {self._wasm_path}")

        self._engine = wasmtime.Engine()
        self._module = wasmtime.Module.from_file(self._engine, str(self._wasm_path))
        self._pool = Queue(maxsize=0)
        for _ in range(self._pool_size):
            slot = _create_slot(self._engine, self._module)
            self._pool.put(slot)
        with self._live_slots_lock:
            self._live_slots = self._pool_size
        logger.debug(
            "WASI module initialized from %s (pool_size=%d)",
            self._wasm_path,
            self._pool_size,
        )

    def dispose(self) -> None:
        """Release all references to the WASM runtime."""
        self._pool = None
        self._module = None
        self._engine = None
        with self._live_slots_lock:
            self._live_slots = 0
        logger.debug("WASI module disposed")

    # ------------------------------------------------------------------
    # Evaluation
    # ------------------------------------------------------------------

    def evaluate(self, wasm_input: WasmInput) -> WasmEvaluationResponse:
        """
        Evaluate a feature flag via the WASI module. Uses a slot from the pool
        (or the single store when pool_size=1) so evaluations are thread-safe.

        If the module traps, the slot is discarded and replaced with a fresh
        one (a trapped store is permanently poisoned, see module docstring) and
        WasmEvaluationTrapError is raised.
        """
        if self._pool is None:
            raise WasmNotLoadedError(
                "EvaluateWasm has not been initialized. Call initialize() first."
            )
        # Capture the queue: dispose() may null out self._pool while we hold a
        # slot, and the finally below must return it to the queue it came from.
        pool = self._pool
        try:
            slot = pool.get(timeout=_POOL_GET_TIMEOUT_SECONDS)
        except Empty:
            # A timeout on its own does not mean a slot was lost: every slot may
            # simply still be busy. Rebuild only the capacity known to be gone
            # (slots whose post-trap replacement failed), otherwise sustained
            # saturation would add a store per waiting caller and grow the pool
            # without bound. Plain saturation degrades to a typed error so the
            # evaluation falls back to the default value.
            with self._live_slots_lock:
                lost_capacity = self._live_slots < self._pool_size
                if lost_capacity:
                    # Reserved before the slot exists, so concurrent callers
                    # cannot each rebuild the same missing slot.
                    self._live_slots += 1
            if not lost_capacity:
                raise WasmPoolTimeoutError(
                    "no WASM evaluation slot available after "
                    f"{_POOL_GET_TIMEOUT_SECONDS:g}s; all {self._pool_size} "
                    "slot(s) are in use. Consider increasing wasm_pool_size."
                )
            try:
                slot = _create_slot(self._engine, self._module)
                logger.warning(
                    "no WASM slot became available within %gs; "
                    "rebuilt a slot the pool had lost",
                    _POOL_GET_TIMEOUT_SECONDS,
                )
            except Exception as exc:
                with self._live_slots_lock:
                    self._live_slots -= 1
                raise WasmPoolTimeoutError(
                    "no WASM evaluation slot available after "
                    f"{_POOL_GET_TIMEOUT_SECONDS:g}s and creating a "
                    f"replacement slot failed: {exc}"
                ) from exc
        try:
            return self._evaluate_with_slot(slot, wasm_input)
        except _STORE_POISONING_ERRORS as exc:
            slot = None  # poisoned: never reuse a store that trapped
            raise WasmEvaluationTrapError(
                f"WASM evaluation trapped; the store has been discarded: {exc}"
            ) from exc
        finally:
            if slot is None:
                try:
                    slot = _create_slot(self._engine, self._module)
                    logger.warning(
                        "WASM store trapped during evaluation; "
                        "replaced it with a fresh one"
                    )
                except Exception:
                    # The pool shrinks by one slot; evaluations keep working on
                    # the remaining slots, and the lost capacity is recorded so
                    # a later timeout can rebuild exactly this slot.
                    with self._live_slots_lock:
                        self._live_slots -= 1
                    logger.exception(
                        "failed to replace a trapped WASM store; "
                        "the evaluation pool lost one slot"
                    )
            if slot is not None:
                pool.put(slot)

    def _evaluate_with_slot(
        self,
        slot: tuple[Any, ...],
        wasm_input: WasmInput,
    ) -> WasmEvaluationResponse:
        """Run evaluation using one Store slot (store, memory, malloc, free, evaluate)."""
        store, memory, malloc_fn, free_fn, evaluate_fn = slot
        input_bytes = wasm_input.model_dump_json().encode("utf-8")
        ptr = malloc_fn(store, len(input_bytes) + 1)
        if not isinstance(ptr, int) or ptr == 0:
            raise WasmInvalidResultError(f"malloc returned an invalid pointer: {ptr!r}")
        trapped = False
        try:
            memory.write(store, input_bytes + b"\x00", ptr)
            try:
                result = evaluate_fn(store, ptr, len(input_bytes))
            except _STORE_POISONING_ERRORS:
                trapped = True
                raise
            if not isinstance(result, int):
                raise WasmInvalidResultError(
                    f"evaluate returned unexpected type {type(result).__name__!r}"
                )
            output_ptr = (result >> 32) & 0xFFFFFFFF
            output_len = result & 0xFFFFFFFF
            if output_ptr == 0 or output_len == 0:
                raise WasmInvalidResultError(
                    "evaluate returned a null or zero-length output pointer"
                )
            # Read the output before any further call into the module: the
            # buffer belongs to the module's GC and is only guaranteed to
            # survive until the next call into this instance.
            output_bytes = memory.read(store, output_ptr, output_ptr + output_len)
            try:
                return WasmEvaluationResponse.model_validate_json(output_bytes)
            except ValidationError as exc:
                raise WasmInvalidResultError(
                    f"module returned malformed output: {exc}"
                ) from exc
        finally:
            # Never call free on a trapped store: the trap left the shadow
            # stack pointer unrestored, so the call would fault and mask the
            # original error. The caller discards the store anyway.
            if not trapped:
                free_fn(store, ptr)
