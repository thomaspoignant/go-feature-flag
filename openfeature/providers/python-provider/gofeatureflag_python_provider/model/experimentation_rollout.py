from typing import Any, Dict, Optional
from pydantic import BaseModel


class ExperimentationRollout(BaseModel):
    """
    Experimentation rollout configuration.
    """

    # Rollout value
    value: Optional[Any] = None
