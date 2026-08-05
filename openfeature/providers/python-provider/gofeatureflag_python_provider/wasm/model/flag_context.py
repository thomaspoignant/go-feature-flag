"""Context passed alongside the flag definition into the WASM evaluator."""

from typing import Any, Optional

from pydantic import BaseModel


class WasmFlagContext(BaseModel):
    """Context passed alongside the flag definition into the WASM evaluator."""

    defaultSdkValue: Optional[Any] = None
    evaluationContextEnrichment: dict[str, Any] = {}
