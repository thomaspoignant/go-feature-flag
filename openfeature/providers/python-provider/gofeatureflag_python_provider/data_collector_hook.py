import datetime
import threading
import time
from urllib.parse import urljoin

import urllib3
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason
from openfeature.hook import Hook, HookContext

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import FeatureEvent, RequestDataCollector

default_targeting_key = 'undefined-targetingKey'


class DataCollectorHook(Hook):
    # _thread_stopper is used to stop the background task when we shutdown the hook
    _thread_stopper: bool = True
    # _thread_data_collector is the thread used to call the relay proxy to collect data
    _thread_data_collector: threading.Thread = None
    # _options is the options of the provider
    _options: GoFeatureFlagOptions = None
    # _data_collector_endpoint is the endpoint of the relay proxy
    _data_collector_endpoint: str = None
    # _http_client is the http client used to call the relay proxy
    _http_client: urllib3.PoolManager = None
    # _data_event_queue is the list of data to collect
    _event_queue: list[FeatureEvent] = []

    def __init__(self, options: GoFeatureFlagOptions, http_client: urllib3.PoolManager):
        self._http_client = http_client
        self._thread_data_collector = threading.Thread(target=self.background_task)
        self._options = options
        self._data_collector_endpoint = urljoin(str(self._options.endpoint), "/v1/data/collector")

    def after(self, hook_context: HookContext, details: FlagEvaluationDetails, hints: dict):
        if self._options.disable_data_collection or details.reason == Reason.CACHED:
            # we don't collect data if the data collection is disabled or if the flag is not cached
            return

        feature_event = FeatureEvent(
            contextKind='anonymousUser' if hook_context.evaluation_context.attributes['anonymous'] else 'user',
            creationDate=int(datetime.datetime.now().timestamp()),
            default=False,
            key=hook_context.flag_key,
            value=details.value,
            variation=details.variant or 'SdkDefault',
            userKey=hook_context.evaluation_context.targeting_key or default_targeting_key,
        )
        self._event_queue.append(feature_event)

    def error(self, hook_context: HookContext, exception: Exception, hints: dict):
        if self._options.disable_data_collection or details.reason == Reason.CACHED:
            # we don't collect data if the data collection is disabled or if the flag is not cached
            return

        feature_event = FeatureEvent(
            contextKind='anonymousUser' if hook_context.evaluation_context.attributes['anonymous'] else 'user',
            creationDate=int(datetime.datetime.now().timestamp()),
            default=True,
            key=hook_context.flag_key,
            value=hook_context.default_value,
            variation='SdkDefault',
            userKey=hook_context.evaluation_context.targeting_key or default_targeting_key,
        )
        self._event_queue.append(feature_event)
        pass

    def initialize(self):
        self._thread_stopper = False
        self._thread_data_collector.start()

    def shutdown(self):
        print("shutdown")
        # setting the _thread_stopper to True will stop the background task
        time.sleep(1)
        self._thread_stopper = True
        self._thread_data_collector.join()

    def background_task(self):
        while not self._thread_stopper:
            time.sleep(self._options.data_flush_interval / 1000)
            if len(self._event_queue) == 0:
                continue

            try:
                goff_request = RequestDataCollector(
                    meta={'provider': 'open-feature-python-sdk'},
                    events=self._event_queue,
                )
                response = self._http_client.request(
                    method="POST",
                    url=urljoin(
                        str(self._options.endpoint),
                        "/v1/data/collector"
                    ),
                    headers={"Content-Type": "application/json"},
                    body=goff_request.model_dump_json(),
                )
                print(response.status)
            except Exception as exc:
                print(exc)
                continue
        # TODO: check how to deal with mock
