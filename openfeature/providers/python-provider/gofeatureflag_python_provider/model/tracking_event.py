from typing import Any, Dict
from pydantic import BaseModel, Field


class TrackingEvent(BaseModel):
    """
    Tracking event for custom tracking.
    """

    # Event kind
    kind: str = "tracking"

    # User key
    user_key: str = Field(alias="userKey")

    # Context kind
    context_kind: str = Field(alias="contextKind")

    # Event key
    key: str

    # Tracking event details
    tracking_event_details: Dict[str, Any] = Field(alias="trackingEventDetails")

    # Creation date
    creation_date: int = Field(alias="creationDate")

    # Evaluation context
    evaluation_context: Dict[str, Any] = Field(alias="evaluationContext")

    class Config:
        """Populate by name and exclude None values."""

        populate_by_name = True
        exclude_none = True
        exclude_unset = True
