"""Response deserialized from the WASM evaluate export."""

from typing import Any, Optional

from pydantic import BaseModel


class WasmEvaluationResponse(BaseModel):
    """
    Response JSON deserialized from WASM linear memory after `evaluate` returns.
    """

    value: Optional[Any] = None
    variationType: Optional[str] = None
    reason: Optional[str] = None
    errorCode: Optional[str] = None
    errorDetails: Optional[str] = None
    trackEvents: bool = False
    metadata: Optional[dict[str, Any]] = None
    # Top-level on the engine response, not inside `metadata`. A scheduled
    # rollout step can override it at evaluation time, so this value — not the
    # one in the stored flag configuration — is the authoritative one.
    version: str = ""
