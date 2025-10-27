import datetime
import time
import urllib3
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import (
    FeatureEvent,
    RequestDataCollector,
)
from http import HTTPStatus
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason
from openfeature.hook import Hook, HookContext
from threading import Thread, Event
from typing import Optional
from urllib.parse import urljoin

default_targeting_key = "undefined-targetingKey"


class DataCollectorHook(Hook):
    # _thread_stopper is used to stop the background task when we shut down the hook
    _thread_stopper: Optional[Event] = None
    # _thread_data_collector is the thread used to call the relay proxy to collect data
    _thread_data_collector: Optional[Thread] = None
    # _options is the options of the provider
    _options: Optional[GoFeatureFlagOptions] = None
    # _data_collector_endpoint is the endpoint of the relay proxy
    _data_collector_endpoint: str
    # _http_client is the http client used to call the relay proxy
    _http_client: urllib3.PoolManager = None
    # _data_event_queue is the list of data to collect
    _event_queue: list[FeatureEvent] = []
    # _exporter_metadata is the metadata we send to the GO Feature Flag relay proxy when we report the evaluation data usage.
    _exporter_metadata: dict = {}

    def __init__(self, options: GoFeatureFlagOptions, http_client: urllib3.PoolManager):
        self._http_client = http_client
        self._thread_data_collector = Thread(target=self.background_task, daemon=True)
        self._options = options
        self._data_collector_endpoint = urljoin(
            str(self._options.endpoint), "/v1/data/collector"
        )
        self._exporter_metadata = options.exporter_metadata
        self._exporter_metadata["provider"] = "python"
        self._exporter_metadata["openfeature"] = True

    def after(
        self, hook_context: HookContext, details: FlagEvaluationDetails, hints: dict
    ):
        if self._options.disable_data_collection or details.reason != Reason.CACHED:
            # we don't collect data if the data collection is disabled or if the flag is not cached
            return
        feature_event = FeatureEvent(
            contextKind=(
                "anonymousUser"
                if hook_context.evaluation_context.attributes.get("anonymous", False)
                else "user"
            ),
            creationDate=int(datetime.datetime.now().timestamp()),
            default=False,
            key=hook_context.flag_key,
            value=details.value,
            variation=details.variant or "SdkDefault",
            userKey=hook_context.evaluation_context.targeting_key
            or default_targeting_key,
        )
        self._event_queue.append(feature_event)

    def error(self, hook_context: HookContext, exception: Exception, hints: dict):
        if self._options.disable_data_collection:
            # we don't collect data if the data collection is disabled
            return

        feature_event = FeatureEvent(
            contextKind=(
                "anonymousUser"
                if hook_context.evaluation_context.attributes.get("anonymous", False)
                else "user"
            ),
            creationDate=int(datetime.datetime.now().timestamp()),
            default=True,
            key=hook_context.flag_key,
            value=hook_context.default_value,
            variation="SdkDefault",
            userKey=hook_context.evaluation_context.targeting_key
            or default_targeting_key,
        )
        self._event_queue.append(feature_event)

    def initialize(self):
        self._event_queue = []
        self._thread_stopper = Event()
        self._thread_data_collector.start()

    def shutdown(self):
        # setting the _thread_stopper to True will stop the background task
        self._thread_stopper.set()
        self._thread_data_collector.join()
        self._collect_data()

    def background_task(self):
        while not self._thread_stopper.is_set():
            waiting_time = self._options.data_flush_interval / 1000
            self._thread_stopper.wait(waiting_time)
            self._collect_data()

    def _collect_data(self):
        if len(self._event_queue) > 0:
            try:
                goff_request = RequestDataCollector(
                    meta=self._exporter_metadata,
                    events=self._event_queue,
                )
                headers = {"Content-Type": "application/json"}
                if self._options.api_key:
                    headers["Authorization"] = "Bearer {}".format(self._options.api_key)

                response = self._http_client.request(
                    method="POST",
                    url=urljoin(str(self._options.endpoint), "/v1/data/collector"),
                    headers=headers,
                    body=goff_request.model_dump_json(),
                )

                if int(response.status) >= HTTPStatus.BAD_REQUEST.value:
                    print(
                        "impossible to contact GO Feature Flag relay proxy instance to collect the data, http_code: {}".format(
                            response.status
                        )
                    )
                    return

                # if the response is ok, we empty the queue
                self._event_queue = []
            except Exception as exc:
                print(
                    "impossible to contact GO Feature Flag relay proxy instance to collect the data: {}".format(
                        exc
                    )
                )
                return
