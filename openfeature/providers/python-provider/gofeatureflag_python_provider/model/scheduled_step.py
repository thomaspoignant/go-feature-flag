from typing import Any, Optional
from pydantic import BaseModel


class ScheduledStep(BaseModel):
    """
    Scheduled step configuration.
    """

    # Step value
    value: Optional[Any] = None
