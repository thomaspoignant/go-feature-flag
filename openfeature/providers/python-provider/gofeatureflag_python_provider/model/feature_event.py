from datetime import datetime
from typing import Any, Dict, Optional
from pydantic import BaseModel


class FeatureEvent(BaseModel):
    """
    Feature event for tracking feature flag evaluations.
    """

    # Event kind
    kind: str = "feature"

    # User key
    user_key: str

    # Context kind
    context_kind: str

    # Flag key
    key: str

    # Variation type
    variation_type: Optional[str] = None

    # Value
    value: Optional[Any] = None

    # Default value
    default_value: Optional[Any] = None

    # Creation date
    creation_date: int

    # Evaluation context
    evaluation_context: Dict[str, Any]
