import asyncio
import logging
from datetime import datetime
from typing import Any, Dict, Optional
from openfeature.evaluation_context import EvaluationContext
from openfeature.provider import FlagResolutionDetails
from openfeature.exception import (
    FlagNotFoundError,
    ParseError,
    TypeMismatchError,
    TargetingKeyMissingError,
    InvalidContextError,
    ProviderNotReadyError,
    ProviderFatalError,
    GeneralError,
)
from .evaluator import IEvaluator
from ..provider_options import GoFeatureFlagOptions
from ..model import (
    EvaluationResponse,
    Flag,
    WasmInput,
    FlagContext,
    FlagConfigResponse,
)
from ..wasm import EvaluateWasm
from ..exception import ImpossibleToRetrieveConfigurationException


class InProcessEvaluator(IEvaluator):
    """
    InProcessEvaluator is an implementation of the IEvaluator interface that evaluates feature flags in-process.
    It uses the WASM evaluation engine to perform flag evaluations locally.
    """

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        api,  # Will be replaced with proper API type
        event_channel=None,  # Will be replaced with proper event channel type
        logger=None,
    ):
        """
        Constructor of the InProcessEvaluator.

        Args:
            options: Options to configure the provider
            api: API to contact GO Feature Flag
            event_channel: Event channel to send events to the event bus or event handler
            logger: Logger instance
        """
        self.api = api
        self.options = options
        self.event_channel = event_channel
        self.logger = logger or logging.getLogger(__name__)
        self.evaluation_engine = EvaluateWasm(logger)

        # Configuration state
        self.etag = None
        self.last_update = datetime(1970, 1, 1)
        self.flags: Dict[str, Flag] = {}
        self.evaluation_context_enrichment: Dict[str, Any] = {}
        self.periodic_runner = None

    async def initialize(self) -> None:
        """
        Initialize the evaluator.
        """
        await self.evaluation_engine.initialize()
        await self._load_configuration(True)

        # Start periodic configuration polling
        if (
            self.options.flag_change_polling_interval_ms
            and self.options.flag_change_polling_interval_ms > 0
        ):
            self.periodic_runner = asyncio.create_task(self._poll())

    async def _poll(self) -> None:
        """
        Poll the configuration from the API.
        """
        while True:
            try:
                await asyncio.sleep(self.options.flag_change_polling_interval_ms / 1000)
                await self._load_configuration(False)
            except Exception as error:
                self.logger.error(f"Failed to load configuration: {error}")

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
        response = await self._generic_evaluate(
            flag_key, default_value, evaluation_context
        )
        self._handle_error(response, flag_key)

        if isinstance(response.value, bool):
            return self._prepare_response(response, flag_key, response.value)

        raise TypeMismatchError(
            f"Flag {flag_key} had unexpected type, expected boolean."
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
        response = await self._generic_evaluate(
            flag_key, default_value, evaluation_context
        )
        self._handle_error(response, flag_key)

        if isinstance(response.value, str):
            return self._prepare_response(response, flag_key, response.value)

        raise TypeMismatchError(
            f"Flag {flag_key} had unexpected type, expected string."
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
        response = await self._generic_evaluate(
            flag_key, default_value, evaluation_context
        )
        self._handle_error(response, flag_key)

        if isinstance(response.value, (int, float)):
            return self._prepare_response(response, flag_key, float(response.value))

        raise TypeMismatchError(
            f"Flag {flag_key} had unexpected type, expected number."
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
        response = await self._generic_evaluate(
            flag_key, default_value, evaluation_context
        )
        self._handle_error(response, flag_key)

        if response.value is not None:
            return self._prepare_response(response, flag_key, response.value)

        raise TypeMismatchError(
            f"Flag {flag_key} had unexpected type, expected object."
        )

    def is_flag_trackable(self, flag_key: str) -> bool:
        """
        Check if the flag is trackable.

        Args:
            flag_key: The key of the flag to check.

        Returns:
            True if the flag is trackable.
        """
        flag = self.flags.get(flag_key)
        if not flag:
            self.logger.warning(f"Flag with key {flag_key} not found")
            # If the flag is not found, this is most likely a configuration change, so we track it by default.
            return True

        return getattr(flag, "track_events", True)

    async def dispose(self) -> None:
        """
        Dispose the evaluator.
        """
        if self.periodic_runner:
            self.periodic_runner.cancel()
            self.periodic_runner = None

        await self.evaluation_engine.dispose()

    async def _generic_evaluate(
        self,
        flag_key: str,
        default_value: Any,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> EvaluationResponse:
        """
        Evaluates a flag with the given key and default value in the context of the provided evaluation context.

        Args:
            flag_key: Name of the feature flag
            default_value: Default value in case of error
            evaluation_context: Context of the evaluation

        Returns:
            An EvaluationResponse containing the output of the evaluation.
        """
        flag = self.flags.get(flag_key)
        if not flag:
            return EvaluationResponse(
                variation_type="default",
                track_events=True,
                reason="ERROR",
                error_code="FLAG_NOT_FOUND",
                error_details=f"Flag with key '{flag_key}' not found",
                value=default_value,
                metadata={},
            )

        # Convert evaluation context to dict
        context_dict = {}
        if evaluation_context:
            # This is a placeholder - need to implement proper context conversion
            context_dict = {}

        input_data = WasmInput(
            flag_key=flag_key,
            flag=flag,
            eval_context=context_dict,
            flag_context=FlagContext(
                default_sdk_value=default_value,
                evaluation_context_enrichment=self.evaluation_context_enrichment,
            ),
        )

        return await self.evaluation_engine.evaluate(input_data)

    async def _load_configuration(self, first_load: bool = False) -> None:
        """
        LoadConfiguration is responsible for loading the configuration of the flags from the API.

        Args:
            first_load: Whether this is the first load

        Raises:
            ImpossibleToRetrieveConfigurationException: In case we are not able to call the relay proxy and to get the flag values.
        """
        try:
            # Call the API to retrieve the flags' configuration and store it in the local copy
            # This is a placeholder - need to implement actual API call
            flag_config_response = (
                None  # await self.api.retrieve_flag_configuration(self.etag, None)
            )

            if not flag_config_response:
                raise ImpossibleToRetrieveConfigurationException(
                    "Flag configuration response is null"
                )

            # TODO: Implement configuration loading logic

        except Exception as error:
            self.logger.error(f"Failed to load configuration: {error}")
            raise

    def _handle_error(self, response: EvaluationResponse, flag_key: str) -> None:
        """
        HandleError is handling the error response from the evaluation API.

        Args:
            response: Response of the evaluation.
            flag_key: Name of the feature flag.

        Raises:
            Error: When the evaluation is on error.
        """
        error_code = response.error_code
        if not error_code:
            # if we no error code it means that the evaluation is successful
            return

        error_details = response.error_details or f"Error for flag {flag_key}"

        if error_code == "FLAG_NOT_FOUND":
            raise FlagNotFoundError(error_details)
        elif error_code == "PARSE_ERROR":
            raise ParseError(error_details)
        elif error_code == "TYPE_MISMATCH":
            raise TypeMismatchError(error_details)
        elif error_code == "TARGETING_KEY_MISSING":
            raise TargetingKeyMissingError(error_details)
        elif error_code == "INVALID_CONTEXT":
            raise InvalidContextError(error_details)
        elif error_code == "PROVIDER_NOT_READY":
            raise ProviderNotReadyError(error_details)
        elif error_code == "PROVIDER_FATAL":
            raise ProviderFatalError(error_details)
        else:
            raise GeneralError(error_details)

    def _prepare_response(
        self, response: EvaluationResponse, flag_key: str, value: Any
    ) -> FlagResolutionDetails[Any]:
        """
        PrepareResponse is preparing the response to be returned to the caller.

        Args:
            response: Response of the evaluation.
            flag_key: Name of the feature flag.
            value: Value of the feature flag.

        Returns:
            FlagResolutionDetails with the flag value and metadata.
        """
        try:
            # This is a placeholder - need to implement proper response preparation
            # TODO: Implement proper FlagResolutionDetails creation
            return FlagResolutionDetails(
                value=value,
                reason=response.reason,
                flag_metadata=response.metadata or {},
                variant=response.variation_type,
            )
        except Exception as error:
            raise TypeMismatchError(
                f"Flag value {flag_key} had unexpected type {type(response.value)}."
            )
