# import json
# import os
# import time

# import pytest
# import requests
# import yaml
# from openfeature import api
# from openfeature.flag_evaluation import FlagEvaluationDetails, Reason, ErrorCode

# from gofeatureflag_python_provider.options import GoFeatureFlagOptions
# from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
# from tests.test_gofeatureflag_python_provider import (
#     _default_evaluation_ctx,
# )


# def is_responsive(url):
#     try:
#         response = requests.get(url)
#         if response.status_code == 200:
#             return True
#     except requests.exceptions.ConnectionError:
#         return False


# @pytest.fixture(scope="session")
# def goff(docker_ip, docker_services):
#     """Ensure that HTTP service is up and responsive."""

#     # `port_for` takes a container port and returns the corresponding host port
#     port = docker_services.port_for("goff", 1031)
#     url = "http://{}:{}".format(docker_ip, port)
#     docker_services.wait_until_responsive(
#         timeout=30.0, pause=0.1, check=lambda: is_responsive(url + "/health")
#     )
#     return url


# def test_test_websocket_cache_invalidation(goff):
#     """
#     test that the cache is invalidated when the config change
#     :param goff: the fixture to launch goff container
#     """
#     flag_key = "bool_targeting_match"
#     default_value = False

#     goff_provider = GoFeatureFlagProvider(
#         options=GoFeatureFlagOptions(
#             endpoint=goff,
#             data_flush_interval=100,
#             disable_data_collection=True,
#             api_key="apikey1",
#         )
#     )
#     api.set_provider(goff_provider)
#     client = api.get_client(domain="test-client")

#     want = FlagEvaluationDetails(
#         flag_key=flag_key,
#         value=True,
#         variant="True",
#         reason=Reason.TARGETING_MATCH,
#         flag_metadata={
#             "description": "this is a test",
#             "pr_link": "https://github.com/thomaspoignant/go-feature-flag/pull/916",
#         },
#     )
#     got = client.get_boolean_details(
#         flag_key=flag_key,
#         default_value=default_value,
#         evaluation_context=_default_evaluation_ctx,
#     )
#     assert got == want

#     got = client.get_boolean_details(
#         flag_key=flag_key,
#         default_value=default_value,
#         evaluation_context=_default_evaluation_ctx,
#     )
#     want.reason = Reason.CACHED
#     assert got == want

#     # test https://github.com/thomaspoignant/go-feature-flag/issues/3613
#     got = client.get_boolean_details(
#         flag_key="nonexistent-flag-key",
#         default_value=False,
#         evaluation_context=_default_evaluation_ctx,
#     )
#     assert got.error_code == ErrorCode.FLAG_NOT_FOUND

#     modify_flag_config()
#     got = client.get_boolean_details(
#         flag_key=flag_key,
#         default_value=default_value,
#         evaluation_context=_default_evaluation_ctx,
#     )
#     want.reason = Reason.TARGETING_MATCH
#     assert got == want
#     api.shutdown()


# def modify_flag_config():
#     """
#     modify the config file to trigger the websocket cache invalidation
#     :return: None
#     """
#     file_location = "{}/tests/config.goff.yaml".format(os.getcwd())
#     print()
#     with open(file_location) as f:
#         list_doc = yaml.safe_load(f)
#         initial_doc = json.loads(json.dumps(list_doc))

#     list_doc["string_key"]["disable"] = True

#     with open(file_location, "w") as f:
#         yaml.dump(list_doc, f)

#     time.sleep(1.4)  # we wait to let the time of the polling to happen in the container
#     with open(file_location, "w") as f:
#         yaml.dump(initial_doc, f)
