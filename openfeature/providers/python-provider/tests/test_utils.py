"""Tests for the shared provider helpers."""

from __future__ import annotations

import pytest

from gofeatureflag_python_provider.utils import (
    CONTEXT_KIND_ANONYMOUS_USER,
    CONTEXT_KIND_USER,
    context_kind,
)
from openfeature.evaluation_context import EvaluationContext


@pytest.mark.parametrize(
    "anonymous_value,expected",
    [
        (True, CONTEXT_KIND_ANONYMOUS_USER),
        (False, CONTEXT_KIND_USER),
        # Only an explicit true means anonymous. A truthiness test would read
        # the string "false" -- and any other non-empty value -- as anonymous.
        ("true", CONTEXT_KIND_USER),
        ("false", CONTEXT_KIND_USER),
        (1, CONTEXT_KIND_USER),
        (0, CONTEXT_KIND_USER),
        (None, CONTEXT_KIND_USER),
    ],
)
def test_anonymous_must_be_exactly_true(anonymous_value, expected):
    ctx = EvaluationContext(
        targeting_key="user-1", attributes={"anonymous": anonymous_value}
    )

    assert context_kind(ctx) == expected


def test_defaults_to_user_when_the_attribute_is_absent():
    ctx = EvaluationContext(
        targeting_key="user-1", attributes={"email": "test@example.com"}
    )

    assert context_kind(ctx) == CONTEXT_KIND_USER


def test_defaults_to_user_when_there_are_no_attributes():
    assert context_kind(EvaluationContext(targeting_key="user-1")) == CONTEXT_KIND_USER


def test_none_attributes_do_not_raise():
    """EvaluationContext is a dataclass, so attributes=None is constructible.

    Raising here would happen inside a hook, and the SDK turns a hook failure
    into an error result -- so the evaluation would return its default value.
    """
    ctx = EvaluationContext(targeting_key="user-1", attributes=None)

    assert context_kind(ctx) == CONTEXT_KIND_USER


def test_a_missing_evaluation_context_does_not_raise():
    assert context_kind(None) == CONTEXT_KIND_USER
