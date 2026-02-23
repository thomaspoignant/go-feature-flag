"""
In-process evaluator: evaluates flags locally via the WASM module.
Fetches flag configuration from the relay proxy, stores it locally, and polls for updates.
"""

import asyncio
import logging
import threading
from typing import Any, Optional, Type, TypeVar, Union

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import (
    FlagNotFoundError,
    GeneralError,
    TypeMismatchError,
)
from openfeature.flag_evaluation import FlagResolutionDetails, Reason

from gofeatureflag_python_provider.evaluator.abstract_evaluator import AbstractEvaluator
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.api import GoFeatureFlagApi
from gofeatureflag_python_provider.wasm import EvaluateWasm, WasmFlagContext, WasmInput

logger = logging.getLogger(__name__)

T = TypeVar("T")

_ERROR_CODE_FLAG_NOT_FOUND = "FLAG_NOT_FOUND"
_ERROR_CODE_TYPE_MISMATCH = "TYPE_MISMATCH"
_ERROR_CODE_TARGETING_KEY_MISSING = "TARGETING_KEY_MISSING"
_ERROR_CODE_INVALID_CONTEXT = "INVALID_CONTEXT"


class InProcessEvaluator(AbstractEvaluator):
    """Evaluates flags in-process via the WASM module. Fetches and polls flag configuration."""

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        api: GoFeatureFlagApi,
    ) -> None:
        self._options = options
        self._api = api
        self._flags: dict[str, Any] = {}
        self._etag: Optional[str] = None
        self._evaluation_context_enrichment: dict[str, Any] = {}
        self._lock = threading.Lock()
        self._poll_stopper: Optional[threading.Event] = None
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval_seconds: int = (
            options.flag_config_poll_interval_seconds or 10
        )
        self._wasm = EvaluateWasm(wasm_path=options.wasm_file_path)

    def initialize(
        self, evaluation_context: Optional[EvaluationContext] = None
    ) -> None:
        """Fetch initial flag configuration, initialize WASM, and start background polling."""
        self._wasm.initialize()

        response = self._api.retrieve_flag_configuration()
        with self._lock:
            self._flags = response.flags or {}
            self._etag = response.etag
            self._evaluation_context_enrichment = (
                response.evaluation_context_enrichment or {}
            )
        self._poll_stopper = threading.Event()
        self._poll_thread = threading.Thread(
            target=self._background_poll,
            daemon=True,
        )
        self._poll_thread.start()

    def _background_poll(self) -> None:
        """Loop: wait poll_interval, then refresh flag config until stopped."""
        while not self._poll_stopper.is_set():
            self._poll_stopper.wait(self._poll_interval_seconds)
            if self._poll_stopper.is_set():
                break
            self._refresh_flag_configuration()

    def _refresh_flag_configuration(self) -> None:
        """Call API with current etag; on 200 update stored flags, on 304 or error keep state."""
        with self._lock:
            etag = self._etag
        try:
            response = self._api.retrieve_flag_configuration(etag=etag)
        except Exception:
            logger.error(
                "Failed to refresh flag configuration",
                exc_info=True,
            )
            return
        with self._lock:
            if response.flags:
                self._flags = response.flags
                self._evaluation_context_enrichment = (
                    response.evaluation_context_enrichment or {}
                )
            if response.etag is not None:
                self._etag = response.etag

    def shutdown(self) -> None:
        """Stop the polling thread, dispose WASM, and release resources."""
        if self._poll_stopper is not None:
            self._poll_stopper.set()
        if self._poll_thread is not None:
            self._poll_thread.join(timeout=5.0)
        self._poll_thread = None
        self._poll_stopper = None
        self._wasm.dispose()
        with self._lock:
            self._flags = {}
            self._etag = None
            self._evaluation_context_enrichment = {}

    # ------------------------------------------------------------------
    # Generic resolver
    # ------------------------------------------------------------------

    @staticmethod
    def _build_eval_context(
        evaluation_context: Optional[EvaluationContext],
    ) -> dict[str, Any]:
        """Convert an OpenFeature EvaluationContext to a flat dict for the WASM input."""
        ctx: dict[str, Any] = {}
        if evaluation_context is None:
            return ctx
        if evaluation_context.targeting_key:
            ctx["targetingKey"] = evaluation_context.targeting_key
        if evaluation_context.attributes:
            ctx.update(evaluation_context.attributes)
        return ctx

    @staticmethod
    def _raise_for_error_code(
        flag_key: str, error_code: str, details: Optional[str]
    ) -> None:
        """Translate a WASM error code into the appropriate OpenFeature exception."""
        if error_code == _ERROR_CODE_FLAG_NOT_FOUND:
            raise FlagNotFoundError(details or f"Flag '{flag_key}' not found")
        if error_code == _ERROR_CODE_TYPE_MISMATCH:
            raise TypeMismatchError(details or f"Type mismatch for flag '{flag_key}'")
        raise GeneralError(
            details or f"Error evaluating flag '{flag_key}': {error_code}"
        )

    def _resolve_generic(
        self,
        flag_key: str,
        default_value: T,
        expected_type: Union[Type[T], tuple],
        evaluation_context: Optional[EvaluationContext],
    ) -> FlagResolutionDetails[T]:
        with self._lock:
            flag = self._flags.get(flag_key)
            enrichment = dict(self._evaluation_context_enrichment)

        if flag is None:
            raise FlagNotFoundError(
                f"Flag '{flag_key}' not found in local configuration"
            )

        wasm_input = WasmInput(
            flagKey=flag_key,
            flag=flag,
            evalContext=self._build_eval_context(evaluation_context),
            flagContext=WasmFlagContext(
                defaultSdkValue=default_value,
                evaluationContextEnrichment=enrichment,
            ),
        )

        response = self._wasm.evaluate(wasm_input)

        if response.errorCode:
            self._raise_for_error_code(
                flag_key, response.errorCode, response.errorDetails
            )

        value = response.value
        if value is not None and not isinstance(value, expected_type):
            raise TypeMismatchError(
                f"Flag '{flag_key}' returned type {type(value).__name__!r}, "
                f"expected {expected_type}"
            )

        resolved_value = value if value is not None else default_value
        return FlagResolutionDetails(
            value=resolved_value,
            reason=response.reason or Reason.DEFAULT,
            variant=response.variationType,
            flag_metadata=response.metadata or {},
        )

    # ------------------------------------------------------------------
    # Sync resolve methods
    # ------------------------------------------------------------------

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        return self._resolve_generic(flag_key, default_value, bool, evaluation_context)

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        return self._resolve_generic(flag_key, default_value, str, evaluation_context)

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        return self._resolve_generic(flag_key, default_value, int, evaluation_context)

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        return self._resolve_generic(
            flag_key, default_value, (float, int), evaluation_context
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[list, dict]]:
        return self._resolve_generic(
            flag_key, default_value, (dict, list), evaluation_context
        )

    # ------------------------------------------------------------------
    # Async resolve methods (delegate to sync counterparts)
    # ------------------------------------------------------------------

    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_boolean_details, flag_key, default_value, evaluation_context
        )

    async def resolve_string_details_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_string_details, flag_key, default_value, evaluation_context
        )

    async def resolve_integer_details_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_integer_details, flag_key, default_value, evaluation_context
        )

    async def resolve_float_details_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_float_details, flag_key, default_value, evaluation_context
        )

    async def resolve_object_details_async(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[dict, list]]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_object_details, flag_key, default_value, evaluation_context
        )

    # ------------------------------------------------------------------
    # Tracking
    # ------------------------------------------------------------------

    def is_flag_trackable(self, flag_key: str) -> bool:
        with self._lock:
            flag = self._flags.get(flag_key)
        if flag is None:
            logger.error("Flag with key %s not found", flag_key)
            return True
        track_events = flag.get("trackEvents", False)
        if isinstance(track_events, bool):
            return track_events
        return bool(track_events)
