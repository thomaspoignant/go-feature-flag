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
        logging.getLogger("gofeatureflag_python_provider").setLevel(
            self.options.get_log_level_int()
        )
        api = GoFeatureFlagApi(self.options)
        self._event_publisher = EventPublisher(api=api, options=self.options)

        if self.options.evaluation_type == EvaluationType.REMOTE:
            self._evaluator = RemoteEvaluator(self.options)
        else:
            self._evaluator = InProcessEvaluator(self.options, api)

        # create the data collector hook if data collection is not disabled
        self._hooks.append(
            DataCollectorHook(
                options=self.options,
                event_publisher=self._event_publisher,
                evaluator=self._evaluator,
            )
        )

        # create the enrichment hook if exporter_metadata is not empty
        if len(self.options.exporter_metadata) > 0:
            self._hooks.append(
                EnrichEvaluationContextHook(
                    metadata=self.options.exporter_metadata,
                )
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
