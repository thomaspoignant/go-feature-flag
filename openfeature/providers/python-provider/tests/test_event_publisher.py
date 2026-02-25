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
    """
    mock_api = Mock()
    mock_api.send_event_to_data_collector.return_value = None

    options = _make_options(max_pending_events=5)
    cap = 5 * 2  # 10
    publisher = EventPublisher(api=mock_api, options=options)

    with patch.object(publisher._logger, "warning") as mock_warning:
        for i in range(15):
            publisher.add_event(_make_event(key="flag", user_key=f"u{i}"))

        assert len(publisher._events) == cap
        # Dropped 5 events total (15 - 10); may log once per overflow
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
    """
    mock_api = Mock()
    mock_api.send_event_to_data_collector.side_effect = RuntimeError("collector down")

    options = _make_options(max_pending_events=4)
    cap = 8
    publisher = EventPublisher(api=mock_api, options=options)
    publisher.start()

    try:
        # Fill to threshold (4) - triggers immediate flush
        for i in range(4):
            publisher.add_event(_make_event(user_key=f"u{i}"))

        # Wait for flush to fail and re-queue
        time.sleep(0.2)

        # Add more events - buffer grows: 4 (requeued) + new = 4 + 8 = 12
        # After 8 more we're at 12, cap is 8, so we drop 4 oldest
        with patch.object(publisher._logger, "warning") as mock_warning:
            for i in range(4, 12):
                publisher.add_event(_make_event(user_key=f"u{i}"))

            # Should have triggered drop when we exceeded cap
            assert mock_warning.call_count >= 1
            assert len(publisher._events) <= cap
    finally:
        publisher.stop()
