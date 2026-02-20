import datetime
import logging
from gofeatureflag_python_provider.evaluator import AbstractEvaluator
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import FeatureEvent
from gofeatureflag_python_provider.services.event_publisher import EventPublisher
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason
from openfeature.hook import Hook, HookContext

logger = logging.getLogger(__name__)
default_targeting_key = "undefined-targetingKey"


class DataCollectorHook(Hook):
    _options: GoFeatureFlagOptions
    _event_publisher: EventPublisher
    _evaluator: AbstractEvaluator

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        event_publisher: EventPublisher,
        evaluator: AbstractEvaluator,
    ):
        if event_publisher is None:
            raise ValueError("event_publisher cannot be None")
        if evaluator is None:
            raise ValueError("evaluator cannot be None")
        self._options = options
        self._event_publisher = event_publisher
        self._evaluator = evaluator

    def after(
        self, hook_context: HookContext, details: FlagEvaluationDetails, hints: dict
    ):
        if (
            self._options.disable_data_collection
            or not self._evaluator.is_flag_trackable(hook_context.flag_key)
        ):
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
        self._event_publisher.add_event(feature_event)

    def error(self, hook_context: HookContext, exception: Exception, hints: dict):
        if (
            self._options.disable_data_collection
            or not self._evaluator.is_flag_trackable(hook_context.flag_key)
        ):
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
        self._event_publisher.add_event(feature_event)

    def initialize(self):
        self._event_publisher.start()

    def shutdown(self):
        self._event_publisher.stop()
