"""
Evaluator interface.
"""

from abc import ABC, abstractmethod
from typing import Mapping, Optional, Sequence, Union
from openfeature.evaluation_context import EvaluationContext
from openfeature.provider import FlagResolutionDetails, FlagValueType


class IEvaluator(ABC):
    """
    IEvaluator is an interface that represents the evaluation of a feature flag.
    It can have multiple implementations: Remote or InProcess.
    """

    @abstractmethod
    def initialize(self, evaluation_context: EvaluationContext) -> None:
        """
        Initialize the evaluator.
        """

    @abstractmethod
    def evaluate_boolean(
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

    @abstractmethod
    async def evaluate_boolean_async(
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

    @abstractmethod
    def evaluate_string(
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
        """

    @abstractmethod
    async def evaluate_string_async(
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

    @abstractmethod
    def evaluate_integer(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """
        Evaluates an integer flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    async def evaluate_integer_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """
        Evaluates an integer flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    def evaluate_float(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Evaluates a float flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    async def evaluate_float_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """
        Evaluates a float flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    def evaluate_object(
        self,
        flag_key: str,
        default_value: Union[Sequence[FlagValueType], Mapping[str, FlagValueType]],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[
        Union[Sequence[FlagValueType], Mapping[str, FlagValueType]]
    ]:
        """
        Evaluates an object flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    async def evaluate_object_async(
        self,
        flag_key: str,
        default_value: Union[Sequence[FlagValueType], Mapping[str, FlagValueType]],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[
        Union[Sequence[FlagValueType], Mapping[str, FlagValueType]]
    ]:
        """
        Evaluates an object flag.

        Args:
            flag_key: The key of the flag to evaluate.
            default_value: The default value to return if the flag is not found.
            evaluation_context: The context in which to evaluate the flag.

        Returns:
            The resolution details of the flag evaluation.
        """

    @abstractmethod
    def is_flag_trackable(self, flag_key: str) -> bool:
        """
        Check if the flag is trackable.

        Args:
            flag_key: The key of the flag to check.

        Returns:
            True if the flag is trackable, false otherwise.
        """

    @abstractmethod
    def shutdown(self) -> None:
        """
        Dispose the evaluator.
        """
