"""
Unit tests for EventPublisher: immediate flush throttling and buffer cap.
"""

from __future__ import annotations

import threading
import time
from unittest.mock import Mock, patch

import pytest

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import FeatureEvent
from gofeatureflag_python_provider.services.event_publisher import EventPublisher


def _make_event(key: str = "test-flag", user_key: str = "user-1") -> FeatureEvent:
    return FeatureEvent(
        contextKind="user",
        userKey=user_key,
        creationDate=int(time.time()),
        key=key,
        variation="SdkDefault",
        value=True,
        default=False,
    )


def _make_options(
    max_pending_events: int = 5,
    data_flush_interval: int | None = 60_000,
) -> GoFeatureFlagOptions:
    return GoFeatureFlagOptions(
        endpoint="http://localhost:1031",
        max_pending_events=max_pending_events,
        data_flush_interval=data_flush_interval,
    )


def test_only_one_immediate_flush_thread_when_collector_slow():
    """
    When the buffer reaches the threshold multiple times while a flush is in
    progress (collector slow/blocking), only one immediate flush thread is
    started. Prevents thread exhaustion.
    """
    send_called = threading.Event()
    release_send = threading.Event()

    def blocking_send(*args, **kwargs):
        send_called.set()
        release_send.wait()
        return None

    mock_api = Mock()
    mock_api.send_event_to_data_collector = Mock(side_effect=blocking_send)

    options = _make_options(max_pending_events=3)
    publisher = EventPublisher(api=mock_api, options=options)
    publisher.start()

    try:
        # Fill buffer to threshold (3) - triggers immediate flush
        for i in range(3):
            publisher.add_event(_make_event(key="a", user_key=f"u{i}"))

        # Wait for flush thread to enter send (blocking)
        assert send_called.wait(timeout=2.0), "send should have been called"

        # Add more events while flush is blocked - should NOT spawn new threads
        threads_before = threading.active_count()
        for i in range(10):
            publisher.add_event(_make_event(key="b", user_key=f"v{i}"))
        threads_after = threading.active_count()

        # Should not have spawned 10 more threads (only 1 flush thread total)
        assert (
            threads_after <= threads_before + 1
        ), "expected at most one additional thread from immediate flush"

        release_send.set()
    finally:
        publisher.stop()

    # send_event_to_data_collector called at least once (periodic flush or immediate)
    assert mock_api.send_event_to_data_collector.call_count >= 1


def test_buffer_cap_drops_oldest_events():
    """
    When the buffer exceeds the cap (max_pending_events * 2), oldest events
    are dropped and a warning is logged. Buffer size stays at cap.

    Prevents the immediate flush from running (by mocking _publish_events) so
    that only the cap-and-drop logic is tested, avoiding races with the
    background flush thread.
    """
    mock_api = Mock()
    mock_api.send_event_to_data_collector.return_value = None

    options = _make_options(max_pending_events=5)
    cap = 5 * 2  # 10
    publisher = EventPublisher(api=mock_api, options=options)

    with patch.object(publisher, "_publish_events"):
        with patch.object(publisher._logger, "warning") as mock_warning:
            for i in range(15):
                publisher.add_event(_make_event(key="flag", user_key=f"u{i}"))

            assert len(publisher._events) == cap
            assert mock_warning.call_count >= 1
            call_args = mock_warning.call_args[0]
            assert "dropped" in call_args[0].lower()

    # Kept the last 10 events (indices 5..14)
    assert publisher._events[0].userKey == "u5"
    assert publisher._events[-1].userKey == "u14"


def test_buffer_cap_with_failing_collector_requeue():
    """
    When flush fails and events are re-queued, if the buffer exceeds cap
    after re-queue, we still enforce the cap on next add_event.

    Uses a threading.Event to wait until the failing send has run and
    re-queued. Uses a long data_flush_interval so the periodic flush
    does not run during the test. Asserts that the buffer never exceeds
    cap and that the collector was attempted (re-queue scenario ran).
    """
    send_attempted = threading.Event()

    def failing_send(*args, **kwargs):
        send_attempted.set()
        raise RuntimeError("collector down")

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = failing_send

    options = _make_options(max_pending_events=4, data_flush_interval=3600_000)
    cap = 8
    publisher = EventPublisher(api=mock_api, options=options)
    publisher.start()

    try:
        for i in range(4):
            publisher.add_event(_make_event(user_key=f"u{i}"))

        assert send_attempted.wait(timeout=2.0), "flush should have attempted send"

        for i in range(4, 12):
            publisher.add_event(_make_event(user_key=f"u{i}"))

        assert len(publisher._events) <= cap
        assert mock_api.send_event_to_data_collector.call_count >= 1
    finally:
        publisher.stop()


def test_concurrent_flushes_do_not_overlap():
    """Publishing must be single-flight: two flushes must not post at once."""
    in_flight = threading.Event()
    overlapped = threading.Event()
    concurrent = 0
    guard = threading.Lock()

    def _slow_send(events, meta):
        nonlocal concurrent
        with guard:
            concurrent += 1
            if concurrent > 1:
                overlapped.set()
        in_flight.set()
        time.sleep(0.2)
        with guard:
            concurrent -= 1

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = _slow_send
    options = GoFeatureFlagOptions(endpoint="http://localhost:1031")
    publisher = EventPublisher(api=mock_api, options=options)

    publisher.add_event(_make_event())
    first = threading.Thread(target=publisher._publish_events, daemon=True)
    first.start()
    assert in_flight.wait(timeout=2.0), "first publish should have started"

    # A second flush arriving mid-post must not start its own request.
    publisher.add_event(_make_event(user_key="u2"))
    publisher._publish_events()
    first.join(timeout=2.0)

    assert not overlapped.is_set()


def test_enqueue_is_not_blocked_by_an_in_flight_publish():
    """add_event runs inside an evaluation hook; a slow collector must not stall it."""
    in_flight = threading.Event()
    release = threading.Event()

    def _blocking_send(events, meta):
        in_flight.set()
        release.wait(timeout=5.0)

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = _blocking_send
    options = GoFeatureFlagOptions(endpoint="http://localhost:1031")
    publisher = EventPublisher(api=mock_api, options=options)

    publisher.add_event(_make_event())
    threading.Thread(target=publisher._publish_events, daemon=True).start()
    assert in_flight.wait(timeout=2.0)

    try:
        started = time.monotonic()
        publisher.add_event(_make_event(user_key="u2"))
        assert time.monotonic() - started < 1.0
    finally:
        release.set()


def test_stop_is_bounded_when_the_collector_hangs():
    """Shutdown must bound how long it waits for background work."""
    release = threading.Event()

    def _hanging_send(events, meta):
        release.wait(timeout=10.0)

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = _hanging_send
    options = GoFeatureFlagOptions(endpoint="http://localhost:1031")
    publisher = EventPublisher(api=mock_api, options=options)
    publisher.start()
    publisher.add_event(_make_event())
    threading.Thread(target=publisher._publish_events, daemon=True).start()
    time.sleep(0.1)

    try:
        started = time.monotonic()
        publisher.stop(timeout=0.5)
        assert time.monotonic() - started < 3.0
    finally:
        release.set()


def test_overflow_flush_waits_for_an_in_flight_publish():
    """The flush a full buffer triggers must not be skipped.

    The publish already running drained the buffer before these events arrived,
    so skipping would leave them queued past the cap until the next periodic
    tick — a whole flush interval away.
    """
    in_flight = threading.Event()
    release = threading.Event()
    sent_batches = []

    def _send(events, meta):
        sent_batches.append(list(events))
        if len(sent_batches) == 1:
            in_flight.set()
            release.wait(timeout=5.0)

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = _send
    publisher = EventPublisher(api=mock_api, options=_make_options(2))

    # Occupy the publish lock with a slow, already-running flush.
    publisher.add_event(_make_event(user_key="u0"))
    threading.Thread(target=publisher._publish_events, daemon=True).start()
    assert in_flight.wait(timeout=2.0)

    # These arrive after that publish took its snapshot, and reaching the
    # threshold spawns the overflow flush.
    publisher.add_event(_make_event(user_key="u1"))
    publisher.add_event(_make_event(user_key="u2"))
    release.set()

    deadline = time.monotonic() + 3.0
    while len(sent_batches) < 2 and time.monotonic() < deadline:
        time.sleep(0.01)

    assert len(sent_batches) == 2
    assert [event.userKey for event in sent_batches[1]] == ["u1", "u2"]


def test_start_after_a_timed_out_stop_leaves_a_single_runner():
    """A stop() that timed out must not leave a second runner behind.

    The old runner is still alive; giving each runner its own stop event means
    the next start() cannot revive it into publishing alongside the new one.
    """
    release = threading.Event()
    in_flight = threading.Event()

    def _hanging_send(events, meta):
        in_flight.set()
        release.wait(timeout=10.0)

    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = _hanging_send
    # Tick fast enough that the runner thread itself is the one stuck in the
    # publish when stop() gives up waiting for it.
    publisher = EventPublisher(api=mock_api, options=_make_options(2, 50))
    publisher.add_event(_make_event())
    publisher.start()
    old_thread = publisher._thread
    old_stopper = publisher._stop_event
    assert in_flight.wait(timeout=2.0)

    try:
        publisher.stop(timeout=0.2)
        assert old_thread.is_alive()  # still stuck in the hanging publish
        assert old_stopper.is_set()

        publisher.start()

        assert publisher._thread is not old_thread
        # The old runner is still waiting on the event it was started with, so
        # it stays stopped. Sharing one event would clear it here and put that
        # thread back to work alongside the new one.
        assert old_stopper.is_set()
        assert publisher._stop_event is not old_stopper
    finally:
        release.set()
        publisher.stop(timeout=1.0)


def test_exporter_metadata_always_carries_the_reserved_keys():
    """Without them, events cannot be attributed to an SDK."""
    mock_api = Mock()
    options = GoFeatureFlagOptions(endpoint="http://localhost:1031")
    publisher = EventPublisher(api=mock_api, options=options)

    assert publisher._exporter_metadata == {"provider": "python", "openfeature": True}

    publisher_with_meta = EventPublisher(
        api=mock_api,
        options=GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            exporter_metadata={"appName": "demo"},
        ),
    )
    assert publisher_with_meta._exporter_metadata == {
        "appName": "demo",
        "provider": "python",
        "openfeature": True,
    }
