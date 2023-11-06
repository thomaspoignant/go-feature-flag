from typing import Optional, Any, Dict
from pydantic import BaseModel
from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import (
    TargetingKeyMissingError,
    InvalidContextError,
)


class GoFeatureFlagUser(BaseModel):
    """
    GoFeatureFlagUser is an object representing
    """

    key: str
    anonymous: Optional[bool] = None
    custom: Optional[dict] = None


def user_from_evaluation_context(ctx: EvaluationContext = None) -> GoFeatureFlagUser:
    """
    user_from_evaluation_context is converting an EvaluationContext into a GoFeatureFlagUser
    :param ctx: the EvaluationContext to convert
    :return: a GoFeatureFlagUser
    """
    if ctx is None:
        ctx = {}
    if ctx is None:
        raise InvalidContextError("GO Feature Flag need an Evaluation context to work.")

    if ctx.targeting_key is None or len(ctx.targeting_key) == 0:
        raise TargetingKeyMissingError(
            "targetingKey field MUST be set in your EvaluationContext"
        )

    anonymous = True
    if "anonymous" in ctx.attributes:
        anonymous = ctx.attributes.get("anonymous")

    return GoFeatureFlagUser(
        key=ctx.targeting_key,
        anonymous=anonymous,
        custom=ctx.attributes,
    )


class RequestFlagEvaluation(BaseModel):
    user: GoFeatureFlagUser
    defaultValue: Any = None
