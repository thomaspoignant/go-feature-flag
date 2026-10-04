from __future__ import annotations

from typing import List

from gofeatureflag_python_provider.options import BaseModel


class FlagConfigRequest(BaseModel):
    """Request body for POST /v1/flag/configuration."""

    flags: List[str] = []
