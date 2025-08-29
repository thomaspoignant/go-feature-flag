from typing import Any, Dict
from pydantic import BaseModel
from .flag import Flag
from .flag_context import FlagContext


class WasmInput(BaseModel):
    """
    Represents the input to the WASM module, containing the flag key, flag, evaluation context, and flag context.
    """

    # Flag key to be evaluated.
    flag_key: str

    # Flag to be evaluated.
    flag: Flag

    # Evaluation context for a flag evaluation.
    eval_context: Dict[str, Any]

    # Flag context containing default SDK value and evaluation context enrichment.
    flag_context: FlagContext
