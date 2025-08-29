from typing import Any, Dict, Optional
from pydantic import BaseModel


class FlagBase(BaseModel):
    """
    Base flag configuration.
    """

    # Flag key
    key: str

    # Flag value
    value: Any

    # Flag metadata
    metadata: Optional[Dict[str, Any]] = None
