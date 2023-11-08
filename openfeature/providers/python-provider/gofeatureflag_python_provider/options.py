import typing

import urllib3
from pydantic import AnyHttpUrl, BaseModel as PydanticBaseModel, ConfigDict


class BaseModel(PydanticBaseModel):
    model_config: ConfigDict = ConfigDict({"arbitrary_types_allowed": True})


class GoFeatureFlagOptions(BaseModel):
    endpoint: AnyHttpUrl
    urllib3_pool_manager: typing.Optional[urllib3.PoolManager] = None

    # flagCacheSize (optional) is the maximum number of flag events we keep in memory to cache your flags.
    # default: 10000
    cache_size: typing.Optional[int] = 10000
