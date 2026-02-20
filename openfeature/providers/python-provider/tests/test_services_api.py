"""
Unit tests for GO Feature Flag API service layer (GoFeatureFlagApi).
HTTP is mocked by patching urllib3.PoolManager so no real requests are made.
"""

import json
from http import HTTPStatus
from unittest.mock import Mock, patch

import pytest

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import FeatureEvent
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
    assert result.etag == "abc123"
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
def test_retrieve_flag_configuration_sends_bearer_auth_when_api_key_set(
    mock_pool_manager_class,
):
    """When api_key is set, Authorization header is sent (Bearer token or X-API-Key per implementation)."""
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
    # Python implementation uses X-API-Key; JS uses Authorization: Bearer
    assert (
        headers.get("X-API-Key") == "secret-key"
        or headers.get("Authorization") == "Bearer secret-key"
    )


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
def test_retrieve_flag_configuration_304_returns_empty_flags(mock_pool_manager_class):
    """On 304 Not Modified, response has empty flags and enrichment (per JS test)."""
    mock_http = Mock()
    mock_pool_manager_class.return_value = mock_http
    mock_http.request.return_value = _mock_response(
        HTTPStatus.NOT_MODIFIED,
        "",
        {"ETag": '"unchanged"', "Last-Modified": "Wed, 21 Oct 2015 07:28:00 GMT"},
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    result = api.retrieve_flag_configuration(etag="unchanged")

    assert result.flags == {}
    assert result.evaluation_context_enrichment == {}
    assert result.etag == "unchanged"
    assert result.last_updated is not None


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
        "{}",
        {"ETag": '"123456789"', "Last-Modified": "invalid-date"},
    )
    options = _make_options()
    api = GoFeatureFlagApi(options)

    result = api.retrieve_flag_configuration()

    # Parsing invalid date leaves last_updated as None (JS: lastUpdated?.getTime() is NaN)
    assert result.last_updated is None


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
