from enum import Enum


class EvaluationType(Enum):
    """
    This enum represents the type of evaluation that can be performed.
    """

    IN_PROCESS = "InProcess"
    REMOTE = "Remote"
