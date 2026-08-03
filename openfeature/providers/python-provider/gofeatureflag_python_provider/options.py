import logging
import os
import typing
from enum import Enum

import urllib3
from pydantic import AnyHttpUrl, BaseModel as PydanticBaseModel, ConfigDict, Field


class EvaluationType(str, Enum):
    REMOTE = "remote"
    INPROCESS = "inprocess"


class BaseModel(PydanticBaseModel):
    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)


class GoFeatureFlagOptions(BaseModel):
    # evaluation_type selects how flags are evaluated: remote (relay proxy) or inprocess (local/WASM).
    # default: INPROCESS
    evaluation_type: EvaluationType = EvaluationType.INPROCESS

    # endpoint is the endpoint of the relay proxy.
    # example: http://localhost:1031
    endpoint: AnyHttpUrl

    # data_collector_base_url (optional) overrides the base URL of the data collector
    # only; flag configuration and evaluation keep using endpoint. It replaces the
    # whole base, including scheme, host, port and path prefix, and authentication,
    # headers and timeout apply to it identically.
    # default: endpoint
    data_collector_base_url: typing.Optional[AnyHttpUrl] = None

    # timeout (optional) in milliseconds, applied to every request to the relay proxy:
    # flag configuration, remote evaluation and data collection.
    # default: 10000
    timeout: typing.Optional[int] = 10_000

    # dataFlushInterval (optional) interval time (in millisecond) we use to call the relay proxy to collect data.
    # The parameter is used only if the cache is enabled, otherwise the collection of the data is done directly
    # when calling the evaluation API.
    # default: 1 minute
    data_flush_interval: typing.Optional[int] = 60000

    # disableDataCollection set to true if you don't want to collect the usage of flags retrieved in the cache.
    # default: false
    disable_data_collection: typing.Optional[bool] = False

    # flag_config_poll_interval_seconds (optional) interval in seconds to poll flag configuration.
    # Used only by InProcessEvaluator.
    # default: 120 seconds
    flag_config_poll_interval_seconds: typing.Optional[int] = 120

    # evaluation_flag_list (optional) restricts the flag configuration fetched from
    # the relay proxy to these keys. Unset or empty means every flag.
    # Used only when evaluation_type is INPROCESS.
    # default: empty (all flags)
    evaluation_flag_list: typing.Optional[typing.List[str]] = None

    # custom_headers (optional) extra headers added to every request to the relay
    # proxy, for deployments behind a gateway that needs its own authentication.
    # They are applied before the provider's own headers, so a configured api_key
    # always wins over a custom Authorization header.
    # default: none
    custom_headers: typing.Optional[dict[str, str]] = None

    # ADVANCED OPTIONS --- be careful when changing these options

    # log_level (optional) logging level: "DEBUG", "INFO", "WARNING", "ERROR" or int (e.g. logging.DEBUG).
    # default: "WARNING"
    log_level: typing.Union[int, str] = "WARNING"

    # http_client (optional) is the http client used to call the relay proxy.
    urllib3_pool_manager: typing.Optional[urllib3.PoolManager] = None

    # api_key (optional) If the relay proxy is configured to authenticate the requests, you should provide
    # an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key.
    # Default: None
    api_key: typing.Optional[str] = None

    # ExporterMetadata (optional) is the metadata we send to the GO Feature Flag relay proxy when we report the
    # evaluation data usage. Values are restricted to string, boolean, integer or
    # float; anything else is rejected here rather than failing later inside the
    # publisher, where the batch would be re-queued and retried forever.
    #
    # ‼️Important: If you are using a GO Feature Flag relay proxy before version v1.41.0, the information of this
    # field will not be added to your feature events.
    exporter_metadata: typing.Optional[
        dict[str, typing.Union[str, bool, int, float]]
    ] = {}

    # max_pending_events (optional) is the maximum number of events buffered in memory before an immediate
    # flush is triggered (fire-and-forget). Used by EventPublisher.
    # default: 10000
    max_pending_events: typing.Optional[int] = 10_000

    # wasm_file_path (optional) is the path to the GO Feature Flag evaluation WASI binary.
    # Used only when evaluation_type is INPROCESS.
    # If not set, the bundled WASI binary is used (version pinned in wasm/_wasi_version.txt).
    wasm_file_path: typing.Optional[str] = None

    # wasm_pool_size (optional) number of WASM Store instances for concurrent in-process evaluation.
    # Used only when evaluation_type is INPROCESS. wasmtime.Store is not thread-safe; a pool
    # allows multiple evaluations to run in parallel.
    # default: the host's CPU core count
    wasm_pool_size: typing.Optional[int] = Field(
        default_factory=lambda: os.cpu_count() or 1
    )

    def get_exporter_metadata(self) -> dict:
        """Exporter metadata plus the reserved keys identifying this SDK.

        These are present whether or not any metadata was configured, so events
        can always be attributed to a provider and a language.
        """
        return {
            **(self.exporter_metadata or {}),
            "provider": "python",
            "openfeature": True,
        }

    def get_wasm_pool_size(self) -> int:
        """Resolve the WASM pool size, defaulting to the host's CPU core count."""
        if self.wasm_pool_size is not None and self.wasm_pool_size > 0:
            return self.wasm_pool_size
        return os.cpu_count() or 1

    def get_log_level_int(self) -> int:
        """Resolve log_level to a logging module level constant."""
        if self.log_level is None:
            return logging.WARNING
        if isinstance(self.log_level, int):
            return self.log_level
        return getattr(logging, str(self.log_level).upper(), logging.WARNING)
