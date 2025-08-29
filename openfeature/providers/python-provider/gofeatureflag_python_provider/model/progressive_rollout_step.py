from typing import Any, Dict, Optional
from pydantic import BaseModel


class ProgressiveRolloutStep(BaseModel):
    """
    Progressive rollout step.
    """

    # Step percentage
    percentage: Optional[float] = None

    # Step value
    value: Optional[Any] = None
