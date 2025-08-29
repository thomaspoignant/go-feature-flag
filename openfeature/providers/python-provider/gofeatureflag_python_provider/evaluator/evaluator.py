from abc import ABC, abstractmethod
from typing import Any, Optional
from openfeature.evaluation_context import EvaluationContext
from openfeature.provider import FlagResolutionDetails


class IEvaluator(ABC):
    """
    IEvaluator is an interface that represents the evaluation of a feature flag.
    It can have multiple implementations: Remote or InProcess.
    """

    @abstractmethod
    async def initialize(self) -> None:
        """
        Initialize the evaluator.
        """
        pass

    @abstractmethod
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
        pass

    @abstractmethod
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
        pass

    @abstractmethod
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
        pass

    @abstractmethod
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
        pass

    @abstractmethod
    def is_flag_trackable(self, flag_key: str) -> bool:
        """
        Check if the flag is trackable.

        Args:
            flag_key: The key of the flag to check.

        Returns:
            True if the flag is trackable, false otherwise.
        """
        pass

    @abstractmethod
    async def dispose(self) -> None:
        """
        Dispose the evaluator.
        """
        pass
