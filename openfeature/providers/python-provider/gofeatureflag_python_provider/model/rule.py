from typing import Any, Dict, List, Optional
from pydantic import BaseModel


class Rule(BaseModel):
    """
    Rule for flag evaluation.
    """

    # Rule name
    name: Optional[str] = None

    # Rule conditions
    conditions: Optional[List[Dict[str, Any]]] = None

    # Rule value
    value: Optional[Any] = None
