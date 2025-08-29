from typing import Any, Dict
from pydantic import BaseModel


class FlagContext(BaseModel):
    """
    Flag context containing default SDK value and evaluation context enrichment.
    """

    # Default SDK value
    default_sdk_value: Any

    # Evaluation context enrichment
    evaluation_context_enrichment: Dict[str, Any]
