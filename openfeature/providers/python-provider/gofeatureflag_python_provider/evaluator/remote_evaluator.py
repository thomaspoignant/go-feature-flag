"""Remote evaluator that uses the OFREP provider for remote evaluation."""

import logging
from typing import Dict, Mapping, Optional, Sequence, Union

from openfeature.contrib.provider.ofrep import OFREPProvider
from openfeature.provider import EvaluationContext, FlagResolutionDetails, FlagValueType

from gofeatureflag_python_provider.evaluator.evaluator import IEvaluator
from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions


class RemoteEvaluator(IEvaluator):
    """Remote evaluator that uses the OFREP provider for remote evaluation.

    This evaluator delegates flag evaluation to a remote GO Feature Flag relay-proxy
    using the OFREP (OpenFeature Remote Evaluation Protocol) standard.

    Attributes:
        ofrep_provider: The underlying OFREP provider instance.
        logger: Logger instance for this evaluator.
    """

    ofrep_provider: OFREPProvider
    logger: Optional[logging.Logger]

    def __init__(
        self, options: GoFeatureFlagOptions, logger: Optional[logging.Logger] = None
    ) -> None:
        """Initialize the remote evaluator.

        Args:
            options: Configuration options for the GO Feature Flag provider.
            logger: Optional logger instance. If None, a default logger will be used.

        Raises:
            ValueError: If required options are missing or invalid.
        """
        self.logger = logger or logging.getLogger(__name__)

        # Create headers factory - always include Content-Type, add auth if API key provided
        def _create_headers() -> Dict[str, str]:
            headers = {"Content-Type": "application/json"}
            if options.api_key:
                headers["Authorization"] = f"Bearer {options.api_key}"
            return headers

        # Extract timeout from pool manager if available
        timeout: Optional[float] = None
        if options.urllib3_pool_manager and hasattr(
            options.urllib3_pool_manager, "timeout"
        ):
            timeout = getattr(
                options.urllib3_pool_manager.timeout, "connect_timeout", lambda: None
            )()

        self.ofrep_provider = OFREPProvider(
            base_url=str(options.endpoint),
            timeout=timeout,
            headers_factory=_create_headers,
        )

    def initialize(self, evaluation_context: EvaluationContext) -> None:
        """Initialize the remote evaluator.

        Args:
            evaluation_context: The evaluation context to initialize with.
        """
        self.logger.debug("Initializing remote evaluator")
        self.ofrep_provider.initialize(evaluation_context=evaluation_context)
        self.logger.info("Remote evaluator initialized successfully")

    def shutdown(self) -> None:
        """Shutdown the remote evaluator and clean up resources."""
        self.ofrep_provider.shutdown()
        self.logger.info("Remote evaluator shutdown successfully")

    def is_flag_trackable(self, flag_key: str) -> bool:
        """Check if the flag is trackable.

        For remote evaluation, we delegate this to the OFREP provider.

        Args:
            flag_key: The key of the flag to check.

        Returns:
            True if the flag is trackable, False otherwise.
        """
        # OFREP provider doesn't expose this method directly, so we return True
        # as remote flags are generally trackable
        return True

    def evaluate_boolean(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Evaluate a boolean flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return self.ofrep_provider.resolve_boolean_details(
            flag_key, default_value, evaluation_context
        )

    async def evaluate_boolean_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Evaluate a boolean flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return await self.ofrep_provider.resolve_boolean_details_async(
            flag_key, default_value, evaluation_context
        )

    def evaluate_string(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Evaluate a string flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return self.ofrep_provider.resolve_string_details(
            flag_key, default_value, evaluation_context
        )

    async def evaluate_string_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Evaluate a string flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return await self.ofrep_provider.resolve_string_details_async(
            flag_key, default_value, evaluation_context
        )

    def evaluate_integer(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Evaluate an integer flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return self.ofrep_provider.resolve_integer_details(
            flag_key, default_value, evaluation_context
        )

    async def evaluate_integer_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Evaluate an integer flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return await self.ofrep_provider.resolve_integer_details_async(
            flag_key, default_value, evaluation_context
        )

    def evaluate_float(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Evaluate a float flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return self.ofrep_provider.resolve_float_details(
            flag_key, default_value, evaluation_context
        )

    async def evaluate_float_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Evaluate a float flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return await self.ofrep_provider.resolve_float_details_async(
            flag_key, default_value, evaluation_context
        )

    async def evaluate_object_async(
        self,
        flag_key: str,
        default_value: Union[Sequence[FlagValueType], Mapping[str, FlagValueType]],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[
        Union[Sequence[FlagValueType], Mapping[str, FlagValueType]]
    ]:
        """Evaluate an object flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if evaluation fails.
            evaluation_context: Optional context for flag evaluation.

        Returns:
            The resolution details of the flag evaluation.
        """
        return await self.ofrep_provider.resolve_object_details_async(
            flag_key, default_value, evaluation_context
        )
