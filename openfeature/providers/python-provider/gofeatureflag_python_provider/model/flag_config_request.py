from typing import Optional, List
from pydantic import BaseModel


class FlagConfigRequest(BaseModel):
    """
    Request for flag configuration.
    """

    # List of flags to retrieve, if not set or empty, we will retrieve all available flags.
    flags: Optional[List[str]] = None
