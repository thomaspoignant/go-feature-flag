"""Raised when a required WASM export is missing."""


class WasmFunctionNotFoundError(RuntimeError):
    """Raised when a required export (malloc / free / evaluate) is missing."""
