import hashlib
import json
from typing import Optional, Any

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import (
    TargetingKeyMissingError,
    InvalidContextError,
)
from pydantic import SkipValidation

from gofeatureflag_python_provider.options import BaseModel


class GoFeatureFlagEvaluationContext(BaseModel):
    """
    GoFeatureFlagUser is an object representing
    """

    key: str
    custom: Optional[dict] = None

    def hash(self):
        dhash = hashlib.md5()
        encoded = json.dumps(
            {"key": self.key, "custom": self.custom}, sort_keys=True
        ).encode()
        dhash.update(encoded)
        return dhash.hexdigest()


def convert_evaluation_context(
    ctx: EvaluationContext = None,
) -> GoFeatureFlagEvaluationContext:
    """
    convert_evaluation_context is converting an OpenFeature EvaluationContext into a GO Feature Flag context
    :param ctx: the EvaluationContext to convert
    :return: a GO Feature Flag context
    """
    if ctx is None:
        ctx = {}
    if ctx is None:
        raise InvalidContextError("GO Feature Flag need an Evaluation context to work.")

    if ctx.targeting_key is None or len(ctx.targeting_key) == 0:
        raise TargetingKeyMissingError(
            "targetingKey field MUST be set in your EvaluationContext"
        )

    return GoFeatureFlagEvaluationContext(
        key=ctx.targeting_key,
        custom=ctx.attributes,
    )


class RequestFlagEvaluation(BaseModel):
    user: GoFeatureFlagEvaluationContext
    defaultValue: SkipValidation[Any] = None
