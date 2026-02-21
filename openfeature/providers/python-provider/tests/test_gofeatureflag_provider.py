import pydantic
import pytest
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider


def test_constructor_options_empty():
    with pytest.raises(pydantic.ValidationError):
        GoFeatureFlagProvider()
