"""
GO Feature Flag API client: retrieve flag configuration and send events to the data collector.
"""

import json
from email.utils import parsedate_to_datetime
from http import HTTPStatus
from typing import Any, Optional
from urllib.parse import urljoin

import urllib3

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import (
    FeatureEvent,
    RequestDataCollector,
)
from gofeatureflag_python_provider.exceptions import (
    DataCollectorError,
    FlagConfigurationUnavailableError,
    UnauthorizedError,
)
from gofeatureflag_python_provider.services.models import (
    FlagConfigRequest,
    FlagConfigResponse,
)

# --- API client ---

DEFAULT_TIMEOUT_SECONDS = 10.0


class GoFeatureFlagApi:
    """
    Client for the GO Feature Flag relay proxy API: flag configuration and data collector.
    """

    _endpoint: str = "http://localhost:1031"
    _timeout: float = DEFAULT_TIMEOUT_SECONDS
    _api_key: Optional[str] = None
    _http: urllib3.PoolManager = None

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        timeout: Optional[float] = None,
    ) -> None:
        if options is None:
            raise ValueError("Options cannot be null")
        self._endpoint = str(options.endpoint).rstrip("/")
        self._timeout = timeout if timeout is not None else DEFAULT_TIMEOUT_SECONDS
        self._api_key = options.api_key

        if options.urllib3_pool_manager is not None:
            self._http = options.urllib3_pool_manager
        else:
            self._http = urllib3.PoolManager(
                num_pools=100,
                timeout=urllib3.Timeout(
                    connect=self._timeout,
                    read=self._timeout,
                ),
                retries=urllib3.Retry(0),
            )

    def _headers(self, extra: Optional[dict[str, str]] = None) -> dict[str, str]:
        out: dict[str, str] = {"Content-Type": "application/json"}
        if self._api_key:
            out["X-API-Key"] = self._api_key
        if extra:
            out.update(extra)
        return out

    def retrieve_flag_configuration(
        self,
        etag: Optional[str] = None,
        flags: Optional[list[str]] = None,
    ) -> FlagConfigResponse:
        """
        Fetch flag configuration from the relay proxy.

        :param etag: If set, send If-None-Match header; server may return 304.
        :param flags: If set, request only these flag keys; empty or None = all flags.
        :return: Flag config response (etag, last_updated, flags, evaluation_context_enrichment).
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

        result = self._parse_flag_config_response(response)
        if status == HTTPStatus.NOT_MODIFIED:
            return result

        try:
            data = json.loads(response.data.decode("utf-8"))
        except (json.JSONDecodeError, UnicodeDecodeError) as e:
            raise FlagConfigurationUnavailableError(
                f"Failed to parse flag configuration response: {e}"
            ) from e

        result.flags = data.get("flags") or {}
        result.evaluation_context_enrichment = (
            data.get("evaluationContextEnrichment") or {}
        )
        return result

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
        """Build FlagConfigResponse from response headers (and empty body for 304)."""
        etag_header = response.headers.get("ETag")
        if etag_header and etag_header.startswith('"') and etag_header.endswith('"'):
            etag_header = etag_header[1:-1]
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
            flags={},
            evaluation_context_enrichment={},
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
        url = urljoin(self._endpoint + "/", "v1/data/collector")
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
