from dataclasses import dataclass

from openfeature.provider.metadata import Metadata


@dataclass
class GoFeatureFlagMetadata(Metadata):
    def __init__(self):
        pass

    name: str = "GO Feature Flag"
