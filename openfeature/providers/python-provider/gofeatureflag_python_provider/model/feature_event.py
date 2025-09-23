from typing import Any, Dict, Optional
from pydantic import BaseModel, Field


class FeatureEvent(BaseModel):
    """
    Feature event for tracking feature flag evaluations.
    """

    # Event kind
    kind: str = "feature"

    # Context kind
    context_kind: str = Field(alias="contextKind")

    # User key
    user_key: str = Field(alias="userKey")

    # Creation date
    creation_date: int = Field(alias="creationDate")

    # Flag key
    key: str

    # Variation type
    variation: Optional[str] = Field(default=None, alias="variation")

    # Value
    value: Optional[Any] = None

    # Default
    default: bool

    # Version
    version: Optional[str] = None

    # Variation
    metadata: Optional[Dict[str, Any]] = None

    class Config:
        """Populate by name and exclude None values."""

        populate_by_name = True
        exclude_none = True
        exclude_unset = True
