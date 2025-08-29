import pytest
from unittest.mock import Mock, AsyncMock
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions
from gofeatureflag_python_provider.model import EvaluationType
from gofeatureflag_python_provider.exception import InvalidOptionsException


class TestGoFeatureFlagProvider:
    """Test cases for GoFeatureFlagProvider."""

    def test_provider_creation_with_valid_options(self):
        """Test creating provider with valid options."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")
        provider = GoFeatureFlagProvider(options)

        assert provider.options == options
        assert provider.get_metadata().name == "GoFeatureFlagProvider"
        assert provider.get_provider_hooks() == []

    def test_provider_creation_with_invalid_endpoint(self):
        """Test creating provider with invalid endpoint."""
        with pytest.raises(
            InvalidOptionsException, match="endpoint is a mandatory field"
        ):
            GoFeatureFlagOptions(endpoint="")

        with pytest.raises(
            InvalidOptionsException, match="endpoint is a mandatory field"
        ):
            GoFeatureFlagOptions(endpoint="   ")

    def test_provider_creation_with_invalid_url(self):
        """Test creating provider with invalid URL."""
        with pytest.raises(
            InvalidOptionsException, match="endpoint must be a valid URL"
        ):
            GoFeatureFlagOptions(endpoint="not-a-url")

        with pytest.raises(
            InvalidOptionsException, match="endpoint must be a valid URL"
        ):
            GoFeatureFlagOptions(endpoint="ftp://example.com")

    def test_provider_creation_with_no_options(self):
        """Test creating provider with no options."""
        with pytest.raises(InvalidOptionsException, match="No options provided"):
            GoFeatureFlagProvider(None)

    def test_remote_evaluator_creation(self):
        """Test creating provider with remote evaluation."""
        options = GoFeatureFlagOptions(
            endpoint="https://example.com", evaluation_type=EvaluationType.REMOTE
        )
        provider = GoFeatureFlagProvider(options)

        # The evaluator should be a RemoteEvaluator
        assert provider.evaluator is not None
        # Note: We can't easily test the type without importing the actual class

    def test_inprocess_evaluator_creation(self):
        """Test creating provider with in-process evaluation."""
        options = GoFeatureFlagOptions(
            endpoint="https://example.com", evaluation_type=EvaluationType.IN_PROCESS
        )
        provider = GoFeatureFlagProvider(options)

        # The evaluator should be an InProcessEvaluator
        assert provider.evaluator is not None
        # Note: We can't easily test the type without importing the actual class

    def test_track_method(self):
        """Test the track method."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")
        provider = GoFeatureFlagProvider(options)

        # Test tracking without context
        provider.track("test_event")

        # Test tracking with context
        mock_context = Mock()
        mock_context.targeting_key = "user123"
        mock_context.attributes = {"kind": "user", "email": "test@example.com"}

        provider.track("test_event", mock_context, {"custom": "data"})

    def test_sync_resolution_methods(self):
        """Test synchronous resolution methods."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")
        provider = GoFeatureFlagProvider(options)

        # Test boolean resolution
        result = provider.resolve_boolean_details("test_flag", True)
        assert result.value is True
        assert result.reason == "DEFAULT"

        # Test string resolution
        result = provider.resolve_string_details("test_flag", "default_value")
        assert result.value == "default_value"
        assert result.reason == "DEFAULT"

        # Test integer resolution
        result = provider.resolve_integer_details("test_flag", 42)
        assert result.value == 42
        assert result.reason == "DEFAULT"

        # Test float resolution
        result = provider.resolve_float_details("test_flag", 3.14)
        assert result.value == 3.14
        assert result.reason == "DEFAULT"

        # Test object resolution
        result = provider.resolve_object_details("test_flag", {"key": "value"})
        assert result.value == {"key": "value"}
        assert result.reason == "DEFAULT"

    @pytest.mark.asyncio
    async def test_initialize_and_shutdown(self):
        """Test provider initialization and shutdown."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")
        provider = GoFeatureFlagProvider(options)

        # Mock the evaluator
        mock_evaluator = AsyncMock()
        provider.evaluator = mock_evaluator

        # Test initialization
        await provider.initialize()
        mock_evaluator.initialize.assert_called_once()

        # Test shutdown
        provider.shutdown()

    def test_context_kind_extraction(self):
        """Test extracting context kind from evaluation context."""
        options = GoFeatureFlagOptions(endpoint="https://example.com")
        provider = GoFeatureFlagProvider(options)

        # Test with context that has attributes
        mock_context = Mock()
        mock_context.attributes = {"kind": "organization"}
        assert provider._get_context_kind(mock_context) == "organization"

        # Test with context that has no kind attribute
        mock_context.attributes = {"email": "test@example.com"}
        assert provider._get_context_kind(mock_context) == "user"

        # Test with no context
        assert provider._get_context_kind(None) == "user"
