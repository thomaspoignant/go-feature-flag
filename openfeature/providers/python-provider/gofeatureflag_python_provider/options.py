import urllib3
import typing
from pydantic import AnyHttpUrl, BaseModel as PydanticBaseModel, ConfigDict


class BaseModel(PydanticBaseModel):
    model_config: ConfigDict = ConfigDict({"arbitrary_types_allowed": True})


class GoFeatureFlagOptions(BaseModel):
    endpoint: AnyHttpUrl
    urllib3_pool_manager: typing.Optional[urllib3.PoolManager] = None
