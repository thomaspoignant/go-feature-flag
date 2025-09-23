from datetime import datetime
import json
from typing import List, Optional, Union, Annotated

import urllib3
from urllib3.exceptions import HTTPError
from urllib3.poolmanager import PoolManager
from urllib3.response import HTTPResponse

from gofeatureflag_python_provider.exception import (
    FlagConfigurationEndpointNotFoundException,
    ImpossibleToRetrieveConfigurationException,
    ImpossibleToSendDataToCollectorException,
    InvalidOptionsException,
    UnauthorizedException,
)
from gofeatureflag_python_provider.model import (
    ExporterMetadata,
    ExporterRequest,
    FlagConfigRequest,
    FlagConfigResponse,
)
from gofeatureflag_python_provider.model.feature_event import FeatureEvent
from gofeatureflag_python_provider.model.tracking_event import TrackingEvent
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions
from pydantic import Field

# Define a discriminated union type for events
Event = Annotated[Union[TrackingEvent, FeatureEvent], Field(discriminator="kind")]

# HTTP constants
APPLICATION_JSON = "application/json"
BEARER_TOKEN = "Bearer "
HTTP_HEADER_AUTHORIZATION = "Authorization"
HTTP_HEADER_CONTENT_TYPE = "Content-Type"
HTTP_HEADER_ETAG = "ETag"
HTTP_HEADER_IF_NONE_MATCH = "If-None-Match"
HTTP_HEADER_LAST_MODIFIED = "Last-Modified"

# HTTP status codes
HTTP_STATUS = {
    "OK": 200,
    "NOT_MODIFIED": 304,
    "BAD_REQUEST": 400,
    "UNAUTHORIZED": 401,
    "FORBIDDEN": 403,
    "NOT_FOUND": 404,
}


class Api:
    """
    GOFeatureFlagApi is a class that provides methods to interact with the GO Feature Flag API.
    """

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        urllib3_pool_manager: Optional[PoolManager] = None,
        logger=None,
    ):
        """
        Constructor for Api.
        :param options: Options provided during the initialization of the provider
        :param urllib3_pool_manager: Optional custom urllib3 pool manager
        :param logger: Optional logger instance
        :raises InvalidOptionsException: when options are not provided
        """
        if not options:
            raise InvalidOptionsException("Options cannot be null")

        self.endpoint = str(options.endpoint).rstrip("/")
        self.api_key = options.api_key
        self.logger = logger

        # Use provided pool manager or create default one
        if urllib3_pool_manager:
            self.pool_manager = urllib3_pool_manager
        else:
            self.pool_manager = urllib3.PoolManager(
                num_pools=100,
                timeout=urllib3.Timeout(connect=10, read=10),
                retries=urllib3.Retry(0),
            )

    def retrieve_flag_configuration(
        self, etag: Optional[str] = None, flags: Optional[List[str]] = None
    ) -> FlagConfigResponse:
        """
        RetrieveFlagConfiguration is a method that retrieves the flag configuration from the GO Feature Flag API.
        :param etag: If provided, we call the API with "If-None-Match" header.
        :param flags: List of flags to retrieve, if not set or empty, we will retrieve all available flags.
        :return: A FlagConfigResponse returning the success data.
        :raises FlagConfigurationEndpointNotFoundException: if the endpoint is not reachable.
        :raises ImpossibleToRetrieveConfigurationException: if the endpoint is returning an error.
        """
        request_body = FlagConfigRequest(flags=flags or [])

        headers = {
            HTTP_HEADER_CONTENT_TYPE: APPLICATION_JSON,
        }

        # Adding the If-None-Match header if etag is provided
        if etag:
            headers[HTTP_HEADER_IF_NONE_MATCH] = etag

        # Add authorization header if API key is provided
        if self.api_key:
            headers[HTTP_HEADER_AUTHORIZATION] = f"{BEARER_TOKEN}{self.api_key}"

        try:
            response = self.pool_manager.request(
                method="POST",
                url=f"{self.endpoint}/v1/flag/configuration",
                headers=headers,
                body=request_body.model_dump_json(by_alias=True),
            )

            return self._handle_flag_configuration_response(response)
        except (HTTPError, TimeoutError) as error:
            raise ImpossibleToRetrieveConfigurationException(
                f"Network error: {error}"
            ) from error

    def send_event_to_data_collector(
        self,
        events_list: List[Union[TrackingEvent, FeatureEvent]],
        exporter_metadata: ExporterMetadata,
    ) -> None:
        """
        Sends a list of events to the GO Feature Flag data collector.
        :param events_list: List of events
        :param exporter_metadata: Metadata associated.
        :raises UnauthorizedException: when we are not authorized to call the API
        :raises ImpossibleToSendDataToCollectorException: when an error occurred when calling the API
        """
        request_body = ExporterRequest(
            meta=exporter_metadata.as_object() if exporter_metadata else {},
            events=events_list,
        )

        headers = {
            HTTP_HEADER_CONTENT_TYPE: APPLICATION_JSON,
        }

        # Add authorization header if API key is provided
        if self.api_key:
            headers[HTTP_HEADER_AUTHORIZATION] = f"{BEARER_TOKEN}{self.api_key}"

        try:
            response = self.pool_manager.request(
                method="POST",
                url=f"{self.endpoint}/v1/data/collector",
                headers=headers,
                body=request_body.model_dump_json(by_alias=True, exclude_none=True),
            )

            self._handle_data_collector_response(response, len(events_list))

        except (HTTPError, TimeoutError) as error:
            raise ImpossibleToSendDataToCollectorException(
                f"Network error: {error}"
            ) from error

    def _handle_flag_configuration_response(
        self, response: HTTPResponse
    ) -> FlagConfigResponse:
        """
        Handle the response from the flag configuration request.
        :param response: HTTP response.
        :return: A FlagConfigResponse object.
        """
        if response.status == HTTP_STATUS["NOT_FOUND"]:
            raise FlagConfigurationEndpointNotFoundException()
        elif response.status in [HTTP_STATUS["UNAUTHORIZED"], HTTP_STATUS["FORBIDDEN"]]:
            raise UnauthorizedException(
                "Impossible to retrieve flag configuration: authentication/authorization error"
            )
        elif response.status == HTTP_STATUS["BAD_REQUEST"]:
            body = _decode_response_data(response)
            raise ImpossibleToRetrieveConfigurationException(
                f"retrieve flag configuration error: Bad request: {body}"
            )
        elif response.status not in [HTTP_STATUS["OK"], HTTP_STATUS["NOT_MODIFIED"]]:
            body = _decode_response_data(response)
            raise ImpossibleToRetrieveConfigurationException(
                f"retrieve flag configuration error: unexpected http code {response.status}: {body}"
            )

        return self._handle_flag_configuration_success(response)

    def _handle_data_collector_response(
        self, response: HTTPResponse, events_count: int
    ) -> None:
        """
        Handle the response from the data collector request.
        :param response: HTTP response.
        :param events_count: Number of events sent.
        """
        if response.status == HTTP_STATUS["OK"]:
            body = _decode_response_data(response)
            if self.logger:
                self.logger.info(
                    f"Published {events_count} events successfully: {body}"
                )
            return
        elif response.status in [HTTP_STATUS["UNAUTHORIZED"], HTTP_STATUS["FORBIDDEN"]]:
            raise UnauthorizedException(
                "Impossible to send events: authentication/authorization error"
            )
        elif response.status == HTTP_STATUS["BAD_REQUEST"]:
            body = _decode_response_data(response)
            raise ImpossibleToSendDataToCollectorException(f"Bad request: {body}")
        else:
            body = _decode_response_data(response)
            raise ImpossibleToSendDataToCollectorException(
                f"send data to the collector error: unexpected http code {response.status}: {body}"
            )

    def _handle_flag_configuration_success(
        self, response: HTTPResponse
    ) -> FlagConfigResponse:
        """
        Handle the success response of the flag configuration request.
        :param response: HTTP response.
        :return: A FlagConfigResponse object.
        """
        etag_header = response.headers.get(HTTP_HEADER_ETAG)
        last_modified_header = response.headers.get(HTTP_HEADER_LAST_MODIFIED)

        last_updated = datetime.now()
        if last_modified_header:
            try:
                last_updated = datetime.fromisoformat(
                    last_modified_header.replace("Z", "+00:00")
                )
            except ValueError:
                last_updated = datetime.now()

        result = FlagConfigResponse(
            etag=etag_header,
            last_updated=last_updated,
            flags={},
            evaluation_context_enrichment={},
        )

        if response.status == HTTP_STATUS["NOT_MODIFIED"]:
            return result

        try:
            body = _decode_response_data(response)
            goff_resp = json.loads(body)
            result.evaluation_context_enrichment = goff_resp.get(
                "evaluationContextEnrichment", {}
            )
            result.flags = goff_resp.get("flags", {})
        except (json.JSONDecodeError, KeyError) as error:
            body = _decode_response_data(response)
            if self.logger:
                self.logger.warning(
                    f'Failed to parse flag configuration response: {error}. Response body: "{body}"'
                )
            # Return the default result with empty flags and enrichment

        return result


def _decode_response_data(response: HTTPResponse) -> str:
    """
    Decode the response data.
    :param response: HTTP response.
    :return: The decoded response data.
    """
    return (
        response.data.decode("utf-8")
        if isinstance(response.data, bytes)
        else (response.data or "")
    )
