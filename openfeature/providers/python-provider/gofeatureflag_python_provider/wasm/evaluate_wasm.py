"""
WASM evaluator: wraps the GO Feature Flag WASI evaluation binary
using wasmtime as the host runtime.

Memory protocol:
  1. JSON-serialize WasmInput → UTF-8 bytes
  2. malloc(len + 1) → input_ptr
  3. Write bytes + null terminator to WASM linear memory
  4. evaluate(input_ptr, len) → result (i64)
  5. free(input_ptr)
  6. output_ptr = result >> 32  (high 32 bits)
     output_len = result & 0xFFFFFFFF  (low 32 bits)
  7. Read output_len bytes from output_ptr, JSON-parse → WasmEvaluationResponse
"""

import logging
from pathlib import Path
from queue import Queue
from typing import Any, Optional

import wasmtime

from gofeatureflag_python_provider.wasm.models import WasmEvaluationResponse, WasmInput

logger = logging.getLogger(__name__)

# Default WASI binary, located at the root of the python-provider package.
_DEFAULT_WASI_RELATIVE_PATH = (
    Path("wasm-releases") / "evaluation" / "gofeatureflag-evaluation_0.2.0.wasi"
)


def _package_root() -> Path:
    """Return the python-provider package root (parent of this file's grandparent)."""
    return Path(__file__).parent.parent.parent


class WasmNotLoadedError(RuntimeError):
    """Raised when evaluate() is called before initialize(), or after dispose()."""


class WasmFunctionNotFoundError(RuntimeError):
    """Raised when a required export (malloc / free / evaluate) is missing."""


class WasmInvalidResultError(RuntimeError):
    """Raised when the WASM module returns an unexpected result."""


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
            self._wasm_path = _package_root() / _DEFAULT_WASI_RELATIVE_PATH
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
        """
        if self._pool is None:
            raise WasmNotLoadedError(
                "EvaluateWasm has not been initialized. Call initialize() first."
            )
        slot = self._pool.get()
        try:
            return self._evaluate_with_slot(slot, wasm_input)
        finally:
            self._pool.put(slot)

    def _evaluate_with_slot(
        self,
        slot: tuple[Any, ...],
        wasm_input: WasmInput,
    ) -> WasmEvaluationResponse:
        """Run evaluation using one Store slot (store, memory, malloc, free, evaluate)."""
        store, memory, malloc_fn, free_fn, evaluate_fn = slot
        input_bytes = wasm_input.model_dump_json().encode("utf-8")
        ptr = malloc_fn(store, len(input_bytes) + 1)
        if not isinstance(ptr, int):
            raise WasmInvalidResultError(
                f"malloc returned unexpected type {type(ptr).__name__!r}"
            )
        memory.write(store, input_bytes + b"\x00", ptr)
        try:
            result = evaluate_fn(store, ptr, len(input_bytes))
        finally:
            free_fn(store, ptr)
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
        output_bytes = memory.read(store, output_ptr, output_ptr + output_len)
        return WasmEvaluationResponse.model_validate_json(output_bytes)
