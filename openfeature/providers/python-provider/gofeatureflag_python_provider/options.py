import logging
import typing
from enum import Enum

import urllib3
from pydantic import AnyHttpUrl, BaseModel as PydanticBaseModel, ConfigDict


class EvaluationType(str, Enum):
    REMOTE = "remote"
    INPROCESS = "inprocess"


class BaseModel(PydanticBaseModel):
    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)


class GoFeatureFlagOptions(BaseModel):
    # evaluation_type selects how flags are evaluated: remote (relay proxy) or inprocess (local/WASM).
    # default: REMOTE
    evaluation_type: EvaluationType = EvaluationType.INPROCESS

    # endpoint is the endpoint of the relay proxy.
    # example: http://localhost:1031
    endpoint: AnyHttpUrl

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

    # reconnectInterval (optional) interval time (in seconds) we use to reconnect to the server if the \
    # connection is stopped.
    # default: 1 minute
    reconnect_interval: typing.Optional[int] = 60

    # flag_config_poll_interval_seconds (optional) interval in seconds to poll flag configuration.
    # Used only by InProcessEvaluator. default: 10
    flag_config_poll_interval_seconds: typing.Optional[int] = 10

    # ADVANCED OPTIONS --- be careful when changing these options

    # log_level (optional) logging level: "DEBUG", "INFO", "WARNING", "ERROR" or int (e.g. logging.DEBUG).
    # default: "WARNING"
    log_level: typing.Union[int, str] = "WARNING"

    # http_client (optional) is the http client used to call the relay proxy.
    urllib3_pool_manager: typing.Optional[urllib3.PoolManager] = None

    # disable_cache_invalidation (optional) set to true if you don't want to invalidate the cache when the remote
    # config changes.
    # default: false
    disable_cache_invalidation: typing.Optional[bool] = False

    # api_key (optional) If the relay proxy is configured to authenticate the requests, you should provide
    # an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key.
    # Default: None
    api_key: typing.Optional[str] = None

    # ExporterMetadata (optional) is the metadata we send to the GO Feature Flag relay proxy when we report the
    # evaluation data usage.
    #
    # ‼️Important: If you are using a GO Feature Flag relay proxy before version v1.41.0, the information of this
    # field will not be added to your feature events.
    exporter_metadata: typing.Optional[dict] = {}

    # max_pending_events (optional) is the maximum number of events buffered in memory before an immediate
    # flush is triggered (fire-and-forget). Used by EventPublisher.
    # default: 10000
    max_pending_events: typing.Optional[int] = 10_000

    # wasm_file_path (optional) is the path to the GO Feature Flag evaluation WASI binary.
    # Used only when evaluation_type is INPROCESS.
    # If not set, the bundled wasm-releases/evaluation/gofeatureflag-evaluation_0.2.0.wasi is used.
    wasm_file_path: typing.Optional[str] = None

    def get_log_level_int(self) -> int:
        """Resolve log_level to a logging module level constant."""
        if self.log_level is None:
            return logging.WARNING
        if isinstance(self.log_level, int):
            return self.log_level
        return getattr(logging, str(self.log_level).upper(), logging.WARNING)
