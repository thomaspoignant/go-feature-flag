"""Raised when no WASM evaluation slot is available within the timeout."""


class WasmPoolTimeoutError(RuntimeError):
    """
    Raised when no evaluation slot became available within the timeout, either
    because every slot is still busy (the pool is undersized for the workload)
    or because slots were lost to irrecoverable store-creation failures and
    rebuilding one failed again.
    """
