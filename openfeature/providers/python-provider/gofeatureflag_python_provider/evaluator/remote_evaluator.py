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

try:
    from openfeature.contrib.provider.ofrep import OFREPProvider
except ImportError:
    OFREPProvider = None  # type: ignore[misc, assignment]


class RemoteEvaluator(AbstractEvaluator):
    """Evaluates flags by delegating to OFREPProvider (OFREP protocol)."""

    def __init__(self, options: GoFeatureFlagOptions) -> None:
        """Create a remote evaluator that uses OFREP to evaluate flags.

        :param options: Provider options (endpoint, optional api_key). The OFREP
            provider is configured with options.endpoint as base URL and Bearer
            auth when options.api_key is set.
        :raises ImportError: If openfeature-provider-ofrep is not installed.
        """
        if OFREPProvider is None:
            raise ImportError(
                "RemoteEvaluator requires openfeature-provider-ofrep. "
                "Install it with: pip install openfeature-provider-ofrep"
            )
        self._options = options
        base_url = str(options.endpoint).rstrip("/")
        headers_factory: Optional[object] = None
        if options.api_key:
            api_key = options.api_key

            def _headers() -> dict[str, str]:
                return {"X-API-Key": f"{api_key}", "Content-Type": "application/json"}

            headers_factory = _headers
        self._ofrep_provider = OFREPProvider(
            base_url=base_url,
            headers_factory=headers_factory,
        )

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
