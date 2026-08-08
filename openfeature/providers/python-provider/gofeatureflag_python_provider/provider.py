"""
OpenFeature provider for GO Feature Flag.

This module provides GoFeatureFlagProvider, which implements the
OpenFeature provider interface for evaluating feature flags.
Delegates to a RemoteEvaluator or InProcessEvaluator based on options.evaluation_type.
"""

import logging

from gofeatureflag_python_provider.evaluator import (
    AbstractEvaluator,
    InProcessEvaluator,
    RemoteEvaluator,
)
from gofeatureflag_python_provider.hooks import (
    DataCollectorHook,
    EnrichEvaluationContextHook,
)
from gofeatureflag_python_provider.metadata import GoFeatureFlagMetadata
from gofeatureflag_python_provider.services.api import GoFeatureFlagApi
from gofeatureflag_python_provider.services.event_publisher import EventPublisher
from gofeatureflag_python_provider.options import (
    BaseModel,
    EvaluationType,
    GoFeatureFlagOptions,
)
from openfeature.evaluation_context import EvaluationContext
from openfeature.event import ProviderEventDetails
from openfeature.flag_evaluation import FlagResolutionDetails
from openfeature.hook import Hook
from openfeature.provider import AbstractProvider
from openfeature.provider.metadata import Metadata
from pydantic import PrivateAttr
from typing import List, Optional, Union

AbstractProviderMetaclass = type(AbstractProvider)
BaseModelMetaclass = type(BaseModel)


class CombinedMetaclass(AbstractProviderMetaclass, BaseModelMetaclass):
    """
    Metaclass combining AbstractProvider and Pydantic BaseModel so the provider
    can use both inheritance and Pydantic configuration.
    """

    pass


class GoFeatureFlagProvider(BaseModel, AbstractProvider, metaclass=CombinedMetaclass):
    """
    OpenFeature provider for GO Feature Flag.
    """

    options: GoFeatureFlagOptions

    def __hash__(self) -> int:
        """Make provider hashable for use as dict key in OpenFeature registry."""
        return id(self)

    def __eq__(self, other: object) -> bool:
        """Identity equality so __hash__ contract is satisfied."""
        return other is self

    _evaluator: AbstractEvaluator = PrivateAttr()
    _data_collector_hook: DataCollectorHook = PrivateAttr()
    _event_publisher: EventPublisher = PrivateAttr()
    _hooks: List[Hook] = PrivateAttr(default_factory=list)

    def __init__(self, **data):
        """
        Constructor of the provider. Passes through to Pydantic configuration.
        Selects RemoteEvaluator or InProcessEvaluator based on options.evaluation_type.

        :param data: data coming from pydantic configuration
        """
        super().__init__(**data)
        # Deliberately the package-root logger, not __name__: every module logs
        # through logging.getLogger(__name__), so setting the level here reaches
        # all of them by inheritance. Applied only when nothing has configured
        # this logger yet — the level is process-wide, so an unconditional
        # setLevel would overwrite what the embedding application asked for.
        package_logger = logging.getLogger("gofeatureflag_python_provider")
        if package_logger.level == logging.NOTSET:
            package_logger.setLevel(self.options.get_log_level_int())
        api = GoFeatureFlagApi(self.options)
        self._event_publisher = EventPublisher(api=api, options=self.options)
        self._evaluator = self._create_evaluator(api)
        self._hooks = []

        # Order matters: enrichment runs before the data collector so the
        # collector observes the enriched context. The enrichment hook is
        # registered unconditionally — exporter metadata always carries the
        # reserved keys identifying this SDK, so it always has something to
        # contribute even when the user configured no metadata of their own.
        self._hooks.append(
            EnrichEvaluationContextHook(
                metadata=self.options.get_exporter_metadata(),
            )
        )
        self._hooks.append(
            DataCollectorHook(
                options=self.options,
                event_publisher=self._event_publisher,
                evaluator=self._evaluator,
            )
        )

    def _create_evaluator(self, api: GoFeatureFlagApi) -> AbstractEvaluator:
        """Create the evaluator configured for this provider."""
        if self.options.evaluation_type == EvaluationType.REMOTE:
            return RemoteEvaluator(self.options)

        return InProcessEvaluator(
            self.options,
            api,
            on_configuration_changed=self._emit_configuration_changed,
            on_stale=self._emit_stale,
            on_ready=self._emit_ready,
        )

    def initialize(
        self, evaluation_context: Optional[EvaluationContext] = None
    ) -> None:
        """Initialize the provider and its evaluator."""
        self._event_publisher.start()
        self._evaluator.initialize(evaluation_context)

    def shutdown(self) -> None:
        """Shut down the provider and release evaluator resources."""
        self._evaluator.shutdown()
        self._event_publisher.stop()

    # ------------------------------------------------------------------
    # Provider events
    #
    # Emitted from the in-process polling loop. The SDK stamps the provider
    # name onto every event from the registry, so it always matches the
    # metadata name. Emitting before the SDK attaches its listener is a no-op.
    # ------------------------------------------------------------------

    def _emit_configuration_changed(self, changed_flags: List[str]) -> None:
        """A poll produced a configuration different from the one in use."""
        self.emit_provider_configuration_changed(
            ProviderEventDetails(
                flags_changed=changed_flags,
                message="flag configuration changed",
            )
        )

    def _emit_stale(self, message: str) -> None:
        """Consecutive refreshes failed; the last known-good config still serves."""
        self.emit_provider_stale(ProviderEventDetails(message=message))

    def _emit_ready(self) -> None:
        """A refresh succeeded after the provider had gone stale."""
        self.emit_provider_ready(
            ProviderEventDetails(message="flag configuration refresh recovered")
        )

    def get_metadata(self) -> Metadata:
        """
        Return the provider metadata (name, version).

        :return: Metadata for this provider (GoFeatureFlagMetadata).
        """
        return GoFeatureFlagMetadata()

    def get_provider_hooks(self) -> List[Hook]:
        """
        Return the list of provider-level hooks.
        Hooks are managed by the provider only; evaluators do not provide hooks.

        :return: List of hooks (may be empty).
        """
        return self._hooks

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """
        Resolve the flag as a boolean.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for boolean.
        """
        return self._evaluator.resolve_boolean_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """
        Resolve the flag as a string.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for string.
        """
        return self._evaluator.resolve_string_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """
        Resolve the flag as an integer.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for integer.
        """
        return self._evaluator.resolve_integer_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Resolve the flag as a float.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for float.
        """
        return self._evaluator.resolve_float_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: dict,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[list, dict]]:
        """
        Resolve the flag as an object.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for object.
        """
        return self._evaluator.resolve_object_details(
            flag_key, default_value, evaluation_context
        )

    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """
        Asynchronously resolve the flag as a boolean.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for boolean.
        """
        return await self._evaluator.resolve_boolean_details_async(
            flag_key, default_value, evaluation_context
        )

    async def resolve_string_details_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """
        Asynchronously resolve the flag as a string.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for string.
        """
        return await self._evaluator.resolve_string_details_async(
            flag_key, default_value, evaluation_context
        )

    async def resolve_integer_details_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """
        Asynchronously resolve the flag as an integer.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for integer.
        """
        return await self._evaluator.resolve_integer_details_async(
            flag_key, default_value, evaluation_context
        )

    async def resolve_float_details_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Asynchronously resolve the flag as a float.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for float.
        """
        return await self._evaluator.resolve_float_details_async(
            flag_key, default_value, evaluation_context
        )

    async def resolve_object_details_async(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[dict, list]]:
        """
        Asynchronously resolve the flag as an object.

        :param flag_key: Flag key to evaluate.
        :param default_value: Default value if the flag cannot be evaluated.
        :param evaluation_context: Optional evaluation context (e.g. user/key and attributes).
        :return: Flag resolution details for object.
        """
        return await self._evaluator.resolve_object_details_async(
            flag_key, default_value, evaluation_context
        )
