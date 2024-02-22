import json
from http import HTTPStatus
from threading import Thread
from typing import List, Optional, Type, Union
from urllib.parse import urljoin

import pylru
import urllib3
import websocket
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import ErrorCode
from openfeature.exception import (
    FlagNotFoundError,
    TypeMismatchError,
    GeneralError,
    OpenFeatureError,
)
from openfeature.flag_evaluation import FlagEvaluationDetails
from openfeature.flag_evaluation import Reason
from openfeature.hook import Hook
from openfeature.provider.metadata import Metadata
from openfeature.provider.provider import AbstractProvider
from pydantic import ValidationError, PrivateAttr

from gofeatureflag_python_provider.data_collector_hook import DataCollectorHook
from gofeatureflag_python_provider.metadata import GoFeatureFlagMetadata
from gofeatureflag_python_provider.options import BaseModel
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.provider_status import ProviderStatus
from gofeatureflag_python_provider.request_flag_evaluation import (
    RequestFlagEvaluation,
    convert_evaluation_context,
)
from gofeatureflag_python_provider.response_flag_evaluation import (
    ResponseFlagEvaluation,
    JsonType,
)

AbstractProviderMetaclass = type(AbstractProvider)
BaseModelMetaclass = type(BaseModel)


class CombinedMetaclass(AbstractProviderMetaclass, BaseModelMetaclass):
    pass


class GoFeatureFlagProvider(BaseModel, AbstractProvider, metaclass=CombinedMetaclass):
    options: GoFeatureFlagOptions
    _http_client: urllib3.PoolManager = PrivateAttr()
    _cache: pylru.lrucache = PrivateAttr()
    _status: ProviderStatus = PrivateAttr(ProviderStatus.NOT_READY)
    _data_collector_hook: Optional[DataCollectorHook] = PrivateAttr()
    _ws: websocket.WebSocketApp = PrivateAttr()
    _ws_thread: Thread = PrivateAttr()

    def __init__(self, **data):
        """
        Constructor of the provider.
        It will initialize the http client for calling the GO Feature Flag relay proxy.

        :param data: data coming from pydantic configuration
        """
        super().__init__(**data)
        if self.options.urllib3_pool_manager is not None:
            self._http_client = self.options.urllib3_pool_manager
        else:
            self._http_client = urllib3.PoolManager(
                num_pools=100,
                timeout=urllib3.Timeout(connect=10, read=10),
                retries=urllib3.Retry(0),
            )
        self._data_collector_hook = DataCollectorHook(
            options=self.options,
            http_client=self._http_client,
        )
        websocket.enableTrace(self.options.debug)
        self._ws = websocket.WebSocketApp(
            self._build_websocket_uri(),
            on_message=self._websocket_message_handler,
            on_open=self._websocket_open_handler,
            on_close=self._websocket_close_handler,
            on_error=self._websocket_error_handler,
        )

    def get_status(self) -> ProviderStatus:
        """
        get_status returns the status of the provider
        :return: the status of the provider
        """
        return self._status

    def initialize(self, evaluation_context: EvaluationContext) -> None:
        """
        initialize is called when the provider is initialized.
        :param evaluation_context: the evaluation context
        :return: None
        """
        self._cache = pylru.lrucache(self.options.cache_size)
        self._data_collector_hook.initialize()
        # start the websocket thread
        if self.options.disable_cache_invalidation is False:
            self._ws_thread = Thread(target=self.run_websocket)
            self._ws_thread.start()
        else:
            self._status = ProviderStatus.READY

    def shutdown(self):
        if self.options.disable_cache_invalidation is False:
            self._ws.close(status=websocket.STATUS_NORMAL)
            self._ws_thread.join()

        if self._cache is not None:
            self._cache.clear()

        if self._data_collector_hook is not None:
            self._data_collector_hook.shutdown()
            self._data_collector_hook = None

        self._status = ProviderStatus.NOT_READY

    def get_metadata(self) -> Metadata:
        return GoFeatureFlagMetadata()

    def get_provider_hooks(self) -> List[Hook]:
        if self._data_collector_hook is None:
            return []
        return [self._data_collector_hook]

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[bool]:
        return self.generic_go_feature_flag_resolver(
            bool, flag_key, default_value, evaluation_context
        )

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[str]:
        return self.generic_go_feature_flag_resolver(
            str, flag_key, default_value, evaluation_context
        )

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[int]:
        return self.generic_go_feature_flag_resolver(
            int, flag_key, default_value, evaluation_context
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[float]:
        return self.generic_go_feature_flag_resolver(
            float, flag_key, default_value, evaluation_context
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: dict,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[Union[list, dict]]:
        return self.generic_go_feature_flag_resolver(
            Union[dict, list], flag_key, default_value, evaluation_context
        )

    def generic_go_feature_flag_resolver(
        self,
        original_type: Type[JsonType],
        flag_key: str,
        default_value: JsonType,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagEvaluationDetails[JsonType]:
        """
        generic_go_feature_flag_resolver is a generic evaluations of your flag with GO Feature Flag relay proxy it works
        with all types.

        :param original_type: type of the request
        :param flag_key:  name of the flag
        :param default_value: default value of the flag
        :param evaluation_context: context to evaluate the flag
        :return: a FlagEvaluationDetails object containing the response for the SDK.
        """
        try:
            if self._status != ProviderStatus.READY:
                return FlagEvaluationDetails[original_type](
                    flag_key=flag_key,
                    value=default_value,
                    reason=Reason.ERROR,
                    error_code=ErrorCode.PROVIDER_NOT_READY,
                    error_message="GO Feature Flag provider is not ready",
                )

            goff_evaluation_context = convert_evaluation_context(evaluation_context)
            goff_request = RequestFlagEvaluation(
                user=goff_evaluation_context,
                defaultValue=default_value,
            )
            evaluation_context_hash = goff_evaluation_context.hash()
            is_from_cache = False

            if evaluation_context_hash in self._cache:
                response_body = self._cache[evaluation_context_hash]
                is_from_cache = True
            else:
                response = self._http_client.request(
                    method="POST",
                    url=urljoin(
                        str(self.options.endpoint),
                        "/v1/feature/{}/eval".format(flag_key),
                    ),
                    headers={"Content-Type": "application/json"},
                    body=goff_request.model_dump_json(),
                )

                if response.status == HTTPStatus.NOT_FOUND.value:
                    raise FlagNotFoundError(
                        "flag {} was not found in your configuration".format(flag_key)
                    )

                if int(response.status) >= HTTPStatus.BAD_REQUEST.value:
                    raise GeneralError(
                        "impossible to contact GO Feature Flag relay proxy instance"
                    )
                response_body = response.data

            response_flag_evaluation = ResponseFlagEvaluation.model_validate_json(
                response_body
            )

            if response_flag_evaluation.cacheable:
                self._cache[evaluation_context_hash] = response_body

            if original_type == int:
                response_json = json.loads(response_body)
                # in some cases, pydantic auto convert float in int.
                if type(response_json.get("value")) != int:
                    raise TypeMismatchError(
                        "unexpected type for flag {}".format(flag_key)
                    )

            if response_flag_evaluation.reason == Reason.DISABLED.value:
                return FlagEvaluationDetails[original_type](
                    flag_key=flag_key,
                    value=default_value,
                    reason=Reason.DISABLED,
                )

            if response_flag_evaluation.errorCode == ErrorCode.FLAG_NOT_FOUND.value:
                raise FlagNotFoundError(
                    "flag {} was not found in your configuration".format(flag_key)
                )

            return FlagEvaluationDetails[original_type](
                flag_key=flag_key,
                value=response_flag_evaluation.value,
                variant=response_flag_evaluation.variationType,
                reason=(
                    Reason.CACHED if is_from_cache else response_flag_evaluation.reason
                ),
                flag_metadata=response_flag_evaluation.metadata,
            )
        except ValidationError as exc:
            raise TypeMismatchError(
                "unexpected type for flag {}: {}".format(flag_key, exc)
            )

        except OpenFeatureError as exc:
            raise exc

        except Exception as exc:
            raise GeneralError(
                "unexpected error while evaluating flag {}: {}".format(flag_key, exc)
            )

    def _build_websocket_uri(self):
        """
        _build_websocket_uri is a helper to build the websocket uri to connect to the GO Feature Flag relay proxy.
        :return: a string representing the websocket uri
        """
        http_uri = urljoin(
            str(self.options.endpoint),
            "/ws/v1/flag/change",
        )
        http_uri = http_uri.replace("http", "ws")
        http_uri = http_uri.replace("https", "wss")
        return http_uri

    def run_websocket(self) -> None:
        """
        run_websocket is a helper to run the websocket connection to the GO Feature Flag server.
        :return: None
        """
        self._ws.run_forever(reconnect=self.options.reconnect_interval)

    def _websocket_message_handler(self, wsapp, message) -> None:
        """
        websocket_message_handler is the handler called when we receive a message from the GO Feature Flag server
        :param wsapp: the websocket app
        :param message: the message received
        :return: None
        """
        # when we receive a message from go-feature-flag server, we clear the cache.
        self._cache.clear()

    def _websocket_open_handler(self, ws_app) -> None:
        """
        websocket_open_handler is the handler called when the websocket is open
        :param ws app: the websocket app
        :return: None
        """
        self._status = ProviderStatus.READY

    def _websocket_error_handler(self, ws_app, error) -> None:
        """
        websocket_error_handler is the handler called when we receive an error from the GO Feature Flag server
        :param ws_app: the websocket app
        :param error: error received
        :return: None
        """
        self._status = ProviderStatus.ERROR

    def _websocket_close_handler(self, ws_app, close_status_code, close_msg) -> None:
        """
        websocket_close_handler is the handler called when the websocket is closed
        :param wsapp: the websocket app
        :param close_status_code: the status code of the close
        :param close_msg: the message of the close
        :return: None
        """
        self._status = ProviderStatus.STALE
