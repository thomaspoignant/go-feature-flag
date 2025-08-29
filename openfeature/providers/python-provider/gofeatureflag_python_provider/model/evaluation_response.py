from typing import Any, Dict, Optional
from pydantic import BaseModel


class EvaluationResponse(BaseModel):
    """
    EvaluationResponse represents the response from the GO Feature Flag evaluation.
    """

    # Variation is the variation of the flag that was returned by the evaluation.
    variation_type: Optional[str] = None

    # trackEvents indicates whether events should be tracked for this evaluation.
    track_events: bool

    # reason is the reason for the evaluation result.
    reason: Optional[str] = None

    # errorCode is the error code for the evaluation result, if any.
    error_code: Optional[str] = None

    # errorDetails provides additional details about the error, if any.
    error_details: Optional[str] = None

    # value is the evaluated value of the flag.
    value: Optional[Any] = None

    # metadata is a dictionary containing additional metadata about the evaluation.
    metadata: Optional[Dict[str, Any]] = None
