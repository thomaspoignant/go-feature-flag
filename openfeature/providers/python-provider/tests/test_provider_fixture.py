from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason
import pydantic
import pytest
from .fixtures import goff

from gofeatureflag_python_provider.exception.invalid_options_exception import (
    InvalidOptionsException,
)
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions


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


def test_in_process_evaluation(goff):
    """
    test in process evaluation with default context
    """
    flag_key = "bool_targeting_match"
    default_value = False

    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint=goff,
            data_flush_interval=100,
            disable_data_collection=True,
            api_key="apikey1",
        )
    )
    api.set_provider(goff_provider)
    client = api.get_client(domain="test-client")

    want = FlagEvaluationDetails(
        flag_key=flag_key,
        value=True,
        variant="True",
        reason=Reason.TARGETING_MATCH,
        flag_metadata={
            "description": "this is a test",
            "pr_link": "https://github.com/thomaspoignant/go-feature-flag/pull/916",
        },
    )
    got = client.get_boolean_details(
        flag_key=flag_key,
        default_value=default_value,
        evaluation_context=_default_evaluation_ctx,
    )
    assert got == want


def test_provider_metadata():
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031", data_flush_interval=100
        )
    )
    assert goff_provider.get_metadata().name == "GO Feature Flag"


def test_constructor_options_none():
    with pytest.raises(InvalidOptionsException):
        GoFeatureFlagProvider(options=None)


def test_constructor_invalid_endpoint_option():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(options=GoFeatureFlagOptions(endpoint="123"))


def test_constructor_invalid_exporter_metadata_option():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider(
            options=GoFeatureFlagOptions(
                endpoint="http://gofeatureflag.org",
                exporter_metadata={"invalid": "metadata"},
            )
        )
