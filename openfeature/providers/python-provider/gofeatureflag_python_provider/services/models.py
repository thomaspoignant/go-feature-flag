"""Request/response models for the GO Feature Flag API service."""

from datetime import datetime
from typing import Any

from gofeatureflag_python_provider.options import BaseModel


class FlagConfigRequest(BaseModel):
    """Request body for POST /v1/flag/configuration."""

    flags: list[str] = []


class FlagConfigResponse(BaseModel):
    """Response from POST /v1/flag/configuration."""

    etag: str | None = None
    last_updated: datetime | None = None
    flags: dict[str, Any] = {}
    evaluation_context_enrichment: dict[str, Any] = {}
