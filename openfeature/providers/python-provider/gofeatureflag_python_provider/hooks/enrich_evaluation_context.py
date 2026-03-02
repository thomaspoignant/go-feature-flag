from typing import Optional

from openfeature.evaluation_context import EvaluationContext
from openfeature.hook import Hook, HookContext


class EnrichEvaluationContextHook(Hook):
    """
    Enriches the evaluation context with additional information before flag resolution.
    Adds a 'gofeatureflag' attribute containing the given metadata when non-empty,
    so the relay proxy can use it (e.g. for analytics).
    """

    def __init__(self, metadata: Optional[dict] = None):
        self._metadata = metadata if metadata is not None else {}

    def before(
        self, hook_context: HookContext, hints: dict
    ) -> Optional[EvaluationContext]:
        ctx = hook_context.evaluation_context
        enriched = dict(ctx.attributes)
        if len(self._metadata) > 0:
            enriched["gofeatureflag"] = {"exporterMetadata": self._metadata}
        return EvaluationContext(
            targeting_key=ctx.targeting_key,
            attributes=enriched,
        )
