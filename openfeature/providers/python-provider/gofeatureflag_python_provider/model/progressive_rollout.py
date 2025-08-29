from typing import Any, Dict, List, Optional
from pydantic import BaseModel


class ProgressiveRollout(BaseModel):
    """
    Progressive rollout configuration.
    """

    # Rollout steps
    steps: Optional[List[Dict[str, Any]]] = None

    # Rollout value
    value: Optional[Any] = None
