from typing import Any, Dict
from pydantic import BaseModel


class Event(BaseModel):
    """
    Base event class.
    """

    # Event kind
    kind: str

    # Event data
    data: Dict[str, Any]
