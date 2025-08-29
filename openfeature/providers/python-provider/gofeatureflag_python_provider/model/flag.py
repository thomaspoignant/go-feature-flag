from typing import Any, Dict, Optional
from pydantic import BaseModel


class Flag(BaseModel):
    """
    Represents a feature flag configuration.
    """

    # Track events flag
    track_events: Optional[bool] = None

    # Additional flag data
    data: Optional[Dict[str, Any]] = None
