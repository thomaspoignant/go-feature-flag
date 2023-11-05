from pydantic.generics import GenericModel
from typing import Optional, Generic, Union, TypeVar

JsonType = TypeVar("JsonType", bool, int, str, float, list, Union[dict, list])


class ResponseFlagEvaluation(GenericModel, Generic[JsonType]):
    errorCode: Optional[str]
    failed: bool
    reason: str
    trackEvents: bool
    value: JsonType
    variationType: Optional[str]
    version: Optional[str]
    metadata: Optional[dict]
