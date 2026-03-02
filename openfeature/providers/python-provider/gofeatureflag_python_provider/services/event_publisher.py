"""
EventPublisher: buffers feature evaluation events and publishes them periodically
or immediately when the pending-event limit is reached.
"""

import logging
import threading
from typing import Any, Optional

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.request_data_collector import FeatureEvent
from gofeatureflag_python_provider.services.api import GoFeatureFlagApi

DEFAULT_FLUSH_INTERVAL_MS: int = 60_000
DEFAULT_MAX_PENDING_EVENTS: int = 10_000

logger = logging.getLogger(__name__)


class EventPublisher:
    """
    Buffers FeatureEvent objects and sends them in batches to the GO Feature Flag
    relay proxy data collector.

    - Periodic flush: every ``data_flush_interval`` ms (default 60 s).
    - Immediate flush: when the buffer reaches ``max_pending_events`` (fire-and-forget).
      Only one immediate flush runs at a time; further triggers above the threshold
      do not spawn additional threads until that flush finishes.
    - On stop: remaining events are flushed synchronously before returning.
    - On send failure: events are re-queued and retried on the next flush.
    - Buffer cap: if the buffer exceeds ``max_pending_events * 2`` (e.g. collector down),
      oldest events are dropped and a warning is logged to prevent unbounded growth.
    """

    def __init__(
        self,
        api: GoFeatureFlagApi,
        options: GoFeatureFlagOptions,
        logger: Optional[logging.Logger] = None,
    ) -> None:
        if api is None:
            raise ValueError("API cannot be null")
        if options is None:
            raise ValueError("Options cannot be null")

        self._api = api
        self._options = options
        self._logger = logger or logging.getLogger(__name__)

        self._events: list[FeatureEvent] = []
        self._lock = threading.Lock()
        self._stop_event = threading.Event()
        self._thread: Optional[threading.Thread] = None
        self._running = False
        self._immediate_flush_scheduled = False

        meta = dict(options.exporter_metadata or {})
        meta["provider"] = "python"
        meta["openfeature"] = True
        self._exporter_metadata: dict[str, Any] = meta

    # ------------------------------------------------------------------
    # Public interface
    # ------------------------------------------------------------------

    def start(self) -> None:
        """Start the periodic flush runner. No-op if already running."""
        if self._running:
            return
        self._running = True
        self._stop_event.clear()
        self._thread = threading.Thread(target=self._run, daemon=True)
        self._thread.start()

    def stop(self) -> None:
        """Stop the periodic runner and flush any remaining events synchronously."""
        if not self._running:
            return
        self._running = False
        self._stop_event.set()
        if self._thread is not None:
            self._thread.join()
            self._thread = None
        self._publish_events()

    def add_event(self, event: FeatureEvent) -> None:
        """
        Add an event to the buffer.

        If the buffer reaches *max_pending_events* after the append, an immediate
        non-blocking flush is triggered (fire-and-forget daemon thread). At most
        one immediate flush runs at a time. If the buffer exceeds the cap
        (max_pending_events * 2), oldest events are dropped.
        """
        max_pending = self._options.max_pending_events or DEFAULT_MAX_PENDING_EVENTS
        cap = max_pending * 2
        with self._lock:
            self._events.append(event)
            if len(self._events) > cap:
                dropped = len(self._events) - cap
                self._events = self._events[-cap:]
                self._logger.warning(
                    "EventPublisher: buffer overflow, dropped %d oldest event(s)",
                    dropped,
                )
            should_flush = (
                len(self._events) >= max_pending and not self._immediate_flush_scheduled
            )
            if should_flush:
                self._immediate_flush_scheduled = True
                t = threading.Thread(target=self._publish_events, daemon=True)
                t.start()

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _run(self) -> None:
        """Background thread: flush periodically until stopped."""
        flush_interval_ms = (
            self._options.data_flush_interval or DEFAULT_FLUSH_INTERVAL_MS
        )
        flush_interval_sec = flush_interval_ms / 1000.0
        while not self._stop_event.wait(timeout=flush_interval_sec):
            self._publish_events()

    def _publish_events(self) -> None:
        """
        Atomically drain the buffer and send events to the collector.

        On failure the drained batch is re-queued at the front of the buffer
        so no events are lost. Always clears _immediate_flush_scheduled when done.
        """
        try:
            with self._lock:
                if not self._events:
                    return
                events_to_send = list(self._events)
                self._events.clear()

            try:
                self._api.send_event_to_data_collector(
                    events_to_send,
                    self._exporter_metadata,
                )
            except Exception as exc:
                self._logger.error(
                    "EventPublisher: error publishing events, re-queuing %d event(s): %s",
                    len(events_to_send),
                    exc,
                    exc_info=True,
                )
                with self._lock:
                    self._events = events_to_send + self._events
        finally:
            with self._lock:
                self._immediate_flush_scheduled = False
