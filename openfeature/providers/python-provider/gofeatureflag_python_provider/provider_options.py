from typing import Optional

from pydantic import (
    AnyHttpUrl,
    BaseModel as PydanticBaseModel,
    ConfigDict,
    Field,
    field_validator,
)
from urllib3.poolmanager import PoolManager

from gofeatureflag_python_provider.model.exporter_metadata import ExporterMetadata

from .model import EvaluationType


class BaseModel(PydanticBaseModel):
    """Base model with arbitrary types allowed for Pydantic validation."""

    model_config: ConfigDict = ConfigDict(arbitrary_types_allowed=True)


class GoFeatureFlagOptions(BaseModel):
    """
    Options for the GO Feature Flag provider.
    """

    model_config = {"arbitrary_types_allowed": True}

    # The endpoint of the GO Feature Flag relay-proxy.
    endpoint: AnyHttpUrl = Field(
        description="The endpoint of the GO Feature Flag relay-proxy",
        examples=["http://localhost:1031", "https://api.gofeatureflag.com"],
    )

    # The type of evaluation to use.
    evaluation_type: EvaluationType = Field(
        default=EvaluationType.IN_PROCESS, description="The type of evaluation to use"
    )

    # api_key (optional) If the relay proxy is configured to authenticate the requests, you should provide
    # an API Key to the provider. Please ask the administrator of the relay proxy to provide an API Key.
    # Default: None
    api_key: Optional[str] = Field(
        default=None, description="API key for relay proxy authentication", min_length=1
    )

    # ExporterMetadata (optional) is the metadata we send to the GO Feature Flag relay proxy when we report the
    # evaluation data usage.
    #
    # ‼️Important: If you are using a GO Feature Flag relay proxy before version v1.41.0, the information of this
    # field will not be added to your feature events.
    exporter_metadata: Optional[ExporterMetadata] = Field(
        default_factory=dict,
        description="Metadata sent to relay proxy for evaluation data usage",
    )

    # disableDataCollection set to true if you don't want to collect the usage of flags retrieved in the cache.
    # default: false
    disable_data_collection: bool = Field(
        default=False, description="Disable data collection for flag usage"
    )

    # The interval for flushing data collection events in milliseconds.
    data_flush_interval: Optional[int] = Field(
        default=60000,
        gt=0,
        description="Data flush interval in milliseconds",
        examples=[30000, 60000, 120000],
    )

    # The maximum number of pending events before flushing.
    max_pending_events: Optional[int] = Field(
        default=10000,
        gt=0,
        description="Maximum number of pending events before flushing",
        examples=[5000, 10000, 20000],
    )

    # The interval for polling flag configuration changes in milliseconds.
    flag_change_polling_interval_ms: Optional[int] = Field(
        default=120000,
        gt=0,
        description="Flag change polling interval in milliseconds",
        examples=[60000, 120000, 300000],
    )

    # ADVANCED OPTIONS --- be careful when changing these options

    # http_client (optional) is the http client used to call the relay proxy.
    urllib3_pool_manager: Optional[PoolManager] = Field(
        default=None, description="Custom urllib3 pool manager for HTTP requests"
    )

    @field_validator("api_key")
    @classmethod
    def validate_api_key(cls, v):
        """Validate API key if provided."""
        if v is not None and len(v.strip()) == 0:
            raise ValueError("API key cannot be empty string")
        return v

    @field_validator("exporter_metadata")
    @classmethod
    def validate_exporter_metadata(cls, v):
        """Validate exporter metadata."""
        if v is not None and not isinstance(v, ExporterMetadata):
            raise ValueError("exporter_metadata must be an ExporterMetadata instance")
        return v or {}
