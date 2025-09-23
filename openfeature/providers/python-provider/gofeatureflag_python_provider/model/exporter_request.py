from typing import Any, Dict, List, Union
from pydantic import BaseModel
from gofeatureflag_python_provider.model.tracking_event import TrackingEvent
from gofeatureflag_python_provider.model.feature_event import FeatureEvent


class ExporterRequest(BaseModel):
    """
    Request for the exporter.
    """

    # Metadata that will be sent in your evaluation data collector
    meta: Dict[str, Any]

    # Events to export
    events: List[Union[TrackingEvent, FeatureEvent]]

    class Config:
        """Populate by name and exclude None values."""

        populate_by_name = True
        exclude_none = True
        exclude_unset = True
