from __future__ import annotations

import json
import os
from pathlib import Path
from unittest.mock import Mock, patch

import pytest

from gofeatureflag_python_provider.options import EvaluationType, GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason


_default_evaluation_ctx = EvaluationContext(
    targeting_key="d45e303a-38c2-11ed-a261-0242ac120002",
    attributes={
        "email": "john.doe@gofeatureflag.org",
        "firstname": "john",
        "lastname": "doe",
        "anonymous": False,
        "professional": True,
        "rate": 3.14,
        "age": 30,
        "company_info": {"name": "my_company", "size": 120},
        "labels": ["pro", "beta"],
    },
)


def _read_config_file(name: str) -> str:
    if os.getcwd().endswith("/tests"):
        path = f"./mock_responses/config/{name}"
    else:
        path = f"./tests/mock_responses/config/{name}"
    return Path(path).read_text()


def _mock_urllib3_response(status: int, body: bytes | str, headers: dict | None = None):
    data = body.encode("utf-8") if isinstance(body, str) else body
    resp = Mock()
    resp.status = status
    resp.data = data
    resp.headers = Mock()
    resp.headers.get = Mock(side_effect=(headers or {}).get)
    return resp


def _make_config_dispatcher(config_file: str = "valid-all-types.json"):
    """Return a side_effect callable that routes urllib3 requests like the Java GoffApiMock."""
    config_body = _read_config_file(config_file)
    etag = f'"{config_file}"'

    def dispatcher(**kwargs):
        url = kwargs.get("url", "")
        if "v1/flag/configuration" in url:
            return _mock_urllib3_response(
                200,
                config_body,
                {"ETag": etag, "Last-Modified": "Wed, 21 Oct 2015 07:28:00 GMT"},
            )
        if "v1/data/collector" in url:
            return _mock_urllib3_response(200, '{"ingestedContentCount":0}')
        raise ValueError(f"Unexpected URL in mock: {url}")

    return dispatcher


def _make_error_dispatcher(status_code: int):
    """Return a dispatcher that always returns the given error status."""

    def dispatcher(**kwargs):
        return _mock_urllib3_response(status_code, "")

    return dispatcher


def _setup_provider_and_client(
    mock_urllib3_request, config_file="valid-all-types.json"
):
    mock_urllib3_request.side_effect = _make_config_dispatcher(config_file)
    provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            evaluation_type=EvaluationType.INPROCESS,
            disable_cache_invalidation=True,
            data_flush_interval=100000,
        )
    )
    api.set_provider(provider)
    client = api.get_client(domain="inprocess-test")
    return client


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_flag_not_found_if_flag_does_not_exist(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_boolean_details(
            flag_key="DOES_NOT_EXISTS",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "DOES_NOT_EXISTS"
        assert got.value is False
        assert got.error_code == ErrorCode.FLAG_NOT_FOUND
        assert got.reason == Reason.ERROR
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_error_if_we_expect_boolean_and_got_another_type(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_boolean_details(
            flag_key="string_key",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "string_key"
        assert got.value is False
        assert got.error_code == ErrorCode.TYPE_MISMATCH
        assert got.reason == Reason.ERROR
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_valid_boolean_flag_with_targeting_match(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "bool_targeting_match"
        assert got.value is True
        assert got.variant == "enabled"
        assert got.reason == Reason.TARGETING_MATCH
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_valid_string_flag(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_string_details(
            flag_key="string_key",
            default_value="",
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "string_key"
        assert got.value == "CC0002"
        assert got.variant == "color1"
        assert got.reason == Reason.STATIC
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_valid_double_flag_with_targeting_match(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_float_details(
            flag_key="double_key",
            default_value=100.10,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "double_key"
        assert got.value == pytest.approx(101.25)
        assert got.variant == "medium"
        assert got.reason == Reason.TARGETING_MATCH
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_valid_integer_flag_with_targeting_match(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_integer_details(
            flag_key="integer_key",
            default_value=1000,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "integer_key"
        assert got.value == 101
        assert got.variant == "medium"
        assert got.reason == Reason.TARGETING_MATCH
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_valid_object_flag_with_targeting_match(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_object_details(
            flag_key="object_key",
            default_value={"default": "true"},
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "object_key"
        assert got.value == {"test": "false"}
        assert got.variant == "varB"
        assert got.reason == Reason.TARGETING_MATCH
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_boolean_default_value_if_flag_is_disabled(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(mock_urllib3_request)
        got = client.get_boolean_details(
            flag_key="disabled_bool",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "disabled_bool"
        assert got.value is False
        assert got.variant == "SdkDefault"
        assert got.reason == Reason.DISABLED
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_error_if_flag_configuration_endpoint_returns_404(mock_urllib3_request):
    try:
        mock_urllib3_request.side_effect = _make_error_dispatcher(404)
        provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="http://localhost:1031",
                evaluation_type=EvaluationType.INPROCESS,
                disable_cache_invalidation=True,
            )
        )
        api.set_provider(provider)
        client = api.get_client(domain="inprocess-test-404")
        got = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.value is False
        assert got.reason == Reason.ERROR
        assert got.error_code is not None
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_error_if_endpoint_not_available(mock_urllib3_request):
    try:
        mock_urllib3_request.side_effect = _make_error_dispatcher(500)
        provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="http://localhost:1031",
                evaluation_type=EvaluationType.INPROCESS,
                disable_cache_invalidation=True,
            )
        )
        api.set_provider(provider)
        client = api.get_client(domain="inprocess-test-500")
        got = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.value is False
        assert got.reason == Reason.ERROR
        assert got.error_code is not None
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_apply_scheduled_rollout_step(mock_urllib3_request):
    try:
        client = _setup_provider_and_client(
            mock_urllib3_request, config_file="valid-scheduled-rollout.json"
        )
        got = client.get_boolean_details(
            flag_key="my-flag",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "my-flag"
        assert got.value is True
        assert got.variant == "enabled"
        assert got.reason == Reason.TARGETING_MATCH
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_not_apply_scheduled_rollout_step_if_date_in_future(
    mock_urllib3_request,
):
    try:
        client = _setup_provider_and_client(
            mock_urllib3_request, config_file="valid-scheduled-rollout.json"
        )
        got = client.get_boolean_details(
            flag_key="my-flag-scheduled-in-future",
            default_value=True,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "my-flag-scheduled-in-future"
        assert got.value is False
        assert got.variant == "disabled"
        assert got.reason == Reason.STATIC
        assert got.flag_metadata is not None
        assert got.flag_metadata.get("description") == "this is a test flag"
    finally:
        api.shutdown()
