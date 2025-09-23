"""
GO Feature Flag provider for OpenFeature.
"""

from ctypes import Union
import logging
from typing import Mapping, Optional, Sequence
from openfeature.provider import (
    AbstractProvider,
    EvaluationContext,
    FlagResolutionDetails,
    FlagValueType,
    Hook,
    Metadata,
)

from gofeatureflag_python_provider.evaluator.evaluator import IEvaluator
from gofeatureflag_python_provider.evaluator.remote_evaluator import RemoteEvaluator
from gofeatureflag_python_provider.exception import InvalidOptionsException
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions


class GoFeatureFlagProviderOld(AbstractProvider):
    """
    GO Feature Flag provider for OpenFeature.

    Attributes:
    ofrep_provider: The underlying OFREP provider instance.
    logger: Logger instance for this evaluator.
    evaluator: Evaluator instance
    hooks: List of provider hooks
    """

    logger: Optional[logging.Logger]
    options: GoFeatureFlagOptions
    evaluator: IEvaluator
    hooks: list[Hook]

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        logger: Optional[logging.Logger] = None,
    ):
        """
        Initialize the GO Feature Flag provider.
        """
        super().__init__()
        if options is None:
            raise InvalidOptionsException("No options provided")
        self.options = options
        self.logger = logger or logging.getLogger(__name__)
        self.evaluator = self._get_evaluator()
        self.hooks: list[Hook] = []

    def get_metadata(self) -> Metadata:
        """
        Get the provider metadata.

        Returns:
            Provider metadata
        """
        return Metadata(name="GO Feature Flag")

    # Synchronous methods
    # --------------------
    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """
        Resolve a boolean flag synchronously.
        """
        return self.evaluator.evaluate_boolean(
            flag_key, default_value, evaluation_context
        )

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """
        Resolve a string flag synchronously.
        """
        return self.evaluator.evaluate_string(
            flag_key, default_value, evaluation_context
        )

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """
        Resolve an integer flag synchronously.
        """
        return self.evaluator.evaluate_integer(
            flag_key, default_value, evaluation_context
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Resolve a float flag synchronously.
        """
        return self.evaluator.evaluate_float(
            flag_key, default_value, evaluation_context
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Union[Sequence[FlagValueType], Mapping[str, FlagValueType]],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[
        Union[Sequence[FlagValueType], Mapping[str, FlagValueType]]
    ]:
        """
        Resolve an object flag synchronously.
        """
        return self.evaluator.evaluate_object(
            flag_key, default_value, evaluation_context
        )

    # --------------------

    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """
        Resolve a boolean flag asynchronously.
        """
        return self.evaluator.evaluate_boolean_async(
            flag_key, default_value, evaluation_context
        )

    def _get_evaluator(self) -> IEvaluator:
        """
        Get the evaluator based on the evaluation type specified in the options.
        """
        return RemoteEvaluator(self.options, self.logger)
