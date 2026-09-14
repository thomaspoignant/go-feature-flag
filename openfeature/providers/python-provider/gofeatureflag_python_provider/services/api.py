"""
GO Feature Flag API client: retrieve flag configuration and send events to the data collector.
"""

import json
import logging
from email.utils import parsedate_to_datetime
from http import HTTPStatus
from typing import Any, Optional
from urllib.parse import urljoin

import urllib3

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model.request_data_collector import (
    FeatureEvent,
    RequestDataCollector,
)
from gofeatureflag_python_provider.exceptions import (
    DataCollectorError,
    FlagConfigurationUnavailableError,
    UnauthorizedError,
)
from gofeatureflag_python_provider.services.model import (
    FlagConfigRequest,
    FlagConfigResponse,
)

# --- API client ---

DEFAULT_TIMEOUT_SECONDS = 10.0
DEFAULT_NUM_POOLS = 100

logger = logging.getLogger(__name__)


class GoFeatureFlagApi:
    """
    Client for the GO Feature Flag relay proxy API: flag configuration and data collector.
    """

    _endpoint: str
    _data_collector_endpoint: Optional[str] = None
    _timeout: float = DEFAULT_TIMEOUT_SECONDS
    _api_key: Optional[str] = None
    _http: urllib3.PoolManager = None

    def __init__(
        self,
        options: GoFeatureFlagOptions,
    ) -> None:
        if options is None:
            raise ValueError("Options cannot be null")
        self._endpoint = str(options.endpoint).rstrip("/")
        # The data collector may sit behind a different base entirely — scheme,
        # host, port and path prefix — while flag configuration and evaluation
        # keep using endpoint. Authentication, headers and timeout apply to both
        # identically.
        self._data_collector_endpoint = (
            str(options.data_collector_base_url).rstrip("/")
            if options.data_collector_base_url is not None
            else self._endpoint
        )
        self._timeout = (
            (options.timeout / 1000.0)
            if options.timeout is not None
            else DEFAULT_TIMEOUT_SECONDS
        )
        self._api_key = options.api_key
        self._custom_headers = dict(options.custom_headers or {})

        if options.urllib3_pool_manager is not None:
            self._http = options.urllib3_pool_manager
        else:
            # Create a default pool manager with the given timeout
            self._http = urllib3.PoolManager(
                num_pools=DEFAULT_NUM_POOLS,
                timeout=urllib3.Timeout(
                    connect=self._timeout,
                    read=self._timeout,
                ),
                retries=urllib3.Retry(0),
            )

    def _headers(self) -> dict[str, str]:
        """
        Return the headers for the API request.

        :return: A dictionary of headers.
        """
        out: dict[str, str] = dict(self._custom_headers)
        out["Content-Type"] = "application/json"
        if self._api_key:
            out["X-API-Key"] = f"{self._api_key}"
        return out

    def retrieve_flag_configuration(
        self,
        etag: Optional[str] = None,
        flags: Optional[list[str]] = None,
    ) -> Optional[FlagConfigResponse]:
        """
        Fetch flag configuration from the relay proxy.

        :param etag: If set, send If-None-Match header; server may return 304.
        :param flags: If set, request only these flag keys; empty or None = all flags.
        :return: Flag config response (etag, last_updated, flags,
            evaluation_context_enrichment), or None when the server answered
            304 Not Modified and there is therefore nothing to apply.
        :raises UnauthorizedError: 401/403.
        :raises FlagConfigurationUnavailableError: 404, 400, 5xx, or network error.
        """
        url = urljoin(self._endpoint + "/", "v1/flag/configuration")
        body = FlagConfigRequest(flags=flags or [])
        body_json = body.model_dump_json()
        headers = self._headers()
        if etag:
            headers["If-None-Match"] = etag

        try:
            response = self._http.request(
                method="POST",
                url=url,
                headers=headers,
                body=body_json,
            )
        except Exception as e:
            raise FlagConfigurationUnavailableError(f"Network error: {e}") from e

        status = int(response.status)
        self._raise_for_flag_config_status(status, response.data)

        if status == HTTPStatus.NOT_MODIFIED:
            # Return before a response object is built at all, so a "not modified"
            # answer is structurally incapable of carrying flags, an enrichment
            # map or an ETag that a caller could mistake for fresh state.
            return None

        return self._parse_flag_config_response(response)

    def _raise_for_flag_config_status(self, status: int, data: bytes) -> None:
        """Raise appropriate exception for non-success flag config status."""
        if status in {HTTPStatus.UNAUTHORIZED, HTTPStatus.FORBIDDEN}:
            raise UnauthorizedError(
                "Impossible to retrieve flag configuration: authentication/authorization error"
            )
        if status == HTTPStatus.NOT_FOUND:
            raise FlagConfigurationUnavailableError(
                "Flag configuration endpoint not found"
            )
        body_text = data.decode("utf-8", errors="replace")
        if status == HTTPStatus.BAD_REQUEST:
            raise FlagConfigurationUnavailableError(
                f"retrieve flag configuration error: Bad request: {body_text}"
            )
        if status >= 500:
            raise FlagConfigurationUnavailableError(
                f"retrieve flag configuration error: unexpected http code {status}: {body_text}"
            )

    def _parse_flag_config_response(
        self, response: urllib3.HTTPResponse
    ) -> FlagConfigResponse:
        """Build a complete FlagConfigResponse from the HTTP response."""
        try:
            data = json.loads(response.data.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError) as e:
            raise FlagConfigurationUnavailableError(
                f"Failed to parse flag configuration response: {e}"
            ) from e

        flags = data.get("flags")
        if not isinstance(flags, dict):
            # A response without a flag map is malformed. An empty map is valid
            # when the relay proxy legitimately serves no flags.
            raise FlagConfigurationUnavailableError(
                "flag configuration response does not contain a flag map"
            )

        # Stored verbatim, surrounding quotes included. The relay proxy issues
        # strong validators, so echoing a dequoted value back as If-None-Match
        # never matches and the server would answer 200 on every poll.
        etag_header = response.headers.get("ETag")
        last_modified = response.headers.get("Last-Modified")
        last_updated = None
        if last_modified:
            try:
                last_updated = parsedate_to_datetime(last_modified)
            except (TypeError, ValueError):
                pass
        return FlagConfigResponse(
            etag=etag_header,
            last_updated=last_updated,
            flags=flags,
            evaluation_context_enrichment=(
                data.get("evaluationContextEnrichment") or {}
            ),
        )

    def send_event_to_data_collector(
        self,
        events: list[FeatureEvent],
        exporter_metadata: Optional[dict[str, Any]] = None,
    ) -> None:
        """
        Send evaluation events to the GO Feature Flag data collector.

        :param events: List of feature events to send.
        :param exporter_metadata: Optional meta object sent with the payload.
        :raises UnauthorizedError: 401/403.
        :raises DataCollectorError: 400, 5xx, or network error.
        """
        url = urljoin(self._data_collector_endpoint + "/", "v1/data/collector")
        payload = RequestDataCollector(
            meta=exporter_metadata or {},
            events=events,
        )
        body_json = payload.model_dump_json()
        headers = self._headers()

        try:
            response = self._http.request(
                method="POST",
                url=url,
                headers=headers,
                body=body_json,
            )
        except Exception as e:
            raise DataCollectorError(f"Network error: {e}") from e

        status = int(response.status)

        if status in (HTTPStatus.UNAUTHORIZED, HTTPStatus.FORBIDDEN):
            raise UnauthorizedError(
                "Impossible to send events: authentication/authorization error"
            )
        if status == HTTPStatus.BAD_REQUEST:
            body_text = response.data.decode("utf-8", errors="replace")
            raise DataCollectorError(f"Bad request: {body_text}")
        if status != HTTPStatus.OK:
            body_text = response.data.decode("utf-8", errors="replace")
            raise DataCollectorError(
                f"send data to the collector error: unexpected http code {status}: {body_text}"
            )

        logger.info("Successfully sent %d event(s) to the data collector", len(events))
