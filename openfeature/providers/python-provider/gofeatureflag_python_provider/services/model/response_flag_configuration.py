from __future__ import annotations

from datetime import datetime
from typing import Any, Dict, Optional

from gofeatureflag_python_provider.options import BaseModel


class FlagConfigResponse(BaseModel):
    """Response from POST /v1/flag/configuration."""

    etag: Optional[str] = None
    last_updated: Optional[datetime] = None
    flags: Dict[str, Any] = {}
    evaluation_context_enrichment: Dict[str, Any] = {}
