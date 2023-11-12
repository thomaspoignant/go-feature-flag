import os
import time
from pathlib import Path
from unittest.mock import Mock, patch

import pydantic
import pytest
from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import Reason, FlagEvaluationDetails

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.provider_status import ProviderStatus

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
        mock_request, flag_key, default_value, ctx: EvaluationContext, evaluationType: str
):
    try:
        mock_request.return_value = Mock(status="200", data=_read_mock_file(flag_key))
        goff_provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://gofeatureflag.org/",
                data_flush_interval=100,
            ),
        )
        api.set_provider(goff_provider)
        wait_provider_ready(goff_provider)
        client = api.get_client(name="test-client")

        if evaluationType == "bool":
            t = client.get_boolean_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
            return t
        elif evaluationType == "string":
            return client.get_string_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluationType == "float":
            return client.get_float_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluationType == "int":
            return client.get_integer_details(
                flag_key=flag_key,
                default_value=default_value,
                evaluation_context=ctx,
            )
        elif evaluationType == "object":
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
            data_flush_interval=100
        )
    )
    assert len(goff_provider.get_provider_hooks()) == 1


def test_constructor_options_none():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(options=None)


def test_constructor_options_empty():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider()


def test_constructor_options_empty_endpoint():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(options=GoFeatureFlagOptions(
            endpoint="", data_flush_interval=100
        ))


def test_constructor_options_invalid_url():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(options=GoFeatureFlagOptions(
            endpoint="not a url",
            data_flush_interval=100
        ))


def test_constructor_options_valid():
    try:
        GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://app.gofeatureflag.org/",
                data_flush_interval=100
            )
        )
    except Exception as exc:
        assert False, f"'constructor has raised an exception {exc}"


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_endpoint_not_available(mock_request):
    try:
        flag_key = "fail_500"
        mock_request.return_value = Mock(status="500")
        goff_provider = GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="https://invalidurl.com",
                data_flush_interval=100
            )
        )
        api.set_provider(goff_provider)
        client = api.get_client(name="test-client")
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


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_flag_does_not_exists(mock_request):
    flag_key = "flag_not_found"
    default_value = False
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert ErrorCode.FLAG_NOT_FOUND == res.error_code
    assert Reason.ERROR == res.reason
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_we_expect_a_boolean_and_got_another_type(
        mock_request,
):
    flag_key = "string_key"
    default_value = False
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_boolean_flag_with_targeting_match_reason(mock_request):
    flag_key = "bool_targeting_match"
    default_value = False
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert res == want


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_custom_reason_if_returned_by_relay_proxy(mock_request):
    flag_key = "unknown_reason"
    default_value = False
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is True
    assert res.error_code is None
    assert "CUSTOM_REASON" == res.reason
    assert "True" == res.variant


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_boolean_default_value_if_the_flag_is_disabled(mock_request):
    flag_key = "disabled_bool"
    default_value = False
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "bool"
    )
    assert flag_key == res.flag_key
    assert res.value is False
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_we_expect_a_string_and_got_another_type(
        mock_request,
):
    flag_key = "object_key"
    default_value = "default"
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert default_value == res.value
    assert ErrorCode.TYPE_MISMATCH == res.error_code
    assert Reason.ERROR == res.reason
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_string_flag_with_targeting_match_reason(mock_request):
    flag_key = "string_key"
    default_value = "default"
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert res.value == "CC0000"
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_string_default_value_if_the_flag_is_disabled(mock_request):
    flag_key = "disabled_string"
    default_value = "default"
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "string"
    )
    assert flag_key == res.flag_key
    assert res.value == "default"
    assert res.error_code is None
    assert res.reason == Reason.DISABLED
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_we_expect_a_integer_and_got_another_type(
        mock_request,
):
    flag_key = "string_key"
    default_value = 200
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 200
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_integer_flag_with_targeting_match_reason(mock_request):
    flag_key = "integer_key"
    default_value = 1200
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 100
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_integer_default_value_if_the_flag_is_disabled(mock_request):
    flag_key = "disabled_int"
    default_value = 1225
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 1225
    assert res.error_code is None
    assert res.reason == Reason.DISABLED
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_double_flag_with_targeting_match_reason(mock_request):
    flag_key = "double_key"
    default_value = 1200.25
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "float"
    )
    assert flag_key == res.flag_key
    assert res.value == pytest.approx(100.25, rel=None, abs=None, nan_ok=False)
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH.value
    assert res.variant == "True"


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_we_expect_a_integer_and_double_type(mock_request):
    flag_key = "double_key"
    default_value = 200
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "int"
    )
    assert flag_key == res.flag_key
    assert res.value == 200
    assert res.error_code == ErrorCode.TYPE_MISMATCH
    assert res.reason == Reason.ERROR
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_double_default_value_if_the_flag_is_disabled(mock_request):
    flag_key = "disabled_float"
    default_value = 1200.25
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "float"
    )
    assert flag_key == res.flag_key
    assert default_value == pytest.approx(res.value, rel=None, abs=None, nan_ok=False)
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_value_flag_with_targeting_match_reason(mock_request):
    flag_key = "object_key"
    default_value = None
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "object"
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


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_use_object_default_value_if_the_flag_is_disabled(mock_request):
    flag_key = "disabled_object"
    default_value = {"default": True}
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "object"
    )
    assert flag_key == res.flag_key
    assert {"default": True} == res.value
    assert res.error_code is None
    assert Reason.DISABLED == res.reason
    assert res.variant is None


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_return_an_error_if_no_targeting_key(mock_request):
    flag_key = "string_key"
    default_value = "empty"
    res = _generic_test(
        mock_request, flag_key, default_value, EvaluationContext(), "string"
    )

    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=default_value,
        reason=Reason.ERROR,
        error_code=ErrorCode.TARGETING_KEY_MISSING,
        error_message="targetingKey field MUST be set in your EvaluationContext",
    )
    assert res == want


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_a_valid_value_flag_with_a_list(mock_request):
    flag_key = "list_key"
    default_value = {}
    res = _generic_test(
        mock_request, flag_key, default_value, _default_evaluation_ctx, "object"
    )
    assert flag_key == res.flag_key
    assert res.value == ["test", "test1", "test2", "false", "test3"]
    assert res.error_code is None
    assert res.reason == Reason.TARGETING_MATCH
    assert "True" == res.variant


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_resolve_from_cache_if_multiple_call_to_the_same_flag_with_same_context(mock_request: Mock):
    flag_key = "bool_targeting_match"
    default_value = False

    mock_request.return_value = Mock(status="200", data=_read_mock_file(flag_key))
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=100,
            disable_data_collection=True,
        )
    )
    api.set_provider(goff_provider)
    wait_provider_ready(goff_provider)
    client = api.get_client(name="test-client")

    got = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )

    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert got == want

    got = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.CACHED,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert got == want
    api.shutdown()


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_call_data_collector_multiple_times_with_cached_event_waiting_ttl(mock_request: Mock):
    flag_key = "bool_targeting_match"
    default_value = False
    mock_request.side_effect = [
        Mock(status="200", data=_read_mock_file(flag_key)),  # first call to get the flag
        Mock(status="200", data={}),  # second call to send the data
        Mock(status="200", data={})
    ]
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=100,
        )
    )
    api.set_provider(goff_provider)
    wait_provider_ready(goff_provider)
    client = api.get_client(name="test-client")

    got = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )

    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert got == want
    got = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    want: FlagEvaluationDetails = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.CACHED,
        flag_metadata={"test": "test1", "test2": False, "test3": 123.3},
    )
    assert got == want
    time.sleep(0.2)
    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    api.shutdown()
    assert mock_request.call_count == 3


@patch("urllib3.poolmanager.PoolManager.request")
def test_should_not_call_data_collector_if_not_having_cache(mock_request: Mock):
    flag_key = "bool_targeting_match"
    default_value = False
    mock_request.side_effect = [
        Mock(status="200", data=_read_mock_file(flag_key)),
    ]
    goff_provider = None
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=1000,
        )
    )

    api.set_provider(goff_provider)
    wait_provider_ready(goff_provider)
    client = api.get_client(name="test-client")

    client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    api.shutdown()
    assert mock_request.call_count == 1


def wait_provider_ready(provider: GoFeatureFlagProvider):
    # check the provider get_status method until it returns ProviderStatus.READY or, we waited more than 5 seconds
    start = time.time()
    while provider.get_status() != ProviderStatus.READY:
        time.sleep(0.1)
        if time.time() - start > 5:
            break


def _read_mock_file(flag_key: str) -> str:
    # This hacky if is here to make test run inside pycharm and from the root of the project
    if os.getcwd().endswith("/tests"):
        path_prefix = "./mock_responses/{}.json"
    else:
        path_prefix = "./tests/mock_responses/{}.json"
    return Path(path_prefix.format(flag_key)).read_text()
