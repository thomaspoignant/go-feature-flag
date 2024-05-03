from typing import Optional, Union, TypeVar
from gofeatureflag_python_provider.options import BaseModel


JsonType = TypeVar("JsonType", bool, int, str, float, list, Union[dict, list])


class ResponseFlagEvaluation(BaseModel):
    errorCode: Optional[str] = None
    failed: bool
    reason: str
    trackEvents: Optional[bool] = None
    value: JsonType
    variationType: Optional[str] = None
    version: Optional[str] = None
    metadata: Optional[dict] = None
    cacheable: Optional[bool] = None
