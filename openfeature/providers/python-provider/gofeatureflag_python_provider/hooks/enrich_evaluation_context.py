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
        """Initialize the enrich evaluation context hook."""
        self._metadata = metadata if metadata is not None else {}

    def before(
        self, hook_context: HookContext, hints: dict
    ) -> Optional[EvaluationContext]:
        """Enrich the evaluation context with additional information before flag resolution."""
        ctx = hook_context.evaluation_context
        enriched = dict(ctx.attributes)
        if len(self._metadata) > 0:
            # 'gofeatureflag' is a namespace shared with the caller, who may have
            # supplied flagList and/or currentDateTime. Set only exporterMetadata
            # and preserve every sibling key; replacing the whole object would
            # silently discard the caller's input.
            existing = enriched.get("gofeatureflag")
            # A value that is not a map is replaced rather than treated as an error.
            merged = dict(existing) if isinstance(existing, dict) else {}
            merged["exporterMetadata"] = self._metadata
            enriched["gofeatureflag"] = merged
        return EvaluationContext(
            targeting_key=ctx.targeting_key,
            attributes=enriched,
        )
