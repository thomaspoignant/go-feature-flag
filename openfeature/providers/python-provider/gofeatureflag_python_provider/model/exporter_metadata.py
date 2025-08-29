from typing import Dict, Union


class ExporterMetadata:
    """
    This class represents the exporter metadata that will be sent in your evaluation data collector
    """

    def __init__(self):
        self._metadata: Dict[str, Union[str, bool, int]] = {}

    def add(self, key: str, value: Union[str, bool, int]) -> "ExporterMetadata":
        """
        Add a metadata to the exporter
        :param key: the key of the metadata
        :param value: the value of the metadata
        :return: self for chaining
        """
        self._metadata[key] = value
        return self

    def as_object(self) -> Dict[str, Union[str, bool, int]]:
        """
        Return the metadata as an immutable object
        :return: the metadata as a dictionary
        """
        return dict(self._metadata)
