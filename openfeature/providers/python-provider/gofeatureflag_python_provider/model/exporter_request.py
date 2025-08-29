from typing import Any, Dict, List
from pydantic import BaseModel


class ExporterRequest(BaseModel):
    """
    Request for the exporter.
    """

    # Metadata that will be sent in your evaluation data collector
    meta: Dict[str, Any]

    # Events to export
    events: List[Dict[str, Any]]
