import datetime
import logging
from typing import Optional

from gofeatureflag_python_provider.evaluator import AbstractEvaluator
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model.request_data_collector import (
    FeatureEvent,
)
from gofeatureflag_python_provider.services.event_publisher import EventPublisher
from gofeatureflag_python_provider.utils import context_kind
from openfeature.flag_evaluation import FlagEvaluationDetails, Reason
from openfeature.hook import Hook, HookContext

default_targeting_key = "undefined-targetingKey"

# Flag metadata marking a result the relay proxy produced via remote fallback.
EVALUATED_REMOTELY_KEY = "gofeatureflag_evaluated_remotely"

# Flag metadata key carrying the flag version reported by the engine.
VERSION_KEY = "version"

# Every event this hook emits is a local evaluation: remote mode produces no
# trackable flags, and fallback results are skipped above.
SOURCE_INPROCESS = "INPROCESS"

logger = logging.getLogger(__name__)


def _version_from(flag_metadata: Optional[dict]) -> Optional[str]:
    """Read the flag version out of flag metadata as a string.

    Flag metadata is authored by whoever wrote the flag, so `version` can hold
    any JSON type. FeatureEvent types it as a string, and an unconvertible value
    would raise inside this hook — which the SDK turns into an error result, so
    a perfectly good evaluation would start returning the default value.
    """
    version = (flag_metadata or {}).get(VERSION_KEY)
    if version is None or isinstance(version, str):
        return version
    return str(version)


class DataCollectorHook(Hook):
    """A hook that collects events for flag evaluations."""

    _options: GoFeatureFlagOptions
    _event_publisher: EventPublisher
    _evaluator: AbstractEvaluator

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        event_publisher: EventPublisher,
        evaluator: AbstractEvaluator,
    ):
        """Initialize the data collector hook."""
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
        """Collect an event for a flag evaluation."""
        if (
            self._options.disable_data_collection
            or not self._evaluator.is_flag_trackable(hook_context.flag_key)
        ):
            logger.debug(
                "Skipping event for flag %s because data collection is disabled or the flag is not trackable",
                hook_context.flag_key,
            )
            return

        if (details.flag_metadata or {}).get(EVALUATED_REMOTELY_KEY):
            # The relay proxy evaluated this one and has already recorded it.
            # Emitting here would count the same evaluation twice.
            logger.info(
                "Skipping event for flag %s because it was evaluated remotely",
                hook_context.flag_key,
            )
            return

        feature_event = FeatureEvent(
            contextKind=context_kind(hook_context.evaluation_context),
            creationDate=int(datetime.datetime.now().timestamp()),
            default=False,
            key=hook_context.flag_key,
            value=details.value,
            variation=details.variant or "SdkDefault",
            version=_version_from(details.flag_metadata),
            source=SOURCE_INPROCESS,
            userKey=hook_context.evaluation_context.targeting_key
            or default_targeting_key,
        )
        self._event_publisher.add_event(feature_event)

    def error(self, hook_context: HookContext, exception: Exception, hints: dict):
        """Collect an error event for a flag evaluation."""
        if (
            self._options.disable_data_collection
            or not self._evaluator.is_flag_trackable(hook_context.flag_key)
        ):
            logger.debug(
                "Skipping error event for flag %s because data collection is disabled or the flag is not trackable",
                hook_context.flag_key,
            )
            return

        feature_event = FeatureEvent(
            contextKind=context_kind(hook_context.evaluation_context),
            creationDate=int(datetime.datetime.now().timestamp()),
            default=True,
            key=hook_context.flag_key,
            value=hook_context.default_value,
            variation="SdkDefault",
            source=SOURCE_INPROCESS,
            userKey=hook_context.evaluation_context.targeting_key
            or default_targeting_key,
        )
        self._event_publisher.add_event(feature_event)
