"""Real evaluations against a relay proxy running in a container.

Every other test in this suite mocks the transport: the in-process tests patch
urllib3 and replay a recorded /v1/flag/configuration body, the remote tests build
fake requests.Response objects. Nothing there checks that those recordings still
match what a relay proxy actually serves, nor that the bundled WASM engine reaches
the same verdict the proxy does — the mocks are only ever checked against
themselves.

This module closes that gap. It boots a real relay proxy (see the `relay_proxy`
fixture in conftest.py, whose flag file is generated from the very fixture the
mocked tests replay) and evaluates through the public OpenFeature API in both
evaluation modes. Expectations are deliberately identical to the ones in
test_gofeatureflag_provider_inprocess.py, so a mock that drifts from reality
fails here.

Skipped when no Docker daemon is reachable.
"""

from __future__ import annotations

from contextlib import contextmanager
from typing import Generator

import pytest

from gofeatureflag_python_provider.options import EvaluationType, GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from openfeature import api
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.flag_evaluation import Reason
from tests.conftest import docker_is_available

pytestmark = pytest.mark.skipif(
    not docker_is_available(), reason="Docker is not available"
)

_EVALUATION_TYPES = [EvaluationType.INPROCESS, EvaluationType.REMOTE]

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


@contextmanager
def _client(endpoint: str, evaluation_type: EvaluationType) -> Generator:
    """An OpenFeature client backed by the containerised relay proxy.

    The OpenFeature api is a process-wide singleton, so the provider has to be
    torn down even when the test body fails, and each client needs its own domain
    to avoid picking up one left behind by an earlier test.
    """
    provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint=endpoint,
            evaluation_type=evaluation_type,
            # Not what these tests exercise, and its background flush thread would
            # only add noise and flake.
            disable_data_collection=True,
        )
    )
    api.set_provider(provider)
    try:
        yield api.get_client(domain=f"integration-{evaluation_type.value}")
    finally:
        api.shutdown()


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_resolve_valid_boolean_flag_with_targeting_match(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_boolean_details(
            flag_key="bool_targeting_match",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "bool_targeting_match"
        assert got.value is True
        assert got.variant == "enabled"
        assert got.reason == Reason.TARGETING_MATCH
        assert got.flag_metadata.get("description") == "this is a test flag"


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_resolve_valid_string_flag(relay_proxy, evaluation_type):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_string_details(
            flag_key="string_key",
            default_value="",
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "string_key"
        assert got.value == "CC0002"
        assert got.variant == "color1"
        assert got.reason == Reason.STATIC
        assert got.flag_metadata.get("description") == "this is a test flag"


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_resolve_valid_integer_flag_with_targeting_match(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_integer_details(
            flag_key="integer_key",
            default_value=1000,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "integer_key"
        assert got.value == 101
        assert got.variant == "medium"
        assert got.reason == Reason.TARGETING_MATCH


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_resolve_valid_double_flag_with_targeting_match(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_float_details(
            flag_key="double_key",
            default_value=100.10,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "double_key"
        assert got.value == pytest.approx(101.25)
        assert got.variant == "medium"
        assert got.reason == Reason.TARGETING_MATCH


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_resolve_valid_object_flag_with_targeting_match(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_object_details(
            flag_key="object_key",
            default_value={"default": "true"},
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "object_key"
        assert got.value == {"test": "false"}
        assert got.variant == "varB"
        assert got.reason == Reason.TARGETING_MATCH


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_use_default_value_if_flag_is_disabled(relay_proxy, evaluation_type):
    """A disabled flag yields the caller's default — but not by the same route.

    The two modes genuinely disagree on how they get there, so this asserts what
    each really does rather than something vague enough to cover both.

    In-process reports it cleanly: reason DISABLED, variant SdkDefault, no error.
    Remote cannot, because the relay proxy answers OFREP with a null value
    ({"value": null, "reason": "DISABLED", "variant": "SdkDefault"}) and the OFREP
    client type-checks that null against the requested bool before the DISABLED
    reason is ever considered. The caller still gets its default, but by way of a
    TYPE_MISMATCH error rather than a disabled verdict.

    If that is ever reconciled, this test should fail and be updated — which is
    the point of pinning it. See tests/mock_responses/ofrep/disabled_bool.json,
    which records a `false` value where the proxy sends null and so hides this.
    """
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_boolean_details(
            flag_key="disabled_bool",
            default_value=True,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "disabled_bool"
        # The caller's default, not the flag's own variations, either way.
        assert got.value is True

        if evaluation_type == EvaluationType.INPROCESS:
            assert got.variant == "SdkDefault"
            assert got.reason == Reason.DISABLED
            assert got.error_code is None
        else:
            assert got.variant is None
            assert got.reason == Reason.ERROR
            assert got.error_code == ErrorCode.TYPE_MISMATCH


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_return_flag_not_found_if_flag_does_not_exist(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_boolean_details(
            flag_key="DOES_NOT_EXISTS",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "DOES_NOT_EXISTS"
        assert got.value is False
        assert got.error_code == ErrorCode.FLAG_NOT_FOUND
        assert got.reason == Reason.ERROR


@pytest.mark.parametrize("evaluation_type", _EVALUATION_TYPES)
def test_should_error_if_we_expect_boolean_and_got_another_type(
    relay_proxy, evaluation_type
):
    with _client(relay_proxy, evaluation_type) as client:
        got = client.get_boolean_details(
            flag_key="string_key",
            default_value=False,
            evaluation_context=_default_evaluation_ctx,
        )
        assert got.flag_key == "string_key"
        assert got.value is False
        assert got.error_code == ErrorCode.TYPE_MISMATCH
        assert got.reason == Reason.ERROR


def test_both_evaluation_modes_agree(relay_proxy):
    """The WASM engine and the relay proxy must reach the same verdict.

    This is the assertion the mocked suite structurally cannot make: there, each
    mode is compared against its own recording, so the two could disagree
    indefinitely without any test noticing.
    """
    flag_keys = [
        "bool_targeting_match",
        "string_key",
        "integer_key",
        "double_key",
        "object_key",
    ]

    results = {}
    for evaluation_type in _EVALUATION_TYPES:
        with _client(relay_proxy, evaluation_type) as client:
            results[evaluation_type] = {
                key: client.get_object_details(
                    flag_key=key,
                    default_value=None,
                    evaluation_context=_default_evaluation_ctx,
                )
                for key in flag_keys
            }

    for key in flag_keys:
        inprocess = results[EvaluationType.INPROCESS][key]
        remote = results[EvaluationType.REMOTE][key]
        assert inprocess.value == remote.value, f"value differs for {key}"
        assert inprocess.variant == remote.variant, f"variant differs for {key}"
        assert inprocess.reason == remote.reason, f"reason differs for {key}"
