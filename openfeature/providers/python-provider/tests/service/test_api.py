import datetime
import json
import os
from pathlib import Path
from unittest.mock import Mock, patch

import pytest
import urllib3
from urllib3.exceptions import HTTPError

from gofeatureflag_python_provider.exception import (
    FlagConfigurationEndpointNotFoundException,
    ImpossibleToRetrieveConfigurationException,
    InvalidOptionsException,
    UnauthorizedException,
)
from gofeatureflag_python_provider.exception.impossible_to_send_data_to_collector_exception import (
    ImpossibleToSendDataToCollectorException,
)
from gofeatureflag_python_provider.model.exporter_metadata import ExporterMetadata
from gofeatureflag_python_provider.model.feature_event import FeatureEvent
from gofeatureflag_python_provider.model.tracking_event import TrackingEvent
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

        def test_should_call_the_configuration_endpoint_with_custom_pool_manager(
            self, goff
        ):
            """Test that the configuration endpoint is called."""

            api = Api(
                options=GoFeatureFlagOptions(
                    endpoint=goff,
                    api_key="my-api-key",
                ),
                urllib3_pool_manager=urllib3.PoolManager(
                    num_pools=100,
                    timeout=urllib3.Timeout(connect=10, read=10),
                    retries=urllib3.Retry(0),
                ),
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
            """Test that the configuration endpoint is called."""
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
        def test_should_log_warning_on_invalid_json_response(self, mock_request):
            """Test that warning is logged on invalid JSON response."""
            mock_request.return_value = Mock(
                status=200,
                data=_read_mock_file("flag-configuration/invalid-json.json"),
                headers={
                    "last-modified": "2021-01-01T00:00:00Z",
                    "etag": '"1234567890"',
                },
            )

            logger_mock = Mock()
            api = Api(
                options=GoFeatureFlagOptions(endpoint="http://goff.local/"),
                logger=logger_mock,
            )
            api.retrieve_flag_configuration()
            mock_request.assert_called_once_with(
                method="POST",
                url="http://goff.local/v1/flag/configuration",
                headers={"Content-Type": "application/json"},
                body=json.dumps({"flags": []}, separators=(",", ":")),
            )
            logger_mock.warning.assert_called_once_with(
                'Failed to parse flag configuration response: Expecting property name enclosed in double quotes: line 1 column 2 (char 1). Response body: "{"'
            )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_if_none_match_header_when_etag_is_provided(
            self, mock_request
        ):
            """Test that the If-None-Match header is included when etag is provided."""
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
            """Test that the flags are included in the request body when provided."""
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

    class TestSendEventToDataCollector:
        """Test group for SendEventToDataCollector method."""

        single_event_list = [
            TrackingEvent(
                kind="tracking",
                context_kind="anonymousUser",
                user_key="ABCD",
                creation_date=1617970548,
                key="xxx",
                evaluation_context={
                    "firstname": "john",
                    "rate": 3.14,
                    "targetingKey": "d45e303a-38c2-11ed-a261-0242ac120002",
                    "company_info": {"size": 120, "name": "my_company"},
                    "anonymous": False,
                    "email": "john.doe@gofeatureflag.org",
                    "age": 30,
                    "lastname": "doe",
                    "professional": True,
                    "labels": ["pro", "beta"],
                },
                tracking_event_details={"toto": 123},
            )
        ]

        tracking_plus_feature_event_list = [
            TrackingEvent(
                kind="tracking",
                context_kind="anonymousUser",
                user_key="ABCD",
                creation_date=1617970548,
                key="xxx",
                evaluation_context={
                    "firstname": "john",
                    "rate": 3.14,
                    "targetingKey": "d45e303a-38c2-11ed-a261-0242ac120002",
                    "company_info": {"size": 120, "name": "my_company"},
                    "anonymous": False,
                    "email": "john.doe@gofeatureflag.org",
                    "age": 30,
                    "lastname": "doe",
                    "professional": True,
                    "labels": ["pro", "beta"],
                },
                tracking_event_details={"toto": 123},
            ),
            FeatureEvent(
                context_kind="anonymousUser",
                creation_date=1617970547,
                key="xxx",
                kind="feature",
                user_key="ABCD",
                value=True,
                variation="enabled",
                version=None,
                default=False,
            ),
        ]

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_call_the_data_collector_endpoint(self, mock_request):
            """Test that the data collector endpoint is called."""
            mock_request.return_value = Mock(
                status=200,
                data='{"ingestedContentCount": 1}',
                headers={},
            )
            logger_mock = Mock()
            api = Api(
                options=GoFeatureFlagOptions(endpoint="http://goff.local/"),
                logger=logger_mock,
            )
            want = json.loads(
                _read_mock_file("data-collector/request/single_event_list.json")
            )

            api.send_event_to_data_collector(self.single_event_list, ExporterMetadata())
            logger_mock.info.assert_called_once_with(
                'Published 1 events successfully: {"ingestedContentCount": 1}'
            )
            mock_request.assert_called_once()
            call_args = mock_request.call_args
            assert call_args[1]["method"] == "POST"
            assert call_args[1]["url"] == "http://goff.local/v1/data/collector"
            assert call_args[1]["headers"] == {"Content-Type": "application/json"}
            actual_body = json.loads(call_args[1]["body"])
            assert actual_body == want

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_api_key_in_authorization_header_when_provided(
            self, mock_request
        ):
            """Test that API key is included in authorization header when provided."""
            mock_request.return_value = Mock(
                status=200,
                data='{"ingestedContentCount": 1}',
                headers={},
            )
            api = Api(
                options=GoFeatureFlagOptions(
                    endpoint="http://goff.local/", api_key="test-key"
                )
            )
            want = json.loads(
                _read_mock_file("data-collector/request/single_event_list.json")
            )

            api.send_event_to_data_collector(self.single_event_list, ExporterMetadata())

            mock_request.assert_called_once()
            call_args = mock_request.call_args
            assert call_args[1]["method"] == "POST"
            assert call_args[1]["url"] == "http://goff.local/v1/data/collector"
            assert call_args[1]["headers"] == {
                "Content-Type": "application/json",
                "Authorization": "Bearer test-key",
            }
            actual_body = json.loads(call_args[1]["body"])
            assert actual_body == want

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_content_type_header(self, mock_request):
            """Test that content-type header is included."""
            mock_request.return_value = Mock(
                status=200,
                data='{"ingestedContentCount": 1}',
                headers={},
            )
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            want = json.loads(
                _read_mock_file("data-collector/request/single_event_list.json")
            )

            api.send_event_to_data_collector(self.single_event_list, ExporterMetadata())

            mock_request.assert_called_once()
            call_args = mock_request.call_args
            assert call_args[1]["method"] == "POST"
            assert call_args[1]["url"] == "http://goff.local/v1/data/collector"
            assert call_args[1]["headers"] == {
                "Content-Type": "application/json",
            }
            actual_body = json.loads(call_args[1]["body"])
            assert actual_body == want

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_include_events_and_metadata_in_request_body(self, mock_request):
            """Test that events and metadata are included in request body."""
            mock_request.return_value = Mock(
                status=200,
                data='{"ingestedContentCount": 1}',
                headers={},
            )

            metadata = ExporterMetadata()
            metadata.add("test", "test")
            metadata.add("test2", 123)
            metadata.add("test3", True)

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            want = json.loads(
                _read_mock_file("data-collector/request/single_event_list_metdata.json")
            )
            api.send_event_to_data_collector(self.single_event_list, metadata)
            mock_request.assert_called_once()
            call_args = mock_request.call_args
            assert call_args[1]["method"] == "POST"
            assert call_args[1]["url"] == "http://goff.local/v1/data/collector"
            assert call_args[1]["headers"] == {
                "Content-Type": "application/json",
            }
            actual_body = json.loads(call_args[1]["body"])
            assert actual_body == want

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_tracking_and_feature_events(self, mock_request):
            """Test that tracking events are handled."""
            mock_request.return_value = Mock(
                status=200,
                data='{"ingestedContentCount": 2}',
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            want = json.loads(
                _read_mock_file(
                    "data-collector/request/tracking_plus_feature_event_list.json"
                )
            )
            api.send_event_to_data_collector(
                self.tracking_plus_feature_event_list, ExporterMetadata()
            )
            mock_request.assert_called_once()
            call_args = mock_request.call_args
            assert call_args[1]["method"] == "POST"
            assert call_args[1]["url"] == "http://goff.local/v1/data/collector"
            assert call_args[1]["headers"] == {
                "Content-Type": "application/json",
            }
            actual_body = json.loads(call_args[1]["body"])
            assert actual_body == want

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_unauthorized_exception_on_401_response(
            self, mock_request
        ):
            """Test that UnauthorizedException is thrown on 401 response."""
            mock_request.return_value = Mock(
                status=401,
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(UnauthorizedException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_unauthorized_exception_on_403_response(
            self, mock_request
        ):
            """Test that UnauthorizedException is thrown on 403 response."""
            mock_request.return_value = Mock(
                status=403,
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(UnauthorizedException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_impossible_to_send_data_to_the_collector_exception_on_400_response(
            self, mock_request
        ):
            """Test that ImpossibleToSendDataToTheCollectorException is thrown on 400 response."""
            mock_request.return_value = Mock(
                status=400,
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToSendDataToCollectorException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_throw_impossible_to_send_data_to_the_collector_exception_on_500_response(
            self, mock_request
        ):
            """Test that ImpossibleToSendDataToTheCollectorException is thrown on 500 response."""
            mock_request.return_value = Mock(
                status=500,
                headers={},
            )

            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToSendDataToCollectorException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_network_errors(self, mock_request):
            """Test that network errors are handled."""
            mock_request.side_effect = HTTPError("Network error")
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToSendDataToCollectorException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )

        @patch("urllib3.poolmanager.PoolManager.request")
        def test_should_handle_timeout(self, mock_request):
            """Test that timeout is handled."""
            mock_request.side_effect = TimeoutError("Timeout")
            api = Api(options=GoFeatureFlagOptions(endpoint="http://goff.local/"))
            with pytest.raises(ImpossibleToSendDataToCollectorException):
                api.send_event_to_data_collector(
                    self.single_event_list, ExporterMetadata()
                )


def _read_mock_file(relative_path: str) -> str:
    """Read a mock file from the mock_responses directory."""

    # This hacky if is here to make test run inside pycharm and from the root of the project
    if os.getcwd().endswith("/tests"):
        path_prefix = "./mock_responses/{}"
    else:
        path_prefix = "./tests/mock_responses/{}"
    return Path(path_prefix.format(relative_path)).read_text(encoding="utf-8")
