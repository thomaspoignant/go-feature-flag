"""Tests for DataCollectorHook."""

from __future__ import annotations

from unittest.mock import MagicMock, call
import pytest

from gofeatureflag_python_provider.hooks.data_collector import (
    DataCollectorHook,
    default_targeting_key,
)
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model.request_data_collector import (
    FeatureEvent,
)
from openfeature.evaluation_context import EvaluationContext
from openfeature.flag_evaluation import FlagEvaluationDetails, FlagType
from openfeature.hook import HookContext

# ---------------------------------------------------------------------------
# Helpers / fixtures
# ---------------------------------------------------------------------------


def _make_options(disable_data_collection: bool = False) -> GoFeatureFlagOptions:
    return GoFeatureFlagOptions(
        endpoint="http://localhost:1031",
        disable_data_collection=disable_data_collection,
    )


def _make_hook_context(
    flag_key: str = "test-flag",
    default_value=False,
    targeting_key: str | None = "user-1",
    anonymous: bool = False,
    flag_type: FlagType = FlagType.BOOLEAN,
) -> HookContext:
    attributes = {}
    if anonymous:
        attributes["anonymous"] = True
    ctx = EvaluationContext(
        targeting_key=targeting_key,
        attributes=attributes,
    )
    return HookContext(
        flag_key=flag_key,
        flag_type=flag_type,
        default_value=default_value,
        evaluation_context=ctx,
    )


def _make_details(
    flag_key: str = "test-flag",
    value=True,
    variant: str | None = "on",
    reason: str = "TARGETING_MATCH",
) -> FlagEvaluationDetails:
    return FlagEvaluationDetails(
        flag_key=flag_key,
        value=value,
        variant=variant,
        reason=reason,
    )


@pytest.fixture
def mock_evaluator():
    evaluator = MagicMock()
    evaluator.is_flag_trackable = MagicMock(return_value=True)
    return evaluator


@pytest.fixture
def mock_event_publisher():
    publisher = MagicMock()
    publisher.add_event = MagicMock()
    publisher.start = MagicMock()
    publisher.stop = MagicMock()
    return publisher


@pytest.fixture
def hook(mock_evaluator, mock_event_publisher):
    return DataCollectorHook(
        options=_make_options(),
        event_publisher=mock_event_publisher,
        evaluator=mock_evaluator,
    )


# ---------------------------------------------------------------------------
# Constructor
# ---------------------------------------------------------------------------


class TestConstructor:
    def test_raises_when_event_publisher_is_none(self, mock_evaluator):
        with pytest.raises((ValueError, TypeError)):
            DataCollectorHook(
                options=_make_options(),
                event_publisher=None,
                evaluator=mock_evaluator,
            )

    def test_raises_when_evaluator_is_none(self, mock_event_publisher):
        with pytest.raises((ValueError, TypeError)):
            DataCollectorHook(
                options=_make_options(),
                event_publisher=mock_event_publisher,
                evaluator=None,
            )


# ---------------------------------------------------------------------------
# after()
# ---------------------------------------------------------------------------


class TestAfter:
    def test_does_not_collect_when_flag_not_trackable(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = False

        hook.after(_make_hook_context(), _make_details(), {})

        mock_evaluator.is_flag_trackable.assert_called_once_with("test-flag")
        mock_event_publisher.add_event.assert_not_called()

    def test_collects_event_for_trackable_flag(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.after(
            _make_hook_context(targeting_key="user-1"),
            _make_details(value=True, variant="on"),
            {},
        )

        mock_evaluator.is_flag_trackable.assert_called_once_with("test-flag")
        mock_event_publisher.add_event.assert_called_once()

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.kind == "feature"
        assert event.key == "test-flag"
        assert event.contextKind == "user"
        assert event.default is False
        assert event.variation == "on"
        assert event.value is True
        assert event.userKey == "user-1"
        assert isinstance(event.creationDate, int)

    def test_collects_event_for_anonymous_user(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.after(
            _make_hook_context(targeting_key="anon-123", anonymous=True),
            _make_details(value=True, variant="on"),
            {},
        )

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.contextKind == "anonymousUser"
        assert event.userKey == "anon-123"

    def test_uses_sdk_default_variation_when_variant_is_none(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.after(_make_hook_context(), _make_details(variant=None), {})

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.variation == "SdkDefault"

    def test_does_not_collect_when_data_collection_disabled(
        self, mock_evaluator, mock_event_publisher
    ):
        hook = DataCollectorHook(
            options=_make_options(disable_data_collection=True),
            event_publisher=mock_event_publisher,
            evaluator=mock_evaluator,
        )
        mock_evaluator.is_flag_trackable.return_value = True

        hook.after(_make_hook_context(), _make_details(), {})

        mock_event_publisher.add_event.assert_not_called()

    def test_uses_default_targeting_key_when_targeting_key_is_none(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.after(_make_hook_context(targeting_key=None), _make_details(), {})

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.userKey == default_targeting_key


# ---------------------------------------------------------------------------
# error()
# ---------------------------------------------------------------------------


class TestError:
    def test_does_not_collect_when_flag_not_trackable(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = False

        hook.error(_make_hook_context(), Exception("boom"), {})

        mock_evaluator.is_flag_trackable.assert_called_once_with("test-flag")
        mock_event_publisher.add_event.assert_not_called()

    def test_collects_error_event_for_trackable_flag(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.error(
            _make_hook_context(default_value=False, targeting_key="user-1"),
            Exception("boom"),
            {},
        )

        mock_evaluator.is_flag_trackable.assert_called_once_with("test-flag")
        mock_event_publisher.add_event.assert_called_once()

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.kind == "feature"
        assert event.key == "test-flag"
        assert event.contextKind == "user"
        assert event.default is True
        assert event.variation == "SdkDefault"
        assert event.value is False
        assert event.userKey == "user-1"
        assert isinstance(event.creationDate, int)

    def test_collects_error_event_for_anonymous_user(
        self, hook, mock_evaluator, mock_event_publisher
    ):
        mock_evaluator.is_flag_trackable.return_value = True

        hook.error(
            _make_hook_context(targeting_key="anon-456", anonymous=True),
            Exception("boom"),
            {},
        )

        event: FeatureEvent = mock_event_publisher.add_event.call_args[0][0]
        assert event.contextKind == "anonymousUser"
        assert event.userKey == "anon-456"

    def test_does_not_collect_when_data_collection_disabled(
        self, mock_evaluator, mock_event_publisher
    ):
        hook = DataCollectorHook(
            options=_make_options(disable_data_collection=True),
            event_publisher=mock_event_publisher,
            evaluator=mock_evaluator,
        )
        mock_evaluator.is_flag_trackable.return_value = True

        hook.error(_make_hook_context(), Exception("boom"), {})

        mock_event_publisher.add_event.assert_not_called()


# ---------------------------------------------------------------------------
# contextKind
# ---------------------------------------------------------------------------


@pytest.mark.parametrize(
    "anonymous_value,expected",
    [
        (True, "anonymousUser"),
        (False, "user"),
        # Only an explicit true means anonymous. A truthiness test would read
        # the string "false" -- and any other non-empty value -- as anonymous.
        ("false", "user"),
        ("true", "user"),
        (1, "user"),
    ],
)
def test_context_kind_requires_anonymous_to_be_exactly_true(
    mock_evaluator, anonymous_value, expected
):
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )
    ctx = EvaluationContext(
        targeting_key="user-1", attributes={"anonymous": anonymous_value}
    )
    hook_context = HookContext(
        flag_key="test-flag",
        flag_type=FlagType.BOOLEAN,
        default_value=False,
        evaluation_context=ctx,
    )

    hook.after(hook_context, _make_details(), {})

    assert publisher.add_event.call_args[0][0].contextKind == expected


def test_context_kind_defaults_to_user_when_attribute_absent(mock_evaluator):
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )

    hook.after(_make_hook_context(), _make_details(), {})

    assert publisher.add_event.call_args[0][0].contextKind == "user"


def test_fallback_results_do_not_produce_a_feature_event(mock_evaluator):
    """The relay proxy evaluated it and already recorded it.

    Emitting here would count the same evaluation twice.
    """
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )
    details = FlagEvaluationDetails(
        flag_key="test-flag",
        value=True,
        variant="on",
        reason="TARGETING_MATCH",
        flag_metadata={"gofeatureflag_evaluated_remotely": True},
    )

    hook.after(_make_hook_context(), details, {})

    publisher.add_event.assert_not_called()


def test_locally_evaluated_results_still_produce_a_feature_event(mock_evaluator):
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )
    details = FlagEvaluationDetails(
        flag_key="test-flag",
        value=True,
        variant="on",
        reason="TARGETING_MATCH",
        flag_metadata={"description": "a flag"},
    )

    hook.after(_make_hook_context(), details, {})

    publisher.add_event.assert_called_once()


# ---------------------------------------------------------------------------
# source and version
# ---------------------------------------------------------------------------


def test_events_are_labelled_as_in_process(mock_evaluator):
    """Every event this provider emits is a local evaluation.

    PROVIDER_CACHE is for values served from a remote-mode cache, and SERVER is
    reserved for events the relay proxy generated itself.
    """
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )

    hook.after(_make_hook_context(), _make_details(), {})
    assert publisher.add_event.call_args[0][0].source == "INPROCESS"

    publisher.reset_mock()
    hook.error(_make_hook_context(), ValueError("boom"), {})
    assert publisher.add_event.call_args[0][0].source == "INPROCESS"


def test_event_carries_the_flag_version_when_present(mock_evaluator):
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )
    details = FlagEvaluationDetails(
        flag_key="test-flag",
        value=True,
        variant="on",
        reason="TARGETING_MATCH",
        flag_metadata={"version": "1.7"},
    )

    hook.after(_make_hook_context(), details, {})

    assert publisher.add_event.call_args[0][0].version == "1.7"


def test_event_version_is_unset_when_the_flag_has_none(mock_evaluator):
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )

    hook.after(_make_hook_context(), _make_details(), {})

    assert publisher.add_event.call_args[0][0].version is None


@pytest.mark.parametrize(
    "raw_version,expected",
    [(2, "2"), (1.7, "1.7"), (True, "True")],
)
def test_a_non_string_flag_version_does_not_break_the_evaluation(
    mock_evaluator, raw_version, expected
):
    """Flag metadata is user-authored, so `version` can be any JSON type.

    An event model that rejects it would raise inside this hook, and the SDK
    turns a hook failure into an error result — so a flag carrying
    `metadata: {version: 2}` would never resolve to its real value again.
    """
    publisher = MagicMock()
    hook = DataCollectorHook(
        options=_make_options(),
        event_publisher=publisher,
        evaluator=mock_evaluator,
    )
    details = FlagEvaluationDetails(
        flag_key="test-flag",
        value=True,
        variant="on",
        reason="TARGETING_MATCH",
        flag_metadata={"version": raw_version},
    )

    hook.after(_make_hook_context(), details, {})

    assert publisher.add_event.call_args[0][0].version == expected
