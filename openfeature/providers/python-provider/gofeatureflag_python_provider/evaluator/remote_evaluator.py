"""
Remote evaluator: evaluates flags via the OpenFeature Remote Evaluation Protocol (OFREP).
Delegates to openfeature-provider-ofrep, which talks to OFREP-compliant backends
(e.g. GO Feature Flag relay proxy at /ofrep/v1).
"""

import asyncio
from typing import Optional, Union

from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagResolutionDetails

from gofeatureflag_python_provider.evaluator.abstract_evaluator import AbstractEvaluator
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.ofrep import build_ofrep_provider


class RemoteEvaluator(AbstractEvaluator):
    """Evaluates flags by delegating to OFREPProvider (OFREP protocol)."""

    def __init__(self, options: GoFeatureFlagOptions) -> None:
        """Create a remote evaluator that uses OFREP to evaluate flags.

        :param options: Provider options (endpoint, optional api_key, timeout).
            The OFREP provider is configured with options.endpoint as base URL
            and an X-API-Key header when options.api_key is set.
        """
        self._options = options
        self._ofrep_provider = build_ofrep_provider(options)

    def initialize(
        self, evaluation_context: Optional[EvaluationContext] = None
    ) -> None:
        """Initialize the evaluator. No-op (OFREPProvider has no initialize)."""
        pass

    def shutdown(self) -> None:
        """Release resources. No-op (OFREPProvider has no shutdown)."""
        pass

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Resolve the flag as a boolean via OFREP."""
        return self._ofrep_provider.resolve_boolean_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Resolve the flag as a string via OFREP."""
        return self._ofrep_provider.resolve_string_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Resolve the flag as an integer via OFREP."""
        return self._ofrep_provider.resolve_integer_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Resolve the flag as a float via OFREP."""
        return self._ofrep_provider.resolve_float_details(
            flag_key, default_value, evaluation_context
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[list, dict]]:
        """Resolve the flag as an object (dict or list) via OFREP."""
        return self._ofrep_provider.resolve_object_details(
            flag_key, default_value, evaluation_context
        )

    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Resolve the flag as a boolean via OFREP (async, runs sync call in thread)."""
        return await asyncio.to_thread(
            self._ofrep_provider.resolve_boolean_details,
            flag_key,
            default_value,
            evaluation_context,
        )

    async def resolve_string_details_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Resolve the flag as a string via OFREP (async, runs sync call in thread)."""
        return await asyncio.to_thread(
            self._ofrep_provider.resolve_string_details,
            flag_key,
            default_value,
            evaluation_context,
        )

    async def resolve_integer_details_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Resolve the flag as an integer via OFREP (async, runs sync call in thread)."""
        return await asyncio.to_thread(
            self._ofrep_provider.resolve_integer_details,
            flag_key,
            default_value,
            evaluation_context,
        )

    async def resolve_float_details_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Resolve the flag as a float via OFREP (async, runs sync call in thread)."""
        return await asyncio.to_thread(
            self._ofrep_provider.resolve_float_details,
            flag_key,
            default_value,
            evaluation_context,
        )

    async def resolve_object_details_async(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[dict, list]]:
        """Resolve the flag as an object via OFREP (async, runs sync call in thread)."""
        return await asyncio.to_thread(
            self._ofrep_provider.resolve_object_details,
            flag_key,
            default_value,
            evaluation_context,
        )

    def is_flag_trackable(self, flag_key: str) -> bool:
        return False
