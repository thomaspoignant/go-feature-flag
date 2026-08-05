"""WASI binary path resolution and wasmtime store slot creation."""

from pathlib import Path

import wasmtime

from gofeatureflag_python_provider.wasm.errors import (
    WasmFunctionNotFoundError,
    WasmNotLoadedError,
)

# Directory holding the bundled WASI binary and version pin.
_WASM_DIR = Path(__file__).parent


def _default_wasi_path() -> Path:
    """
    Return the path to the bundled WASI binary for the pinned version.

    The version is read from the co-located ``_wasi_version.txt``, which is the
    single source of truth shipped with the provider; the bump automation only
    updates that file.
    """
    version = (_WASM_DIR / "_wasi_version.txt").read_text(encoding="utf-8").strip()
    return _WASM_DIR / f"gofeatureflag-evaluation_{version}.wasi"


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
