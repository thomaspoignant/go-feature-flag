from typing import Any, Dict
from pydantic import BaseModel


class TrackingEvent(BaseModel):
    """
    Tracking event for custom tracking.
    """

    # Event kind
    kind: str = "tracking"

    # User key
    user_key: str

    # Context kind
    context_kind: str

    # Event key
    key: str

    # Tracking event details
    tracking_event_details: Dict[str, Any]

    # Creation date
    creation_date: int

    # Evaluation context
    evaluation_context: Dict[str, Any]
