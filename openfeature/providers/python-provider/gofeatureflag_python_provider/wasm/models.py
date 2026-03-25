"""
Pydantic models for the WASM evaluator input/output.
"""

from typing import Any, Optional

from pydantic import BaseModel


class WasmFlagContext(BaseModel):
    """Context passed alongside the flag definition into the WASM evaluator."""

    defaultSdkValue: Optional[Any] = None
    evaluationContextEnrichment: dict[str, Any] = {}


class WasmInput(BaseModel):
    """
    Full input payload serialized to JSON and written into WASM linear memory
    before calling the exported `evaluate` function.
    """

    flagKey: str
    flag: dict[str, Any]
    evalContext: dict[str, Any]
    flagContext: WasmFlagContext


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
