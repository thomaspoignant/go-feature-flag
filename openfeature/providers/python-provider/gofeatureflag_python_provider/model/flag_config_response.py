from datetime import datetime
from typing import Any, Dict, Optional
from pydantic import BaseModel

from gofeatureflag_python_provider.model.flag import Flag


class FlagConfigResponse(BaseModel):
    """
    Response from the flag configuration API.
    """

    # ETag for caching
    etag: Optional[str] = None

    # Last update timestamp
    last_updated: Optional[datetime] = None

    # Flags configuration
    flags: Optional[Dict[str, Flag]] = None

    # Evaluation context enrichment
    evaluation_context_enrichment: Optional[Dict[str, Any]] = None
