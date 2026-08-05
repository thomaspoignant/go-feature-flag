"""Small helpers shared across the provider."""

from typing import Optional

from openfeature.evaluation_context import EvaluationContext

# Reserved evaluation context attribute marking the evaluation as anonymous.
ANONYMOUS_ATTRIBUTE = "anonymous"

# The two values FeatureEvent.contextKind accepts.
CONTEXT_KIND_ANONYMOUS_USER = "anonymousUser"
CONTEXT_KIND_USER = "user"


def context_kind(evaluation_context: Optional[EvaluationContext]) -> str:
    """Return the kind of context that generated an event.

    Only the boolean `True` marks the evaluation as anonymous. The comparison is
    `is True` on purpose: a truthiness test would read the string "false" — and
    any other non-empty value — as anonymous.

    EvaluationContext is a plain dataclass with no runtime validation, so both
    the context and its attributes can reach us as None. Neither should raise:
    a failure here happens inside a hook, and the SDK turns a hook failure into
    an error result, so a perfectly good evaluation would start returning the
    default value.
    """
    if evaluation_context is None:
        return CONTEXT_KIND_USER

    attributes = evaluation_context.attributes or {}
    if attributes.get(ANONYMOUS_ATTRIBUTE) is True:
        return CONTEXT_KIND_ANONYMOUS_USER
    return CONTEXT_KIND_USER
