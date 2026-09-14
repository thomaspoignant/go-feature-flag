"""Shared helpers for flag evaluation."""

from typing import Any, Type, Union


def changed_flag_keys(
    previous: dict[str, Any],
    current: dict[str, Any],
) -> list[str]:
    """Return the keys that were added, removed or modified between two configs."""
    return sorted(
        key
        for key in previous.keys() | current.keys()
        if previous.get(key) != current.get(key)
    )


def matches_type(value: Any, expected_type: Union[Type, tuple]) -> bool:
    """isinstance(), except that a bool never satisfies a non-bool resolver.

    Python makes bool a subclass of int, so isinstance(True, int) is True and a
    boolean flag would otherwise silently satisfy the integer and float
    resolvers instead of reporting a type mismatch.
    """
    if isinstance(value, bool) and expected_type is not bool:
        return False
    return isinstance(value, expected_type)
