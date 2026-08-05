"""Input payload for the WASM evaluate export."""

from typing import Any

from pydantic import BaseModel

from gofeatureflag_python_provider.wasm.model.flag_context import WasmFlagContext


class WasmInput(BaseModel):
    """
    Full input payload serialized to JSON and written into WASM linear memory
    before calling the exported `evaluate` function.
    """

    flagKey: str
    flag: dict[str, Any]
    evalContext: dict[str, Any]
    flagContext: WasmFlagContext
