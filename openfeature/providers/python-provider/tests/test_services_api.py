"""
Unit tests for GO Feature Flag API service layer (GoFeatureFlagApi).
HTTP is mocked by patching urllib3.PoolManager so no real requests are made.
"""

from __future__ import annotations

import json
from http import HTTPStatus
from unittest.mock import Mock, patch

import pytest

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model.request_data_collector import (
    FeatureEvent,
)
from gofeatureflag_python_provider.services import (
    DataCollectorError,
    FlagConfigurationUnavailableError,
    GoFeatureFlagApi,
    UnauthorizedError,
)


def _make_options(
    endpoint: str = "http://localhost:1031",
    api_key: str | None = None,
):
    return GoFeatureFlagOptions(endpoint=endpoint, api_key=api_key)


def _mock_response(status: int, body: bytes | str, headers: dict | None = None):
    """Build a mock HTTP response with .status, .data, .headers.get()."""
    data = body.encode("utf-8") if isinstance(body, str) else body
    resp = Mock()
    resp.status = status
    resp.data = data
    resp.headers = Mock()
    resp.headers.get = Mock(side_effect=(headers or {}).get)
    return resp


# --- retrieve_flag_configuration ---


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_success_returns_parsed_response(
    mock_pool_manager_class,
):
    """On 200, response body is parsed into FlagConfigResponse."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps(
            {
                "flags": {"my_flag": {"defaultValue": True}},
                "evaluationContextEnrichment": {"region": "eu"},
            }
        ),
        {"ETag": '"abc123"', "Last-Modified": "Wed, 18 Feb 2025 12:00:00 GMT"},
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    result = api.retrieve_flag_configuration()

    assert result.flags == {"my_flag": {"defaultValue": True}}
    assert result.evaluation_context_enrichment == {"region": "eu"}
    # Stored verbatim: the relay proxy issues strong validators, so the quotes
    # have to survive to be echoed back as If-None-Match.
    assert result.etag == '"abc123"'
    assert result.last_updated is not None
    call = mock_http.request.call_args
    assert call.kwargs["method"] == "POST"
    assert "v1/flag/configuration" in call.kwargs["url"]
    assert call.kwargs["headers"]["Content-Type"] == "application/json"
    body = json.loads(call.kwargs["body"])
    assert body["flags"] == []


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_sends_flags_filter_when_provided(
    mock_pool_manager_class,
):
    """When flags=[...] is passed, request body contains that list."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration(flags=["f1", "f2"])

    body = json.loads(mock_http.request.call_args.kwargs["body"])
    assert body["flags"] == ["f1", "f2"]


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_sends_etag_header_when_provided(
    mock_pool_manager_class,
):
    """When etag is passed, If-None-Match header is set."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration(etag="my-etag")

    headers = mock_http.request.call_args.kwargs["headers"]
    assert headers.get("If-None-Match") == "my-etag"


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_sends_api_key_header_when_api_key_set(
    mock_pool_manager_class,
):
    """When api_key is set, an X-API-Key header is sent."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options(api_key="secret-key")
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration()

    headers = mock_http.request.call_args.kwargs["headers"]
    assert headers.get("X-API-Key") == "secret-key"
    assert "Authorization" not in headers


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_no_authorization_header_when_api_key_not_provided(
    mock_pool_manager_class,
):
    """When api_key is not provided, Authorization/X-API-Key header is not sent (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options(api_key=None)
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration()

    headers = mock_http.request.call_args.kwargs["headers"]
    assert "X-API-Key" not in headers
    assert "Authorization" not in headers


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_calls_configuration_endpoint(
    mock_pool_manager_class,
):
    """Should call the configuration endpoint with POST and body { flags: [] } (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options(endpoint="http://localhost:8080")
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration()

    call = mock_http.request.call_args
    assert call.kwargs["url"] == "http://localhost:8080/v1/flag/configuration"
    assert call.kwargs["method"] == "POST"
    assert json.loads(call.kwargs["body"]) == {"flags": []}


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_includes_content_type_header(
    mock_pool_manager_class,
):
    """Should include Content-Type: application/json header (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}, "evaluationContextEnrichment": {}}),
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    api.retrieve_flag_configuration()

    assert (
        mock_http.request.call_args.kwargs["headers"].get("Content-Type")
        == "application/json"
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_304_returns_none(mock_pool_manager_class):
    """On 304 Not Modified the API returns None, even when an ETag is echoed.

    A "not modified" answer must be structurally incapable of carrying state a
    caller could apply: no flags, no enrichment and above all no ETag.
    """
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.NOT_MODIFIED,
        "",
        {"ETag": '"unchanged"', "Last-Modified": "Wed, 21 Oct 2015 07:28:00 GMT"},
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    result = api.retrieve_flag_configuration(etag='"unchanged"')

    assert result is None


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_401_raises_unauthorized(mock_pool_manager_class):
    """On 401, UnauthorizedError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.UNAUTHORIZED, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(UnauthorizedError) as exc_info:
        api.retrieve_flag_configuration()

    assert (
        "authentication" in str(exc_info.value).lower()
        or "authorization" in str(exc_info.value).lower()
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_403_raises_unauthorized(mock_pool_manager_class):
    """On 403, UnauthorizedError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.FORBIDDEN, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(UnauthorizedError):
        api.retrieve_flag_configuration()


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_404_raises_flag_config_unavailable(
    mock_pool_manager_class,
):
    """On 404, FlagConfigurationUnavailableError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.NOT_FOUND, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(FlagConfigurationUnavailableError) as exc_info:
        api.retrieve_flag_configuration()

    assert "not found" in str(exc_info.value).lower()


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_400_raises_flag_config_unavailable(
    mock_pool_manager_class,
):
    """On 400, FlagConfigurationUnavailableError is raised with body."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.BAD_REQUEST,
        '{"error":"invalid"}',
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(FlagConfigurationUnavailableError) as exc_info:
        api.retrieve_flag_configuration()

    assert "Bad request" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_500_raises_flag_config_unavailable(
    mock_pool_manager_class,
):
    """On 500, FlagConfigurationUnavailableError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(500, "server error")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(FlagConfigurationUnavailableError) as exc_info:
        api.retrieve_flag_configuration()

    assert "500" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_invalid_last_modified_returns_none(
    mock_pool_manager_class,
):
    """Should handle invalid Last-Modified header (per JS test: lastUpdated is NaN / None)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK,
        json.dumps({"flags": {}}),
        {"ETag": '"123456789"', "Last-Modified": "invalid-date"},
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    result = api.retrieve_flag_configuration()

    # Parsing invalid date leaves last_updated as None (JS: lastUpdated?.getTime() is NaN)
    assert result.last_updated is None


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_accepts_an_empty_flag_map(
    mock_pool_manager_class,
):
    """A relay proxy serving no flags returns a valid, empty configuration."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK, json.dumps({"flags": {}}), {"ETag": '"v1"'}
    )
    api = GoFeatureFlagApi(_make_options())

    result = api.retrieve_flag_configuration()

    assert result.flags == {}
    assert result.etag == '"v1"'


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
@pytest.mark.parametrize("body", ["{}", json.dumps({"flags": None})])
def test_retrieve_flag_configuration_rejects_a_response_without_a_flag_map(
    mock_pool_manager_class, body
):
    """A body with no flag map is malformed, and must not read as 'no flags'.

    Silently turning it into an empty configuration would drop every flag the
    provider is serving.
    """
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, body, {})
    api = GoFeatureFlagApi(_make_options())

    with pytest.raises(FlagConfigurationUnavailableError, match="flag map"):
        api.retrieve_flag_configuration()


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_network_error_raises_flag_config_unavailable(
    mock_pool_manager_class,
):
    """On network error from pool manager, FlagConfigurationUnavailableError is raised (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.side_effect = Exception("connection refused")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(FlagConfigurationUnavailableError) as exc_info:
        api.retrieve_flag_configuration()

    assert "Network error" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_retrieve_flag_configuration_timeout_raises_flag_config_unavailable(
    mock_pool_manager_class,
):
    """On timeout (request aborted / TimeoutError), FlagConfigurationUnavailableError is raised (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.side_effect = TimeoutError("timed out")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(FlagConfigurationUnavailableError) as exc_info:
        api.retrieve_flag_configuration()

    assert "Network error" in str(exc_info.value)


# --- send_event_to_data_collector ---


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_calls_data_collector_endpoint(
    mock_pool_manager_class,
):
    """Should call the data collector endpoint with POST (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "Success")
    options = _make_options(endpoint="http://localhost:8080")
    api = GoFeatureFlagApi(options)

    api.send_event_to_data_collector([])

    call = mock_http.request.call_args
    assert call.kwargs["url"] == "http://localhost:8080/v1/data/collector"
    assert call.kwargs["method"] == "POST"


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_includes_content_type_header(
    mock_pool_manager_class,
):
    """Should include Content-Type: application/json header (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    api.send_event_to_data_collector([])

    assert (
        mock_http.request.call_args.kwargs["headers"].get("Content-Type")
        == "application/json"
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_no_authorization_header_when_api_key_not_provided(
    mock_pool_manager_class,
):
    """When api_key is not provided, Authorization/X-API-Key header is not sent (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    options = _make_options(api_key=None)
    api = GoFeatureFlagApi(options)

    api.send_event_to_data_collector([])

    headers = mock_http.request.call_args.kwargs["headers"]
    assert "X-API-Key" not in headers
    assert "Authorization" not in headers


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_includes_api_key_when_provided(
    mock_pool_manager_class,
):
    """When api_key is set, Authorization/X-API-Key header is sent (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    options = _make_options(api_key="my-api-key")
    api = GoFeatureFlagApi(options)

    api.send_event_to_data_collector([])

    headers = mock_http.request.call_args.kwargs["headers"]
    assert (
        headers.get("X-API-Key") == "my-api-key"
        or headers.get("Authorization") == "Bearer my-api-key"
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_success(mock_pool_manager_class):
    """On 200, no exception is raised; events and metadata in body (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)
    events = [
        FeatureEvent(
            contextKind="user",
            userKey="u1",
            creationDate=1234567890,
            key="flag1",
            variation="True",
            value=True,
            default=False,
        ),
    ]

    api.send_event_to_data_collector(events)

    call = mock_http.request.call_args
    assert call.kwargs["method"] == "POST"
    assert "v1/data/collector" in call.kwargs["url"]
    assert call.kwargs["headers"]["Content-Type"] == "application/json"
    body = json.loads(call.kwargs["body"])
    assert body["meta"] == {}
    assert len(body["events"]) == 1
    assert body["events"][0]["key"] == "flag1"


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_sends_metadata(mock_pool_manager_class):
    """exporter_metadata is sent as meta in the payload."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    api.send_event_to_data_collector([], exporter_metadata={"provider": "python"})

    body = json.loads(mock_http.request.call_args.kwargs["body"])
    assert body["meta"] == {"provider": "python"}


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_401_raises_unauthorized(mock_pool_manager_class):
    """On 401, UnauthorizedError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.UNAUTHORIZED, "")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(UnauthorizedError):
        api.send_event_to_data_collector([])


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_400_raises_data_collector_error(
    mock_pool_manager_class,
):
    """On 400, DataCollectorError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.BAD_REQUEST,
        "invalid payload",
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(DataCollectorError) as exc_info:
        api.send_event_to_data_collector([])

    assert "Bad request" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_500_raises_data_collector_error(
    mock_pool_manager_class,
):
    """On 500, DataCollectorError is raised."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(500, "server error")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(DataCollectorError) as exc_info:
        api.send_event_to_data_collector([])

    assert "500" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_network_error_raises_data_collector_error(
    mock_pool_manager_class,
):
    """On network error from pool manager, DataCollectorError is raised (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.side_effect = Exception("Network error")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(DataCollectorError) as exc_info:
        api.send_event_to_data_collector([])

    assert "Network error" in str(exc_info.value)


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_send_event_to_data_collector_timeout_raises_data_collector_error(
    mock_pool_manager_class,
):
    """On timeout (request aborted / TimeoutError), DataCollectorError is raised (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.side_effect = TimeoutError("timed out")
    options = _make_options()
    api = GoFeatureFlagApi(options)

    with pytest.raises(DataCollectorError) as exc_info:
        api.send_event_to_data_collector([])

    assert "Network error" in str(exc_info.value)


# --- constructor ---


def test_constructor_raises_when_options_null():
    """Passing None for options raises ValueError."""
    with pytest.raises(ValueError) as exc_info:
        GoFeatureFlagApi(None)  # type: ignore[arg-type]

    assert "null" in str(exc_info.value).lower()


# --- authentication ---


def test_api_key_is_sent_as_a_api_key_header():
    """The relay proxy accepts both, but X-API-Key is the contracted scheme."""
    api = GoFeatureFlagApi(_make_options(api_key="secret-key"))

    headers = api._headers()

    assert headers["X-API-Key"] == "secret-key"
    assert "Authorization" not in headers


def test_no_authentication_header_when_api_key_is_unset():
    api = GoFeatureFlagApi(_make_options())

    headers = api._headers()

    assert "Authorization" not in headers
    assert "X-API-Key" not in headers


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_bearer_header_is_applied_to_the_data_collector(mock_pool_manager_class):
    """Authentication must reach every authenticated endpoint, not just evaluation."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    api = GoFeatureFlagApi(_make_options(api_key="secret-key"))

    api.send_event_to_data_collector([], {"provider": "python"})

    headers = mock_http.request.call_args.kwargs["headers"]
    assert headers["X-API-Key"] == "secret-key"


# --- timeout and data collector base URL ---


def test_configured_timeout_is_applied_in_seconds():
    """The option is milliseconds; urllib3 wants seconds."""
    assert GoFeatureFlagApi(_make_options())._timeout == 10.0
    assert (
        GoFeatureFlagApi(
            GoFeatureFlagOptions(endpoint="http://localhost:1031", timeout=2_500)
        )._timeout
        == 2.5
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_data_collector_base_url_retargets_only_the_collector(mock_pool_manager_class):
    """It replaces the whole base — scheme, host, port and path prefix.

    Flag configuration must keep using `endpoint`.
    """
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    options = GoFeatureFlagOptions(
        endpoint="http://relay.example:1031/goff",
        data_collector_base_url="https://collector.example:8443/ingest",
    )
    api = GoFeatureFlagApi(options)

    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    api.send_event_to_data_collector([], {})
    assert (
        mock_http.request.call_args.kwargs["url"]
        == "https://collector.example:8443/ingest/v1/data/collector"
    )

    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK, json.dumps({"flags": {}})
    )
    api.retrieve_flag_configuration()
    assert (
        mock_http.request.call_args.kwargs["url"]
        == "http://relay.example:1031/goff/v1/flag/configuration"
    )


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_data_collector_falls_back_to_endpoint_when_unset(mock_pool_manager_class):
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    api = GoFeatureFlagApi(_make_options(endpoint="http://relay.example:1031/goff"))

    api.send_event_to_data_collector([], {})

    assert (
        mock_http.request.call_args.kwargs["url"]
        == "http://relay.example:1031/goff/v1/data/collector"
    )


# --- custom headers ---


@patch("gofeatureflag_python_provider.services.api.urllib3.PoolManager")
def test_custom_headers_are_applied_to_every_request(mock_pool_manager_class):
    """For deployments behind a gateway that needs its own authentication."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    options = GoFeatureFlagOptions(
        endpoint="http://localhost:1031",
        custom_headers={"X-Gateway-Token": "gw-secret"},
    )
    api = GoFeatureFlagApi(options)

    mock_http.request.return_value = _mock_response(
        HTTPStatus.OK, json.dumps({"flags": {}})
    )
    api.retrieve_flag_configuration()
    assert (
        mock_http.request.call_args.kwargs["headers"]["X-Gateway-Token"] == "gw-secret"
    )

    mock_http.request.return_value = _mock_response(HTTPStatus.OK, "")
    api.send_event_to_data_collector([], {})
    assert (
        mock_http.request.call_args.kwargs["headers"]["X-Gateway-Token"] == "gw-secret"
    )


def test_configured_api_key_wins_over_a_custom_authorization_header():
    """Least surprising when both are set: the option you set explicitly is used."""
    api = GoFeatureFlagApi(
        GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            api_key="real-key",
            custom_headers={"Authorization": "Bearer real-key"},
        )
    )

    assert api._headers()["X-API-Key"] == "real-key"
    assert api._headers()["Authorization"] == "Bearer real-key"


def test_custom_authorization_is_used_when_no_api_key_is_set():
    api = GoFeatureFlagApi(
        GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            custom_headers={"X-API-Key": "gateway-key"},
        )
    )

    assert api._headers()["X-API-Key"] == "gateway-key"
