"""Raised when the WASM module dies mid-call."""


class WasmEvaluationTrapError(RuntimeError):
    """
    Raised when the WASM module died mid-call (stack overflow, unrecoverable
    panic, out-of-memory, or an abort that exits through WASI proc_exit). The
    store has been discarded and replaced with a fresh one; the evaluation
    itself failed.
    """
