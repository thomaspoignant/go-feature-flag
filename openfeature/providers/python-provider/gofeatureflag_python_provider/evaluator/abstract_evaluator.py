"""
Abstract interface for flag evaluation.

Implementations: RemoteEvaluator (relay proxy), InProcessEvaluator (local/WASM).
"""

from abc import ABC, abstractmethod
from typing import Optional, Union

from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagResolutionDetails


class AbstractEvaluator(ABC):
    """Interface for evaluating feature flags (remote or in-process)."""

    @abstractmethod
    def initialize(
        self, evaluation_context: Optional[EvaluationContext] = None
    ) -> None:
        """Initialize the evaluator (e.g. cache, WebSocket, WASM)."""
        ...

    @abstractmethod
    def shutdown(self) -> None:
        """Release resources (connections, threads, etc.)."""
        ...

    # --- Sync evaluation ---

    @abstractmethod
    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Resolve the flag as a boolean."""
        ...

    @abstractmethod
    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Resolve the flag as a string."""
        ...

    @abstractmethod
    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Resolve the flag as an integer."""
        ...

    @abstractmethod
    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Resolve the flag as a float."""
        ...

    @abstractmethod
    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[list, dict]]:
        """Resolve the flag as an object."""
        ...

    # --- Async evaluation ---

    @abstractmethod
    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Asynchronously resolve the flag as a boolean."""
        ...

    @abstractmethod
    async def resolve_string_details_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Asynchronously resolve the flag as a string."""
        ...

    @abstractmethod
    async def resolve_integer_details_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Asynchronously resolve the flag as an integer."""
        ...

    @abstractmethod
    async def resolve_float_details_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Asynchronously resolve the flag as a float."""
        ...

    @abstractmethod
    async def resolve_object_details_async(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[dict, list]]:
        """Asynchronously resolve the flag as an object."""
        ...

    @abstractmethod
    def is_flag_trackable(self, flag_key: str) -> bool:
        """Check if the flag is trackable."""
        ...
