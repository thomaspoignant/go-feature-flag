from dataclasses import dataclass

from openfeature.provider.metadata import Metadata


@dataclass
class GoFeatureFlagMetadata(Metadata):
    name: str = "GO Feature Flag"
