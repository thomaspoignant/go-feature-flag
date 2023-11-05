import urllib3
import typing
from pydantic import BaseModel as PydanticBaseModel, AnyHttpUrl


class BaseModel(PydanticBaseModel):
    class Config:
        arbitrary_types_allowed = True


class GoFeatureFlagOptions(BaseModel):
    endpoint: AnyHttpUrl
    urllib3PoolManager: typing.Optional[urllib3.PoolManager]
