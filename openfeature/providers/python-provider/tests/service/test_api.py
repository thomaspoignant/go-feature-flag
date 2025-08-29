import datetime
import json
import os
from pathlib import Path
from unittest.mock import Mock, patch

import pytest
from urllib3.exceptions import HTTPError

from gofeatureflag_python_provider.exception import (
    FlagConfigurationEndpointNotFoundException,
    ImpossibleToRetrieveConfigurationException,
    InvalidOptionsException,
    UnauthorizedException,
)
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions
from gofeatureflag_python_provider.service.api import Api


class TestApi:
    """Test class for Api following the TypeScript test structure."""

    def test_constructor_should_throw_if_options_are_missing(self):
        """Test that constructor throws if options are missing."""

        with pytest.raises(InvalidOptionsException):
            Api(options=None)

    @pytest.mark.usefixtures("goff")
    class TestIntegrationUsingDockerCompose:
        """Test group for integration using docker compose."""

        def test_should_call_the_configuration_endpoint(self, goff):
            """Test that the configuration endpoint is called."""

            api = Api(
                options=GoFeatureFlagOptions(
                    endpoint=goff,
                    api_key="my-api-key",
                )
            )
            config_response = api.retrieve_flag_configuration()
            assert config_response is not None

            # check that config_response.last_updated is a datetime
            assert isinstance(config_response.last_updated, datetime.datetime)
            assert isinstance(config_response.flags, dict)
            assert config_response.evaluation_context_enrichment["testenv"] == "pytest"
            assert len(config_response.flags.keys()) == 12
            assert config_response.flags["string_key_with_version"] is not None
            assert config_response.evaluation_context_enrichment is not None

    class TestRetrieveFlagConfiguration:
        """Test group for RetrieveFlagConfiguration method."""

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_call_the_configuration_endpoint2(self, mock_request):
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/default.json"),
                headers={
                    "last-modified": "2021-01-01T00:00:00Z",
                    "etag": '"1234567890"',
                },
            )
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            api.retrieve_flag_configuration()
            mock_request.assert_called_once_with(
                method="POST",
                url="http://goff.local/v1/flag/configuration",
                headers={"Content-Type": "application/json"},
                body=json.dumps({"flags": []}, separators=(",", ":")),
            )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_if_none_match_header_when_etag_is_provided(
            self, mock_request
        ):
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/default.json"),
                headers={
                    "last-modified": "2021-01-01T00:00:00Z",
                    "etag": '"1234567890"',
                },
            )
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))

            # Call with etag parameter
            api.retrieve_flag_configuration(etag='"abc123"')

            # Verify the request includes the If-None-Match header
            mock_request.assert_called_once_with(
                method="POST",
                url="http://goff.local/v1/flag/configuration",
                headers={
                    "Content-Type": "application/json",
                    "If-None-Match": '"abc123"',
                },
                body=json.dumps({"flags": []}, separators=(",", ":")),
            )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_flags_in_request_body_when_provided(self, mock_request):
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/default.json"),
                headers={
                    "last-modified": "2021-01-01T00:00:00Z",
                    "etag": '"1234567890"',
                },
            )
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            api.retrieve_flag_configuration(flags=["flag1", "flag2", "flag3"])

            # Verify the request includes the flags in the request body
            mock_request.assert_called_once_with(
                method="POST",
                url="http://goff.local/v1/flag/configuration",
                headers={"Content-Type": "application/json"},
                body=json.dumps(
                    {"flags": ["flag1", "flag2", "flag3"]}, separators=(",", ":")
                ),
            )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_unauthorized_exception_on_401_response(
            self, mock_request
        ):
            """Test that UnauthorizedException is thrown on 401 response."""
            mock_request.return_value = Mock(
                status=401,
                data=b"Unauthorized",
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(UnauthorizedException) as exc_info:
                api.retrieve_flag_configuration()
            assert "authentication/authorization error" in str(exc_info.value)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_unauthorized_exception_on_403_response(
            self, mock_request
        ):
            """Test that UnauthorizedException is thrown on 403 response."""
            mock_request.return_value = Mock(
                status=403,
                data=b"Forbidden",
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(UnauthorizedException) as exc_info:
                api.retrieve_flag_configuration()

            assert "authentication/authorization error" in str(exc_info.value)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_impossible_to_retrieve_configuration_exception_on_400_response(
            self, mock_request
        ):
            """Test that ImpossibleToRetrieveConfigurationException is thrown on 400 response."""
            mock_request.return_value = Mock(
                status=400,
                data=b"Bad Request",
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToRetrieveConfigurationException) as exc_info:
                api.retrieve_flag_configuration()

            assert "Bad request: Bad Request" in str(exc_info.value)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_impossible_to_retrieve_configuration_exception_on_500_response(
            self, mock_request
        ):
            """Test that ImpossibleToRetrieveConfigurationException is thrown on 500 response."""
            mock_request.return_value = Mock(
                status=500,
                data=b"Internal Server Error",
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToRetrieveConfigurationException) as exc_info:
                api.retrieve_flag_configuration()

            assert "unexpected http code 500: Internal Server Error" in str(
                exc_info.value
            )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_flag_configuration_endpoint_not_found_exception_on_404_response(
            self, mock_request
        ):
            """Test that FlagConfigurationEndpointNotFoundException is thrown on 404 response."""
            mock_request.return_value = Mock(
                status=404,
                data=b"Not Found",
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(FlagConfigurationEndpointNotFoundException):
                api.retrieve_flag_configuration()

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_return_valid_flag_config_response_on_200_response(
            self, mock_request
        ):
            """Test that valid FlagConfigResponse is returned on 200 response."""
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/default.json"),
                headers={
                    "last-modified": "2021-01-01T00:00:00Z",
                    "ETag": '"1234567890"',
                },
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            response = api.retrieve_flag_configuration()

            assert response is not None
            assert response.etag == '"1234567890"'
            assert isinstance(response.last_updated, datetime.datetime)
            assert isinstance(response.flags, dict)
            assert len(response.flags) == 13
            assert isinstance(response.evaluation_context_enrichment, dict)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_304_response_without_flags_and_context(
            self, mock_request
        ):
            """Test that 304 response is handled without flags and context."""
            mock_request.return_value = Mock(
                status=304,
                data=b"",
                headers={
                    "etag": '"1234567890"',
                },
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            response = api.retrieve_flag_configuration()

            assert response is not None
            assert isinstance(response.last_updated, datetime.datetime)
            assert response.flags == {}
            assert response.evaluation_context_enrichment == {}

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_invalid_last_modified_header(self, mock_request):
            """Test that invalid last-modified header is handled."""
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/default.json"),
                headers={
                    "last-modified": "invalid-date",
                    "etag": '"1234567890"',
                },
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            response = api.retrieve_flag_configuration()

            assert response is not None
            # Should fall back to current datetime when header is invalid
            assert isinstance(response.last_updated, datetime.datetime)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_network_errors(self, mock_request):
            """Test that network errors are handled."""

            mock_request.side_effect = HTTPError("Network error")
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))

            with pytest.raises(ImpossibleToRetrieveConfigurationException) as exc_info:
                api.retrieve_flag_configuration()

            assert "Network error: Network error" in str(exc_info.value)

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_timeout(self, mock_request):
            """Test that timeout is handled."""
            mock_request.side_effect = TimeoutError("Timeout")
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToRetrieveConfigurationException) as exc_info:
                api.retrieve_flag_configuration()

            assert "Network error: Timeout" in str(exc_info.value)

    # class TestSendEventToDataCollector:
    #     """Test group for SendEventToDataCollector method."""

    #     def test_should_call_the_data_collector_endpoint(self, api):
    #         """Test that the data collector endpoint is called."""
    #         pass

    #     def test_should_include_api_key_in_authorization_header_when_provided(
    #         self, base_options
    #     ):
    #         """Test that API key is included in authorization header when provided."""
    #         pass

    #     def test_should_not_include_authorization_header_when_api_key_is_not_provided(
    #         self, api
    #     ):
    #         """Test that authorization header is not included when API key is not provided."""
    #         pass

    #     def test_should_include_content_type_header(self, api):
    #         """Test that content-type header is included."""
    #         pass

    #     def test_should_include_events_and_metadata_in_request_body(self, api):
    #         """Test that events and metadata are included in request body."""
    #         pass

    #     def test_should_handle_tracking_events(self, api):
    #         """Test that tracking events are handled."""
    #         pass

    #     def test_should_throw_unauthorized_exception_on_401_response(self, api):
    #         """Test that UnauthorizedException is thrown on 401 response."""
    #         pass

    #     def test_should_throw_unauthorized_exception_on_403_response(self, api):
    #         """Test that UnauthorizedException is thrown on 403 response."""
    #         pass

    #     def test_should_throw_impossible_to_send_data_to_the_collector_exception_on_400_response(
    #         self, api
    #     ):
    #         """Test that ImpossibleToSendDataToTheCollectorException is thrown on 400 response."""
    #         pass

    #     def test_should_throw_impossible_to_send_data_to_the_collector_exception_on_500_response(
    #         self, api
    #     ):
    #         """Test that ImpossibleToSendDataToTheCollectorException is thrown on 500 response."""
    #         pass

    #     def test_should_handle_network_errors(self, base_options):
    #         """Test that network errors are handled."""
    #         pass

    #     def test_should_handle_timeout(self, base_options):
    #         """Test that timeout is handled."""
    #         pass


def _read_mock_file(relative_path: str) -> str:
    """Read a mock file from the mock_responses directory."""

    # This hacky if is here to make test run inside pycharm and from the root of the project
    if os.getcwd().endswith("/tests"):
        path_prefix = "./mock_responses/{}"
    else:
        path_prefix = "./tests/mock_responses/{}"
    return Path(path_prefix.format(relative_path)).read_text(encoding="utf-8")
