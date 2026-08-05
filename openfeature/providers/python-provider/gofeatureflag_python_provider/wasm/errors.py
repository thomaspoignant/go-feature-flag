"""WASM evaluator errors."""


class WasmNotLoadedError(RuntimeError):
    """Raised when evaluate() is called before initialize(), or after dispose()."""


class WasmFunctionNotFoundError(RuntimeError):
    """Raised when a required export (malloc / free / evaluate) is missing."""


class WasmInvalidResultError(RuntimeError):
    """Raised when the WASM module returns an unexpected result."""


class WasmEvaluationTrapError(RuntimeError):
    """
    Raised when the WASM module died mid-call (stack overflow, unrecoverable
    panic, out-of-memory, or an abort that exits through WASI proc_exit). The
    store has been discarded and replaced with a fresh one; the evaluation
    itself failed.
    """


class WasmPoolTimeoutError(RuntimeError):
    """
    Raised when no evaluation slot became available within the timeout, either
    because every slot is still busy (the pool is undersized for the workload)
    or because slots were lost to irrecoverable store-creation failures and
    rebuilding one failed again.
    """
