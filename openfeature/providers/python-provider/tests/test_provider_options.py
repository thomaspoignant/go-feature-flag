import pytest
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions
from gofeatureflag_python_provider.model import EvaluationType


class TestGoFeatureFlagProviderOptions:
    """Test cases for GoFeatureFlagProviderOptions."""

    def test_valid_options(self):
        """Test creating provider options with valid values."""
        options = GoFeatureFlagOptions(
            endpoint="https://example.com",
            evaluation_type=EvaluationType.REMOTE,
            timeout=5000,
            flag_change_polling_interval_ms=60000,
            data_flush_interval=60000,
            max_pending_events=5000,
            disable_data_collection=True,
            api_key="test-key",
        )

        assert options.endpoint == "https://example.com"
        assert options.evaluation_type == EvaluationType.REMOTE
        assert options.timeout == 5000
        assert options.flag_change_polling_interval_ms == 60000
        assert options.data_flush_interval == 60000
        assert options.max_pending_events == 5000
        assert options.disable_data_collection is True
        assert options.api_key == "test-key"

    def test_default_options(self):
        """Test creating provider options with default values."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")

        assert options.endpoint == "https://example.com"
        assert options.evaluation_type == EvaluationType.IN_PROCESS
        assert options.timeout == 10000
        assert options.flag_change_polling_interval_ms == 120000
        assert options.data_flush_interval == 120000
        assert options.max_pending_events == 10000
        assert options.disable_data_collection is False
        assert options.api_key is None
        assert options.exporter_metadata is None

    def test_validation_constraints(self):
        """Test that validation constraints are enforced."""
        # Test timeout constraint
        with pytest.raises(ValueError, match="Input should be greater than 0"):
            GoFeatureFlagOptions(endpoint="https://example.com", timeout=0)

        # Test flag_change_polling_interval_ms constraint
        with pytest.raises(ValueError, match="Input should be greater than 0"):
            GoFeatureFlagOptions(
                endpoint="https://example.com", flag_change_polling_interval_ms=-1
            )

        # Test data_flush_interval constraint
        with pytest.raises(ValueError, match="Input should be greater than 0"):
            GoFeatureFlagOptions(endpoint="https://example.com", data_flush_interval=0)

        # Test max_pending_events constraint
        with pytest.raises(ValueError, match="Input should be greater than 0"):
            GoFeatureFlagOptions(endpoint="https://example.com", max_pending_events=-5)
