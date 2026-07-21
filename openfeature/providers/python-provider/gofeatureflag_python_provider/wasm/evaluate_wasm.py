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
from pathlib import Path
from queue import Empty, Queue
from typing import Any, Optional

import wasmtime
from pydantic import ValidationError

from gofeatureflag_python_provider.wasm.models import WasmEvaluationResponse, WasmInput

logger = logging.getLogger(__name__)

# Directory holding this module and the bundled WASI binary.
_WASM_DIR = Path(__file__).parent


def _read_wasm_version() -> str:
    """
    Read the pinned WASI version from the co-located ``_wasi_version.txt``.

    This file is the single source of truth for the WASI version shipped with
    the provider; the bump automation only updates this file.
    """
    return (_WASM_DIR / "_wasi_version.txt").read_text(encoding="utf-8").strip()


def _default_wasi_path() -> Path:
    """Return the path to the bundled WASI binary for the pinned version."""
    return _WASM_DIR / f"gofeatureflag-evaluation_{_read_wasm_version()}.wasi"


class WasmNotLoadedError(RuntimeError):
    """Raised when evaluate() is called before initialize(), or after dispose()."""


class WasmFunctionNotFoundError(RuntimeError):
    """Raised when a required export (malloc / free / evaluate) is missing."""


class WasmInvalidResultError(RuntimeError):
    """Raised when the WASM module returns an unexpected result."""


class WasmEvaluationTrapError(RuntimeError):
    """
    Raised when the WASM module traps during evaluation (stack overflow,
    unrecoverable panic, out-of-memory...). The store that trapped has been
    discarded and replaced with a fresh one; the evaluation itself failed.
    """


class WasmInputTooDeepError(RuntimeError):
    """
    Raised before calling the WASM module when the input payload is nested too
    deeply. The module decodes JSON recursively on a small fixed-size stack, so
    deeply nested evaluation contexts would overflow it and trap.
    """


class WasmPoolTimeoutError(RuntimeError):
    """
    Raised when no evaluation slot became available within the timeout and a
    replacement slot could not be created. The pool can only stay empty that
    long if slots were lost to irrecoverable store-creation failures or the
    pool is badly undersized for the workload.
    """


# How long evaluate() waits for a free slot before assuming the pool has been
# drained (e.g. every replacement after a trap failed) and trying to self-heal.
_POOL_GET_TIMEOUT_SECONDS = 30.0


# Maximum {}/[] nesting depth accepted for the JSON payload sent to the WASM
# module. The module decodes JSON recursively on a fixed-size shadow stack
# (64KB in binaries <= 0.2.3, where ~250 levels overflow it), so deep payloads
# are rejected host-side before they can trap the module.
_MAX_INPUT_NESTING_DEPTH = 128


def _exceeds_depth(value: Any, limit: int) -> bool:
    """Return True if `value` nests dicts/lists/tuples deeper than `limit` levels."""
    stack = [(value, 1)]
    while stack:
        node, depth = stack.pop()
        if isinstance(node, dict):
            children: Any = node.values()
        elif isinstance(node, (list, tuple)):
            children = node
        else:
            continue
        if depth > limit:
            return True
        stack.extend((child, depth + 1) for child in children)
    return False


def _create_slot(
    engine: wasmtime.Engine,
    module: wasmtime.Module,
) -> tuple[
    wasmtime.Store,
    wasmtime.Memory,
    wasmtime.Func,
    wasmtime.Func,
    wasmtime.Func,
]:
    """Create one Store and its instance (thread-safe evaluation slot)."""
    store = wasmtime.Store(engine)
    wasi_cfg = wasmtime.WasiConfig()
    wasi_cfg.inherit_stdout()
    wasi_cfg.inherit_stderr()
    store.set_wasi(wasi_cfg)
    linker = wasmtime.Linker(engine)
    linker.define_wasi()
    instance = linker.instantiate(store, module)
    exports = instance.exports(store)
    memory = exports["memory"]
    malloc_fn = exports["malloc"]
    free_fn = exports["free"]
    evaluate_fn = exports["evaluate"]
    for name, fn in (
        ("memory", memory),
        ("malloc", malloc_fn),
        ("free", free_fn),
        ("evaluate", evaluate_fn),
    ):
        if fn is None:
            raise WasmFunctionNotFoundError(f"WASI export '{name}' not found")
    start_fn = exports.get("_start")
    if start_fn is not None:
        try:
            start_fn(store)
        except wasmtime.ExitTrap as exc:
            if exc.code != 0:
                raise WasmNotLoadedError(
                    f"WASI _start exited with non-zero code: {exc.code}"
                ) from exc
    return store, memory, malloc_fn, free_fn, evaluate_fn


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
        self._pool_size = 1 if pool_size is None or pool_size < 1 else pool_size
        self._engine: Optional[wasmtime.Engine] = None
        self._module: Optional[wasmtime.Module] = None
        self._pool: Optional[Queue[tuple[Any, ...]]] = None

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
        if _exceeds_depth(
            [
                wasm_input.flag,
                wasm_input.evalContext,
                wasm_input.flagContext.defaultSdkValue,
                wasm_input.flagContext.evaluationContextEnrichment,
            ],
            _MAX_INPUT_NESTING_DEPTH,
        ):
            raise WasmInputTooDeepError(
                "evaluation input exceeds the maximum supported nesting depth "
                f"({_MAX_INPUT_NESTING_DEPTH})"
            )
        # Capture the queue: dispose() may null out self._pool while we hold a
        # slot, and the finally below must return it to the queue it came from.
        pool = self._pool
        try:
            slot = pool.get(timeout=_POOL_GET_TIMEOUT_SECONDS)
        except Empty:
            # The pool drained (replacement slots failed to build after traps)
            # or is badly undersized. Try to heal it instead of blocking the
            # caller forever; on failure degrade to a typed error so the
            # evaluation falls back to the default value.
            try:
                slot = _create_slot(self._engine, self._module)
                logger.warning(
                    "no WASM slot became available within %.0fs; "
                    "created a replacement slot",
                    _POOL_GET_TIMEOUT_SECONDS,
                )
            except Exception as exc:
                raise WasmPoolTimeoutError(
                    "no WASM evaluation slot available after "
                    f"{_POOL_GET_TIMEOUT_SECONDS:.0f}s and creating a "
                    f"replacement slot failed: {exc}"
                ) from exc
        try:
            return self._evaluate_with_slot(slot, wasm_input)
        except wasmtime.Trap as exc:
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
                    # The pool shrinks by one slot; evaluations keep working
                    # on the remaining slots.
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
            except wasmtime.Trap:
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
