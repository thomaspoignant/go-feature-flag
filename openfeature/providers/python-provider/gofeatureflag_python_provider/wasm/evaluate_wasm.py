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
from typing import Optional

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


class EvaluateWasm:
    """
    Loads and executes the GO Feature Flag evaluation WASI module.

    Usage::

        evaluator = EvaluateWasm()          # or EvaluateWasm(wasm_path="/custom/path.wasi")
        evaluator.initialize()
        response = evaluator.evaluate(wasm_input)
        evaluator.dispose()
    """

    def __init__(self, wasm_path: Optional[str] = None) -> None:
        if wasm_path:
            self._wasm_path = Path(wasm_path)
        else:
            self._wasm_path = _package_root() / _DEFAULT_WASI_RELATIVE_PATH

        self._engine: Optional[wasmtime.Engine] = None
        self._store: Optional[wasmtime.Store] = None
        self._memory: Optional[wasmtime.Memory] = None
        self._malloc_fn: Optional[wasmtime.Func] = None
        self._free_fn: Optional[wasmtime.Func] = None
        self._evaluate_fn: Optional[wasmtime.Func] = None

    # ------------------------------------------------------------------
    # Lifecycle
    # ------------------------------------------------------------------

    def initialize(self) -> None:
        """
        Load the WASI binary, configure the WASI runtime, instantiate the module,
        call _start to initialise the Go runtime, and verify that
        malloc / free / evaluate exports are present.
        """
        if not self._wasm_path.exists():
            raise WasmNotLoadedError(f"WASI binary not found at: {self._wasm_path}")

        self._engine = wasmtime.Engine()

        wasi_cfg = wasmtime.WasiConfig()
        wasi_cfg.inherit_stdout()
        wasi_cfg.inherit_stderr()

        self._store = wasmtime.Store(self._engine)
        self._store.set_wasi(wasi_cfg)

        module = wasmtime.Module.from_file(self._engine, str(self._wasm_path))
        linker = wasmtime.Linker(self._engine)
        linker.define_wasi()

        instance = linker.instantiate(self._store, module)
        exports = instance.exports(self._store)

        self._memory = exports["memory"]
        self._malloc_fn = exports["malloc"]
        self._free_fn = exports["free"]
        self._evaluate_fn = exports["evaluate"]

        for name, fn in (
            ("memory", self._memory),
            ("malloc", self._malloc_fn),
            ("free", self._free_fn),
            ("evaluate", self._evaluate_fn),
        ):
            if fn is None:
                raise WasmFunctionNotFoundError(f"WASI export '{name}' not found")

        # Boot the Go runtime.  _start initialises the scheduler/GC and then
        # exits with code 0 via proc_exit, which wasmtime surfaces as ExitTrap.
        start_fn = exports.get("_start")
        if start_fn is not None:
            try:
                start_fn(self._store)
            except wasmtime.ExitTrap as exc:
                if exc.code != 0:
                    raise WasmNotLoadedError(
                        f"WASI _start exited with non-zero code: {exc.code}"
                    ) from exc

        logger.debug("WASI module initialized from %s", self._wasm_path)

    def dispose(self) -> None:
        """Release all references to the WASM runtime."""
        self._memory = None
        self._malloc_fn = None
        self._free_fn = None
        self._evaluate_fn = None
        self._store = None
        self._engine = None
        logger.debug("WASI module disposed")

    # ------------------------------------------------------------------
    # Evaluation
    # ------------------------------------------------------------------

    def evaluate(self, wasm_input: WasmInput) -> WasmEvaluationResponse:
        """
        Evaluate a feature flag via the WASI module.

        :param wasm_input: Fully-populated WasmInput.
        :returns: WasmEvaluationResponse parsed from the WASI output.
        :raises WasmNotLoadedError: if initialize() has not been called.
        :raises WasmInvalidResultError: if the module returns invalid data.
        """
        if self._store is None or self._memory is None:
            raise WasmNotLoadedError(
                "EvaluateWasm has not been initialized. Call initialize() first."
            )

        input_bytes = wasm_input.model_dump_json().encode("utf-8")
        input_ptr = self._copy_to_memory(input_bytes)

        try:
            result = self._evaluate_fn(self._store, input_ptr, len(input_bytes))
        finally:
            self._free_fn(self._store, input_ptr)

        return self._decode_result(result)

    # ------------------------------------------------------------------
    # Private helpers
    # ------------------------------------------------------------------

    def _copy_to_memory(self, data: bytes) -> int:
        """Allocate WASI memory for *data* (plus null terminator) and write it."""
        ptr = self._malloc_fn(self._store, len(data) + 1)
        if not isinstance(ptr, int):
            raise WasmInvalidResultError(
                f"malloc returned unexpected type {type(ptr).__name__!r}"
            )

        mem_ptr = self._memory.data_ptr(self._store)
        for i, byte in enumerate(data):
            mem_ptr[ptr + i] = byte
        mem_ptr[ptr + len(data)] = 0  # null terminator
        return ptr

    def _decode_result(self, result: int) -> WasmEvaluationResponse:
        """
        Unpack the i64 returned by evaluate() into (output_ptr, output_len),
        read the JSON string from WASI memory, and return a WasmEvaluationResponse.
        """
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

        mem_ptr = self._memory.data_ptr(self._store)
        output_bytes = bytes(mem_ptr[output_ptr : output_ptr + output_len])
        return WasmEvaluationResponse.model_validate_json(output_bytes)
