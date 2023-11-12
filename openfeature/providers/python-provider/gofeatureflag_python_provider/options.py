import typing

import urllib3
from pydantic import AnyHttpUrl, BaseModel as PydanticBaseModel, ConfigDict


class BaseModel(PydanticBaseModel):
    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)


class GoFeatureFlagOptions(BaseModel):
    endpoint: AnyHttpUrl
    urllib3_pool_manager: typing.Optional[urllib3.PoolManager] = None

    # flagCacheSize (optional) is the maximum number of flag events we keep in memory to cache your flags.
    # default: 10000
    cache_size: typing.Optional[int] = 10000

    # dataFlushInterval (optional) interval time (in millisecond) we use to call the relay proxy to collect data.
    # The parameter is used only if the cache is enabled, otherwise the collection of the data is done directly
    # when calling the evaluation API.
    # default: 1 minute
    data_flush_interval: typing.Optional[int] = 60000

    # disableDataCollection set to true if you don't want to collect the usage of flags retrieved in the cache.
    # default: false
    disable_data_collection: typing.Optional[bool] = False
