"""Unit tests for evaluator.util helpers."""

import pytest

from gofeatureflag_python_provider.evaluator.util import changed_flag_keys, matches_type

# ---------------------------------------------------------------------------
# changed_flag_keys
# ---------------------------------------------------------------------------


def test_changed_flag_keys_identical_configs_return_empty():
    flags = {"a": {"defaultValue": True}, "b": {"defaultValue": 1}}
    assert changed_flag_keys(flags, flags) == []


def test_changed_flag_keys_detects_added_key():
    previous = {"a": {"defaultValue": True}}
    current = {"a": {"defaultValue": True}, "b": {"defaultValue": False}}
    assert changed_flag_keys(previous, current) == ["b"]


def test_changed_flag_keys_detects_removed_key():
    previous = {"a": {"defaultValue": True}, "b": {"defaultValue": False}}
    current = {"a": {"defaultValue": True}}
    assert changed_flag_keys(previous, current) == ["b"]


def test_changed_flag_keys_detects_modified_key():
    previous = {"a": {"defaultValue": True}}
    current = {"a": {"defaultValue": False}}
    assert changed_flag_keys(previous, current) == ["a"]


def test_changed_flag_keys_returns_sorted_union_of_changes():
    previous = {"c": {"v": 1}, "a": {"v": 1}, "b": {"v": 1}}
    current = {"c": {"v": 2}, "a": {"v": 1}}  # c modified, b removed
    assert changed_flag_keys(previous, current) == ["b", "c"]


def test_changed_flag_keys_empty_to_empty():
    assert changed_flag_keys({}, {}) == []


def test_changed_flag_keys_empty_to_populated():
    assert changed_flag_keys({}, {"x": {}}) == ["x"]


def test_changed_flag_keys_populated_to_empty():
    assert changed_flag_keys({"x": {}}, {}) == ["x"]


# ---------------------------------------------------------------------------
# matches_type
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "value,expected_type",
    [
        (True, bool),
        (False, bool),
        (42, int),
        (3.14, float),
        ("hello", str),
        ({"a": 1}, dict),
        ([1, 2], list),
    ],
)
def test_matches_type_accepts_exact_type(value, expected_type):
    assert matches_type(value, expected_type) is True


def test_matches_type_rejects_bool_for_int():
    """bool is a subclass of int; the guard must reject it for numeric resolvers."""
    assert matches_type(True, int) is False
    assert matches_type(False, int) is False


def test_matches_type_rejects_bool_for_float():
    assert matches_type(True, float) is False
    assert matches_type(False, float) is False


def test_matches_type_rejects_bool_for_numeric_tuple():
    """Float resolver accepts (float, int); a bool must still be rejected."""
    assert matches_type(True, (float, int)) is False
    assert matches_type(False, (float, int)) is False


def test_matches_type_accepts_int_for_float_tuple():
    """JSON does not distinguish 100 from 100.0, so an int satisfies float."""
    assert matches_type(101, (float, int)) is True
    assert matches_type(3.5, (float, int)) is True


def test_matches_type_rejects_wrong_type():
    assert matches_type("not-a-bool", bool) is False
    assert matches_type(1, str) is False
    assert matches_type([], dict) is False
    assert matches_type({}, list) is False


def test_matches_type_accepts_object_tuple():
    assert matches_type({"a": 1}, (dict, list)) is True
    assert matches_type([1], (dict, list)) is True
    assert matches_type("nope", (dict, list)) is False
