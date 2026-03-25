"""Request/response models for the GO Feature Flag API service."""

from __future__ import annotations

from datetime import datetime
from typing import Any, Dict, List, Optional

from gofeatureflag_python_provider.options import BaseModel


class FlagConfigRequest(BaseModel):
    """Request body for POST /v1/flag/configuration."""

    flags: List[str] = []


class FlagConfigResponse(BaseModel):
    """Response from POST /v1/flag/configuration."""

    etag: Optional[str] = None
    last_updated: Optional[datetime] = None
    flags: Dict[str, Any] = {}
    evaluation_context_enrichment: Dict[str, Any] = {}
