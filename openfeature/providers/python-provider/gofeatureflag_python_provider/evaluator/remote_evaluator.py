from typing import Any, Optional
from openfeature.evaluation_context import EvaluationContext
from openfeature.provider import FlagResolutionDetails
from openfeature.contrib.provider.ofrep import OFREPProvider
from .evaluator import IEvaluator
from ..provider_options import GoFeatureFlagOptions


class RemoteEvaluator(IEvaluator):
    """
    Remote evaluator that uses the OFREP provider for remote evaluation.
    """

    def __init__(self, options: GoFeatureFlagOptions, logger=None):
        """
        Initialize the remote evaluator.

        Args:
            options: Provider options
            logger: Logger instance
        """
        self.logger = logger

        # Create OFREP provider options
        ofrep_options = {
            "base_url": options.endpoint,
            "timeout": (options.timeout or 10000) / 1000.0,  # Convert ms to seconds
        }

        # Add headers factory if API key is provided
        if options.api_key:

            def headers_factory():
                return {
                    "Content-Type": "application/json",
                    "Authorization": f"Bearer {options.api_key}",
                }

            ofrep_options["headers_factory"] = headers_factory

        self.ofrep_provider = OFREPProvider(**ofrep_options)

    async def evaluate_boolean(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """
        Evaluates a boolean flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """
        context = evaluation_context or EvaluationContext()
        return await self.ofrep_provider.resolve_boolean_evaluation(
            flag_key, default_value, context
        )

    async def evaluate_string(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """
        Evaluates a string flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """
        context = evaluation_context or EvaluationContext()
        return await self.ofrep_provider.resolve_string_evaluation(
            flag_key, default_value, context
        )

    async def evaluate_number(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Evaluates a number flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """
        context = evaluation_context or EvaluationContext()
        return await self.ofrep_provider.resolve_number_evaluation(
            flag_key, default_value, context
        )

    async def evaluate_object(
        self,
        flag_key: str,
        default_value: Any,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Any]:
        """
        Evaluates an object flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """
        context = evaluation_context or EvaluationContext()
        return await self.ofrep_provider.resolve_object_evaluation(
            flag_key, default_value, context
        )

    def is_flag_trackable(self, flag_key: str) -> bool:
        """
        Checks if the flag is trackable.

        Args:
            flag_key: The key of the flag to check.

        Returns:
            False for remote evaluation.
        """
        return False

    async def dispose(self) -> None:
        """
        Disposes the evaluator.
        """
        if self.logger:
            self.logger.info("Disposing Remote evaluator")

    async def initialize(self) -> None:
        """
        Initializes the evaluator.
        """
        if self.logger:
            self.logger.info("Initializing Remote evaluator")
