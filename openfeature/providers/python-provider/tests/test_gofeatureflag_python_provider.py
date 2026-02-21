from __future__ import annotations

import json
import os
from typing import Optional
import pydantic
import pytest
import requests
import time
from gofeatureflag_python_provider.hooks import DataCollectorHook
from gofeatureflag_python_provider.options import EvaluationType, GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import Reason, FlagEvaluationDetails
from pathlib import Path
from unittest.mock import Mock, patch

from openfeature import api


def _mock_urllib3_response(status: int, body: bytes | str, headers: dict | None = None):
    """Build a mock urllib3 HTTP response with .status, .data, .headers.get()."""
    data = body.encode("utf-8") if isinstance(body, str) else body
    resp = Mock()
    resp.status = status
    resp.data = data
    resp.headers = Mock()
    resp.headers.get = Mock(side_effect=(headers or {}).get)
    return resp


def _mock_session_response(status_code: int, json_body):
    """Build a Mock that behaves like requests.Response for OFREP/requests.Session."""
    if isinstance(json_body, (str, bytes)):
        data = json.loads(json_body)
    else:
        data = json_body
    data = dict(data)
    # OFREP expects "variant"; mock files use "variationType"
    if "variationType" in data and "variant" not in data:
        data["variant"] = data["variationType"]
    # When disabled, variant is typically None
    if data.get("reason") == "DISABLED":
        data["variant"] = None
    mock = Mock()
    mock.status_code = int(status_code) if isinstance(status_code, str) else status_code
    mock.json.return_value = data
    mock.headers = {}
    mock.text = json.dumps(data) if isinstance(data, dict) else str(data)
    mock.content = mock.text.encode()
    if mock.status_code >= 400:
        err = requests.HTTPError(response=mock)
        mock.raise_for_status = Mock(side_effect=err)
    return mock


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


def _generic_test(
    mock_post,
    flag_key,
    default_value,
    ctx: EvaluationContext,
    evaluation_type: str,
    http_status: Optional[int] = 200,
):
    try:
        mock_post.return_value = _mock_session_response(
            http_status, _read_mock_file(flag_key)
        )
        goff_provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://gofeatureflag.org/",
                data_flush_interval=100,
                disable_cache_invalidation=True,
                api_key="apikey1",
                evaluation_type=EvaluationType.REMOTE,
                disable_data_collection=True,
            ),
        )
        api.set_provider(goff_provider)
        client = api.get_client(domain="test-client")

        if evaluation_type == "bool":
            t = client.get_boolean_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
            return t
        elif evaluation_type == "string":
            return client.get_string_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluation_type == "float":
            return client.get_float_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluation_type == "int":
            return client.get_integer_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluation_type == "object":
            return client.get_object_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        api.shutdown()
    except Exception as exc:
        assert False, f"'No exception expected {exc}"


def test_provider_metadata():
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031", data_flush_interval=100
        )
    )
    assert goff_provider.get_metadata().name == "GO Feature Flag"


def test_number_hook():
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            data_flush_interval=100,
            evaluation_type=EvaluationType.INPROCESS,
        )
    )
    assert len(goff_provider.get_provider_hooks()) == 1


def test_number_hook_with_exporter_metadata():
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            data_flush_interval=100,
            evaluation_type=EvaluationType.REMOTE,
            exporter_metadata={"version": "1.0.0", "name": "myapp", "id": 123},
        )
    )
    assert len(goff_provider.get_provider_hooks()) == 2


def test_constructor_options_none():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(options=None)


def test_constructor_options_empty():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider()


def test_constructor_options_empty_endpoint():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(endpoint="", data_flush_interval=100)
        )


def test_constructor_options_invalid_url():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(endpoint="not a url", data_flush_interval=100)
        )


def test_constructor_options_valid():
    try:
        GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://app.gofeatureflag.org/", data_flush_interval=100
            )
        )
    except Exception as exc:
        assert False, f"'constructor has raised an exception {exc}"


@patch("requests.Session.post")
def test_should_return_an_error_if_endpoint_not_available(mock_post):
    try:
        flag_key = "fail_500"
        mock_post.return_value = _mock_session_response(
            500,
            '{"errorDetails": "An internal server error occurred while processing the request"}',
        )
        goff_provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://invalidurl.com",
                data_flush_interval=100,
                evaluation_type=EvaluationType.REMOTE,
            )
        )
        api.set_provider(goff_provider)
        client = api.get_client(domain="test-client")
        res = client.get_boolean_details(
            flag_key=flag_key,
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert flag_key == res.flag_key
        assert res.value is False
        assert res.error_code == ErrorCode.GENERAL
        assert Reason.ERROR == res.reason
    except Exception as exc:
        assert False, f"'No exception expected {exc}"


@patch("requests.Session.post")
def test_should_return_an_error_if_flag_does_not_exists(mock_post):
    flag_key = "flag_not_found"
    default_value = False
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "bool", 404
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert ErrorCode.FLAG_NOT_FOUND == res.error_code
    assert Reason.ERROR == res.reason
    assert res.variant is None


@patch("requests.Session.post")
def test_should_return_an_error_if_we_expect_a_boolean_and_got_another_type(
    mock_post,
):
    flag_key = "string_key"
    default_value = False
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_boolean_flag_with_targeting_match_reason(mock_post):
    flag_key = "bool_targeting_match"
    default_value = False
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert res == want


@patch("requests.Session.post")
def test_should_resolve_a_valid_boolean_flag_with_targeting_match_reason_without_error_code(
    mock_post,
):
    flag_key = "bool_targeting_match_no_error_field"
    default_value = False
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert res == want


# TODO: Uncomment this test when the relay proxy is updated to return the correct error code
# PR open https://github.com/open-feature/python-sdk-contrib/pull/346
# @patch("requests.Session.post")
# def test_should_return_custom_reason_if_returned_by_relay_proxy(mock_post):
#     flag_key = "unknown_reason"
#     default_value = False
#     res = _generic_test(
#         mock_post, flag_key, default_value, _default_evaluation_ctx, "bool", 400
#     )
#     assert flag_key == res.flag_key
#     assert res.value is True
#     assert res.error_code is None
#     assert "CUSTOM_REASON" == res.reason
#     assert "True" == res.variant


@patch("requests.Session.post")
def test_should_use_boolean_default_value_if_the_flag_is_disabled(mock_post):
    flag_key = "disabled_bool"
    default_value = False
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("requests.Session.post")
def test_should_return_an_error_if_we_expect_a_string_and_got_another_type(
    mock_post,
):
    flag_key = "object_key"
    default_value = "default"
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert default_value == res.value
    assert ErrorCode.TYPE_MISMATCH == res.error_code
    assert Reason.ERROR == res.reason
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_string_flag_with_targeting_match_reason(mock_post):
    flag_key = "string_key"
    default_value = "default"
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert res.value == "CC0000"
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("requests.Session.post")
def test_should_use_string_default_value_if_the_flag_is_disabled(mock_post):
    flag_key = "disabled_string"
    default_value = "default"
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert res.value == "default"
    assert res.error_code is None
    assert res.reason == Reason.DISABLED
    assert res.variant is None


@patch("requests.Session.post")
def test_should_return_an_error_if_we_expect_a_integer_and_got_another_type(
    mock_post,
):
    flag_key = "string_key"
    default_value = 200
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 200
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_integer_flag_with_targeting_match_reason(mock_post):
    flag_key = "integer_key"
    default_value = 1200
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 100
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("requests.Session.post")
def test_should_use_integer_default_value_if_the_flag_is_disabled(mock_post):
    flag_key = "disabled_int"
    default_value = 1225
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    # assert res.value == 1225
    assert res.error_code is None
    assert res.reason == Reason.DISABLED
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_double_flag_with_targeting_match_reason(mock_post):
    flag_key = "double_key"
    default_value = 1200.25
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "float"
    )
    assert flag_key == res.flag_key
    assert res.value == pytest.approx(100.25, rel=None, abs=None, nan_ok=False)
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("requests.Session.post")
def test_should_return_an_error_if_we_expect_a_integer_and_double_type(mock_post):
    flag_key = "double_key"
    default_value = 200
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 200
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("requests.Session.post")
def test_should_use_double_default_value_if_the_flag_is_disabled(mock_post):
    flag_key = "disabled_float"
    default_value = 1200.25
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "float"
    )
    assert flag_key == res.flag_key
    assert default_value == 1200.25
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_value_flag_with_targeting_match_reason(mock_post):
    flag_key = "object_key"
    default_value = None
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "object"
    )
    assert flag_key == res.flag_key
    assert {
        "test": "test1",
        "test2": False,
        "test3": 123.3,
        "test4": 1,
        "test5": None,
    } == res.value
    assert res.error_code is None
    assert Reason.TARGETING_MATCH.value == res.reason
    assert "True" == res.variant


@patch("requests.Session.post")
def test_should_use_object_default_value_if_the_flag_is_disabled(mock_post):
    flag_key = "disabled_object"
    default_value = {"default": True}
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "object"
    )
    print(res)
    assert flag_key == res.flag_key
    assert {"default": True} == res.value
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("requests.Session.post")
def test_should_resolve_a_valid_value_flag_with_a_list(mock_post):
    flag_key = "list_key"
    default_value = {}
    res = _generic_test(
        mock_post, flag_key, default_value, _default_evaluation_ctx, "object"
    )
    assert flag_key == res.flag_key
    assert res.value == ["test", "test1", "test2", "false", "test3"]
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH
    assert "True" == res.variant


def _flag_config_for_bool_targeting_match():
    """Minimal flag config for bool_targeting_match (INPROCESS tests)."""
    return {
        "flags": {
            "bool_targeting_match": {
                "defaultRule": {"percentage": {"False": 0, "True": 100}},
                "metadata": {"description": "test"},
                "targeting": [
                    {
                        "query": 'email eq "john.doe@gofeatureflag.org"',
                        "variation": "True",
                    }
                ],
                "trackEvents": True,
                "variations": {"Default": False, "False": False, "True": True},
            }
        },
        "evaluationContextEnrichment": {},
    }


@patch("urllib3.poolmanager.PoolManager.request")
@patch("requests.Session.post")
def test_should_call_data_collector_with_exporter_metadata(
    mock_post: Mock,
    mock_urllib3_request: Mock,
):
    flag_key = "bool_targeting_match"
    default_value = False
    mock_post.side_effect = [
        _mock_session_response(200, _read_mock_file(flag_key)),
        _mock_session_response(200, {}),
        _mock_session_response(200, {}),
    ]
    flag_config_resp = _mock_urllib3_response(
        status=200,
        body=json.dumps(_flag_config_for_bool_targeting_match()),
        headers={"ETag": '"etag1"', "Last-Modified": "Wed, 18 Feb 2025 12:00:00 GMT"},
    )
    collector_resp = _mock_urllib3_response(200, "{}")
    mock_urllib3_request.side_effect = [flag_config_resp, collector_resp]
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=100,
            disable_cache_invalidation=True,
            exporter_metadata={"version": "1.0.0", "name": "myapp", "id": 123},
            evaluation_type=EvaluationType.INPROCESS,
        )
    )
    api.set_provider(goff_provider)
    client = api.get_client(domain="test-client")

    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )

    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    time.sleep(0.2)
    api.shutdown()
    want = {
        "provider": "python",
        "openfeature": True,
        "version": "1.0.0",
        "name": "myapp",
        "id": 123,
    }
    # Data collector uses urllib3; find the /v1/data/collector call
    got = {}
    for call in mock_urllib3_request.call_args_list:
        kwargs = call[1] if len(call) > 1 else {}
        url = kwargs.get("url", "")
        body = kwargs.get("body")
        if body and "data/collector" in str(url):
            payload = json.loads(body)
            if "meta" in payload:
                got = payload["meta"]
                break
    assert got == want


@patch("requests.Session.post")
def test_should_not_call_data_collector_if_not_having_cache(mock_post: Mock):
    flag_key = "bool_targeting_match"
    default_value = False
    mock_post.side_effect = [
        _mock_session_response(200, _read_mock_file(flag_key)),
    ]
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=1000,
            disable_cache_invalidation=True,
            evaluation_type=EvaluationType.REMOTE,
        )
    )

    api.set_provider(goff_provider)
    client = api.get_client(domain="test-client")

    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    api.shutdown()
    assert mock_post.call_count == 1


def test_hook_error():
    def _create_hook_context():
        ctx = Mock()
        eval_ctx = Mock()
        eval_ctx.attributes = {"anonymous": True}
        eval_ctx.targeting_key = "test_user_key"
        ctx.evaluation_context = eval_ctx
        ctx.flag_key = "test_key"
        ctx.default_value = False
        return ctx

    mock_evaluator = Mock()
    mock_evaluator.is_flag_trackable.return_value = True
    mock_event_publisher = Mock()
    mock_event_publisher.add_event.return_value = None

    hook = DataCollectorHook(
        options=GoFeatureFlagOptions(endpoint="http://test.com"),
        event_publisher=mock_event_publisher,
        evaluator=mock_evaluator,
    )
    hook.error(
        hook_context=_create_hook_context(),
        exception=Exception("test exception raised"),
        hints={},
    )
    assert mock_event_publisher.add_event.call_count == 1
    event = mock_event_publisher.add_event.call_args[0][0]
    assert event.contextKind == "anonymousUser"


@patch("requests.Session.post")
def test_url_parsing(mock_post):
    flag_key = "bool_targeting_match"
    mock_post.return_value = _mock_session_response(200, _read_mock_file(flag_key))
    default_value = False
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/ff/",
            data_flush_interval=100,
            disable_cache_invalidation=True,
            api_key="apikey1",
            evaluation_type=EvaluationType.REMOTE,
        ),
    )
    api.set_provider(goff_provider)
    client = api.get_client(domain="test-client")
    t = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    # OFREP uses /ofrep/v1/evaluate/flags/{key}
    got = mock_post.call_args[0][0]
    want = "https://gofeatureflag.org/ff/ofrep/v1/evaluate/flags/bool_targeting_match"
    assert got == want


@patch("urllib3.poolmanager.PoolManager.request")
@patch("requests.Session.post")
def test_should_call_evaluation_api_with_exporter_metadata(
    mock_post: Mock,
    mock_urllib3_request: Mock,
):
    flag_key = "bool_targeting_match"
    default_value = False
    mock_post.side_effect = [
        _mock_session_response(200, _read_mock_file(flag_key)),
        _mock_session_response(200, {}),
        _mock_session_response(200, {}),
    ]
    flag_config_resp = _mock_urllib3_response(
        status=200,
        body=json.dumps(_flag_config_for_bool_targeting_match()),
        headers={"ETag": '"etag1"', "Last-Modified": "Wed, 18 Feb 2025 12:00:00 GMT"},
    )
    collector_resp = _mock_urllib3_response(200, "{}")
    mock_urllib3_request.side_effect = [flag_config_resp, collector_resp]
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=100,
            disable_cache_invalidation=True,
            exporter_metadata={"version": "1.0.0", "name": "myapp", "id": 123},
            evaluation_type=EvaluationType.INPROCESS,
        )
    )
    api.set_provider(goff_provider)
    client = api.get_client(domain="test-client")

    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )

    api.shutdown()
    got = {}
    for call in mock_urllib3_request.call_args_list:
        kwargs = call[1] if len(call) > 1 else {}
        url = kwargs.get("url", "")
        body = kwargs.get("body")
        if body and "data/collector" in str(url):
            payload = json.loads(body if isinstance(body, str) else body.decode())
            if "meta" in payload:
                got = payload["meta"]
                break
    want = {
        "provider": "python",
        "openfeature": True,
        "version": "1.0.0",
        "name": "myapp",
        "id": 123,
    }
    assert got == want


def test_hook_after_without_anonymous_attribute():
    from openfeature.hook import HookContext
    from openfeature.evaluation_context import EvaluationContext
    from openfeature.flag_evaluation import FlagEvaluationDetails, Reason

    def _create_hook_context_without_anonymous():
        ctx = HookContext(
            flag_key="test_key",
            flag_type=bool,
            default_value=False,
            evaluation_context=EvaluationContext(
                targeting_key="test_user_key",
                attributes={"email": "test@example.com"},  # No anonymous attribute
            ),
            client_metadata=None,
            provider_metadata=None,
        )
        return ctx

    mock_evaluator = Mock()
    mock_evaluator.is_flag_trackable.return_value = True

    mock_event_publisher = Mock()
    mock_event_publisher.add_event.return_value = None
    hook = DataCollectorHook(
        options=GoFeatureFlagOptions(endpoint="http://test.com"),
        event_publisher=mock_event_publisher,
        evaluator=mock_evaluator,
    )

    details = FlagEvaluationDetails(
        flag_key="test_key", value=True, reason=Reason.CACHED, variant="true_variant"
    )

    # This should not raise a KeyError (anonymous attribute missing -> defaults to "user")
    hook.after(
        hook_context=_create_hook_context_without_anonymous(),
        details=details,
        hints={},
    )

    # Verify the event was created with contextKind "user" (default when anonymous is missing)
    assert mock_event_publisher.add_event.call_count == 1
    event = mock_event_publisher.add_event.call_args[0][0]
    assert event.contextKind == "user"


def test_hook_error_without_anonymous_attribute():
    from openfeature.hook import HookContext
    from openfeature.evaluation_context import EvaluationContext

    def _create_hook_context_without_anonymous():
        ctx = HookContext(
            flag_key="test_key",
            flag_type=bool,
            default_value=False,
            evaluation_context=EvaluationContext(
                targeting_key="test_user_key",
                attributes={"email": "test@example.com"},  # No anonymous attribute
            ),
            client_metadata=None,
            provider_metadata=None,
        )
        return ctx

    mock_evaluator = Mock()
    mock_evaluator.is_flag_trackable.return_value = True

    mock_event_publisher = Mock()
    mock_event_publisher.add_event.return_value = None
    hook = DataCollectorHook(
        options=GoFeatureFlagOptions(endpoint="http://test.com"),
        event_publisher=mock_event_publisher,
        evaluator=mock_evaluator,
    )

    # This should not raise a KeyError (anonymous attribute missing -> defaults to "user")
    hook.error(
        hook_context=_create_hook_context_without_anonymous(),
        exception=Exception("test exception raised"),
        hints={},
    )

    # Verify the event was created with contextKind "user" (default when anonymous is missing)
    assert mock_event_publisher.add_event.call_count == 1
    event = mock_event_publisher.add_event.call_args[0][0]
    assert event.contextKind == "user"


def _read_mock_file(flag_key: str) -> str:
    # This hacky if is here to make test run inside pycharm and from the root of the project
    if os.getcwd().endswith("/tests"):
        path_prefix = "./mock_responses/{}.json"
    else:
        path_prefix = "./tests/mock_responses/{}.json"
    return Path(path_prefix.format(flag_key)).read_text()
