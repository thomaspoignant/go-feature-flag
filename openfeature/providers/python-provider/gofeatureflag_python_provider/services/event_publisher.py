"""
EventPublisher: buffers feature evaluation events and publishes them periodically
or immediately when the pending-event limit is reached.
"""

import logging
import threading
from typing import Any, Optional

from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.model.request_data_collector import (
    FeatureEvent,
)
from gofeatureflag_python_provider.services.api import GoFeatureFlagApi

DEFAULT_FLUSH_INTERVAL_MS: int = 60_000
DEFAULT_MAX_PENDING_EVENTS: int = 10_000
# Upper bound on how long a caller waits for background work: shutdown waiting
# for the runner thread, and any flush that queues behind an in-flight publish.
DEFAULT_SHUTDOWN_TIMEOUT_SECONDS: float = 5.0

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
    - Outside the running window (before ``start()``, after ``stop()``) events are
      dropped rather than buffered: nothing would ever drain them, and a buffer
      allowed to fill would trigger a publish — posting to the collector after
      shutdown reported it was done.
    """

    def __init__(
        self,
        api: GoFeatureFlagApi,
        options: GoFeatureFlagOptions,
    ) -> None:
        if api is None:
            raise ValueError("API cannot be null")
        if options is None:
            raise ValueError("Options cannot be null")

        self._api = api
        self._options = options

        self._events: list[FeatureEvent] = []
        self._lock = threading.Lock()
        # Guards the publish itself so concurrent flushes cannot overlap. It is
        # deliberately never taken by add_event: enqueuing runs inside an
        # evaluation hook and must not block on collector availability.
        self._flush_lock = threading.Lock()
        # One event per runner rather than one shared by every runner: a runner
        # that outlived its stop() keeps its own set event and exits, where a
        # shared event would be cleared by the next start() and revive it,
        # leaving two runners publishing in parallel.
        self._stop_event = threading.Event()
        self._thread: Optional[threading.Thread] = None
        self._running = False
        self._immediate_flush_scheduled = False
        # Drop bookkeeping. Both counters exist to keep the logs readable: an
        # event arriving outside the running window, or a buffer stuck above the
        # cap, happens once per evaluation, so logging each one turns a single
        # condition into a warning per flag evaluation.
        self._dropped_while_not_running = 0
        self._not_running_warned = False
        self._dropped_overflow = 0
        self._overflow_warned = False

        self._exporter_metadata: dict[str, Any] = options.get_exporter_metadata()

    # ------------------------------------------------------------------
    # Public interface
    # ------------------------------------------------------------------

    def start(self) -> None:
        """Start the periodic flush runner. No-op if already running."""
        if self._running:
            return
        with self._lock:
            dropped = self._dropped_while_not_running
            self._dropped_while_not_running = 0
            self._not_running_warned = False
        if dropped:
            # Only reported here: while stopped there is no flush to report on,
            # so this is the first moment the total can be attributed to the
            # period it belongs to.
            logger.warning(
                "EventPublisher: starting, %d event(s) were dropped while it was "
                "not running",
                dropped,
            )
        self._running = True
        self._stop_event = threading.Event()
        self._thread = threading.Thread(
            target=self._run, args=(self._stop_event,), daemon=True
        )
        self._thread.start()

    def stop(self, timeout: float = DEFAULT_SHUTDOWN_TIMEOUT_SECONDS) -> None:
        """Stop the periodic runner and flush any remaining events.

        Bounded: shutdown waits for an in-flight publish rather than racing it,
        but never blocks indefinitely on an unreachable collector.
        """
        if not self._running:
            return
        self._running = False
        self._stop_event.set()
        thread = self._thread
        self._thread = None
        if thread is not None:
            thread.join(timeout=timeout)
            if thread.is_alive():
                # Its stop event is set and is its own, so it exits as soon as
                # the publish it is stuck in returns. Reported because until
                # then a request to the collector is still outstanding.
                logger.warning(
                    "EventPublisher: flush thread still finishing a publish after "
                    "%.1fs; it will exit on its own",
                    timeout,
                )
        self._publish_events(wait=True, timeout=timeout)

    def add_event(self, event: FeatureEvent) -> None:
        """
        Add an event to the buffer.

        Events are only accepted while the publisher is running. Outside that
        window the event is dropped: no runner exists to drain the buffer, so
        holding it would grow to the cap and stay there, and the flush a full
        buffer triggers would post to the collector after stop() returned.

        If the buffer reaches *max_pending_events* after the append, an immediate
        non-blocking flush is triggered (fire-and-forget daemon thread). At most
        one immediate flush runs at a time. If the buffer exceeds the cap
        (max_pending_events * 2), oldest events are dropped.
        """
        max_pending = self._options.max_pending_events or DEFAULT_MAX_PENDING_EVENTS
        cap = max_pending * 2
        with self._lock:
            if not self._running:
                self._dropped_while_not_running += 1
                if self._not_running_warned:
                    logger.debug(
                        "EventPublisher: dropping event for flag %s, publisher is "
                        "not running",
                        event.key,
                    )
                else:
                    self._not_running_warned = True
                    logger.warning(
                        "EventPublisher: dropping events, the publisher is not "
                        "running (start() was never called, or stop() already ran). "
                        "Further drops are logged at debug level",
                    )
                return

            self._events.append(event)
            if len(self._events) > cap:
                dropped = len(self._events) - cap
                self._events = self._events[-cap:]
                self._dropped_overflow += dropped
                if self._overflow_warned:
                    logger.debug(
                        "EventPublisher: buffer overflow, dropped %d oldest event(s)",
                        dropped,
                    )
                else:
                    self._overflow_warned = True
                    logger.warning(
                        "EventPublisher: buffer full (%d events), dropped %d oldest "
                        "event(s); the collector is not accepting batches. Further "
                        "drops are logged at debug level until it recovers",
                        cap,
                        dropped,
                    )
            should_flush = (
                len(self._events) >= max_pending and not self._immediate_flush_scheduled
            )
            if should_flush:
                self._immediate_flush_scheduled = True
                t = threading.Thread(target=self._immediate_flush, daemon=True)
                t.start()

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _run(self, stopper: threading.Event) -> None:
        """Background thread: flush periodically until its own stopper is set."""
        flush_interval_ms = (
            self._options.data_flush_interval or DEFAULT_FLUSH_INTERVAL_MS
        )
        flush_interval_sec = flush_interval_ms / 1000.0
        while not stopper.wait(timeout=flush_interval_sec):
            self._publish_events()

    def _immediate_flush(self) -> None:
        """Target of the fire-and-forget flush spawned by add_event.

        Waits for an in-flight publish instead of skipping: this flush was
        triggered by a full buffer, and the running publish drained the buffer
        before those events arrived, so skipping would leave them queued until
        the next periodic tick — a whole flush interval of growth past the cap.
        """
        try:
            self._publish_events(wait=True)
        finally:
            with self._lock:
                self._immediate_flush_scheduled = False

    def _report_overflow_recovery(self) -> None:
        """Close the overflow report opened by add_event, once a batch got through.

        Pairs one warning at the start of an outage with one at its end carrying
        the total, so the operator learns how much was lost without a warning
        per evaluation in between.
        """
        with self._lock:
            if not self._overflow_warned:
                return
            dropped = self._dropped_overflow
            self._overflow_warned = False
            self._dropped_overflow = 0
        logger.warning(
            "EventPublisher: collector accepted a batch again, %d event(s) were "
            "dropped while the buffer was full",
            dropped,
        )

    def _publish_events(
        self,
        wait: bool = False,
        timeout: float = DEFAULT_SHUTDOWN_TIMEOUT_SECONDS,
    ) -> None:
        """
        Atomically drain the buffer and send events to the collector.

        At most one publish is in flight at a time. A periodic flush that finds
        one already running simply returns — the next tick comes soon enough.
        Callers whose drain cannot wait for a tick (shutdown, and the
        overflow-triggered flush) pass wait=True and queue behind the in-flight
        publish instead, bounded by *timeout* so an unreachable collector never
        blocks them indefinitely.

        On failure the drained batch is re-queued at the front of the buffer so
        no events are lost and chronological order is preserved.
        """
        if wait:
            acquired = self._flush_lock.acquire(timeout=timeout)
        else:
            acquired = self._flush_lock.acquire(blocking=False)
        if not acquired:
            return
        try:
            with self._lock:
                if not self._events:
                    return
                events_to_send = list(self._events)
                self._events.clear()

            # Posted with the buffer lock released: enqueuing an event runs
            # inside an evaluation hook, so holding it here would couple
            # flag-evaluation latency to data-collector availability.
            try:
                self._api.send_event_to_data_collector(
                    events_to_send,
                    self._exporter_metadata,
                )
                self._report_overflow_recovery()
            except Exception as exc:
                logger.exception(
                    "EventPublisher: error publishing events, re-queuing %d event(s): %s",
                    len(events_to_send),
                    exc,
                )
                with self._lock:
                    self._events = events_to_send + self._events
        finally:
            self._flush_lock.release()
