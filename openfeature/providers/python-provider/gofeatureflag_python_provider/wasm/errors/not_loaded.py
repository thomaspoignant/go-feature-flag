"""Raised when evaluate() is called before initialize(), or after dispose()."""


class WasmNotLoadedError(RuntimeError):
    """Raised when evaluate() is called before initialize(), or after dispose()."""
