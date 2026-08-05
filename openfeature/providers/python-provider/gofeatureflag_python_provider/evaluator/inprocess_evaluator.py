"""
In-process evaluator: evaluates flags locally via the WASM module.
Fetches flag configuration from the relay proxy, stores it locally, and polls for updates.
"""

import asyncio
import logging
import random
import threading
from enum import Enum
from typing import Any, Callable, Optional, Type, TypeVar, Union

from openfeature.evaluation_context import EvaluationContext
from openfeature.exception import (
    FlagNotFoundError,
    GeneralError,
    InvalidContextError,
    OpenFeatureError,
    ParseError,
    ProviderFatalError,
    ProviderNotReadyError,
    TargetingKeyMissingError,
    TypeMismatchError,
)
from openfeature.flag_evaluation import FlagResolutionDetails, Reason
from pydantic_core import PydanticSerializationError

from gofeatureflag_python_provider.evaluator.abstract_evaluator import AbstractEvaluator
from gofeatureflag_python_provider.evaluator.util import changed_flag_keys, matches_type
from gofeatureflag_python_provider.exceptions import UnauthorizedError
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from gofeatureflag_python_provider.services.api import GoFeatureFlagApi
from gofeatureflag_python_provider.services.ofrep import build_ofrep_provider
from gofeatureflag_python_provider.wasm import (
    EvaluateWasm,
    WasmEvaluationTrapError,
    WasmFlagContext,
    WasmInput,
    WasmInvalidResultError,
    WasmNotLoadedError,
    WasmPoolTimeoutError,
)

logger = logging.getLogger(__name__)

T = TypeVar("T")


class EngineErrorCode(str, Enum):
    FLAG_NOT_FOUND = "FLAG_NOT_FOUND"
    TYPE_MISMATCH = "TYPE_MISMATCH"
    TARGETING_KEY_MISSING = "TARGETING_KEY_MISSING"
    INVALID_CONTEXT = "INVALID_CONTEXT"
    PARSE_ERROR = "PARSE_ERROR"
    GENERAL = "GENERAL"


# Raw engine codes where the provider, rather than the flag, looks to be at
# fault. The relay proxy is authoritative and reachable, so re-evaluating there
# turns a local failure into a correct answer.
#
# PARSE_ERROR means the module could not read the payload this provider built
# for it, and GENERAL is the engine's catch-all — which the bundled engine also
# returns for anything a newer relay proxy understands and it does not. Both can
# therefore succeed remotely.
#
# FLAG_CONFIG is deliberately absent: it names the flag itself as invalid, a
# verdict the relay proxy reaches from the same configuration, so a fallback
# would only add latency to the same wrong answer.
_FALLBACK_ERROR_CODES = frozenset(
    {EngineErrorCode.PARSE_ERROR, EngineErrorCode.GENERAL}
)

# Flag metadata marking a result that came from the relay proxy rather than the
# local engine. Also tells the data collector to skip it: the relay proxy has
# already recorded that evaluation.
_EVALUATED_REMOTELY_KEY = "gofeatureflag_evaluated_remotely"

# Flag metadata key carrying the flag version reported by the engine.
_VERSION_KEY = "version"

# How many consecutive failed refreshes before the provider reports itself stale.
# The last known-good configuration keeps serving throughout.
_STALE_AFTER_CONSECUTIVE_FAILURES = 3

# Used when the option is explicitly unset. Mirrors the option's own default.
_DEFAULT_POLL_INTERVAL_SECONDS = 120

# Fraction by which each poll interval is randomly shortened or lengthened, so a
# fleet restarted together does not poll the relay proxy in lockstep.
_POLL_JITTER_RATIO = 0.1


class InProcessEvaluator(AbstractEvaluator):
    """Evaluates flags in-process via the WASM module. Fetches and polls flag configuration."""

    def __init__(
        self,
        options: GoFeatureFlagOptions,
        api: GoFeatureFlagApi,
        on_configuration_changed: Optional[Callable[[list[str]], None]] = None,
        on_stale: Optional[Callable[[str], None]] = None,
        on_ready: Optional[Callable[[], None]] = None,
    ) -> None:
        self._options = options
        self._api = api
        # Provider events are emitted through callbacks so the evaluator stays
        # independent of the OpenFeature provider that owns it.
        self._on_configuration_changed = on_configuration_changed
        self._on_stale = on_stale
        self._on_ready = on_ready
        self._consecutive_refresh_failures = 0
        self._stale = False
        # OFREP client for remote fallback, built on first use.
        self._ofrep_fallback = None
        self._fallback_lock = threading.Lock()
        # None means "no configuration loaded yet", which is distinct from
        # "loaded and empty" and must be reported as PROVIDER_NOT_READY rather
        # than blamed on the caller's flag key.
        self._flags: Optional[dict[str, Any]] = None
        self._etag: Optional[str] = None
        self._evaluation_context_enrichment: dict[str, Any] = {}
        self._lock = threading.Lock()
        self._poll_stopper: Optional[threading.Event] = None
        self._poll_thread: Optional[threading.Thread] = None
        self._poll_interval_seconds: int = (
            options.flag_config_poll_interval_seconds or _DEFAULT_POLL_INTERVAL_SECONDS
        )
        self._wasm = EvaluateWasm(
            wasm_path=options.wasm_file_path,
            pool_size=options.get_wasm_pool_size(),
        )

    def initialize(
        self, evaluation_context: Optional[EvaluationContext] = None
    ) -> None:
        """Fetch initial flag configuration, initialize WASM, and start background polling.

        Safe to call more than once: any poller and WASM pool left behind by a
        previous initialization are torn down first, so a second call neither
        leaks a polling thread nor double-instantiates the evaluation engine.

        Failing to become ready terminates abnormally, which is how a provider
        reports it to the SDK. Only credentials are fatal; every other failure
        stays recoverable and leaves the provider in ERROR rather than FATAL.
        The SDK short-circuits evaluation for NOT_READY and FATAL but not for
        ERROR, so evaluations still reach this evaluator afterwards and report
        PROVIDER_NOT_READY off the configuration that was never loaded. Nothing
        re-initializes on its own: polling never started.
        """
        self._stop_polling()
        self._wasm.dispose()

        try:
            self._wasm.initialize()
            response = self._api.retrieve_flag_configuration(
                flags=self._options.evaluation_flag_list
            )
            if response is None:
                # No If-None-Match was sent, so 304 is a protocol violation here.
                raise GeneralError(
                    "relay proxy answered 304 Not Modified to an unconditional "
                    "flag configuration request"
                )
        except UnauthorizedError as exc:
            # Credentials cannot be repaired by retrying, so this has to be
            # fatal rather than an error the provider retries unattended.
            self._wasm.dispose()
            raise ProviderFatalError(str(exc)) from exc
        except WasmNotLoadedError as exc:
            self._wasm.dispose()
            raise ProviderFatalError(str(exc)) from exc
        except Exception as exc:
            # The engine is loaded before the configuration is fetched, and the
            # SDK never calls shutdown() on a provider whose initialize() raised,
            # so an engine left behind here outlives the provider that owns it.
            #
            # Everything leaving initialize() is an OpenFeatureError: the SDK
            # reads error_code off one to choose between ERROR and FATAL, and a
            # foreign exception only ever collapses to a bare GENERAL.
            self._wasm.dispose()
            raise GeneralError(
                f"failed to initialize the in-process evaluator: {exc}"
            ) from exc

        with self._lock:
            self._flags = response.flags or {}
            self._etag = response.etag
            self._evaluation_context_enrichment = (
                response.evaluation_context_enrichment or {}
            )
        self._start_polling()

    def _start_polling(self) -> None:
        """Start the background polling thread."""
        stopper = threading.Event()
        self._poll_stopper = stopper
        self._poll_thread = threading.Thread(
            target=self._background_poll,
            args=(stopper,),
            daemon=True,
        )
        self._poll_thread.start()

    def _stop_polling(self) -> None:
        """Signal the polling thread and join it. Safe to call when not polling."""
        if self._poll_stopper is not None:
            self._poll_stopper.set()
        if self._poll_thread is not None:
            self._poll_thread.join(timeout=5.0)
        self._poll_thread = None
        self._poll_stopper = None

    def _next_poll_delay(self) -> float:
        """The poll interval with jitter applied.

        Without it, a fleet restarted together polls the relay proxy in lockstep
        for as long as it stays up.
        """
        jitter = 1.0 + random.uniform(-_POLL_JITTER_RATIO, _POLL_JITTER_RATIO)
        return self._poll_interval_seconds * jitter

    def _background_poll(self, stopper: threading.Event) -> None:
        """Loop: wait poll_interval, then refresh flag config until stopped.

        The stopper is passed in rather than read from self. Re-initialization
        clears the attribute, so a poller that outlived its join timeout would
        otherwise dereference None instead of exiting on its own event.
        """
        while not stopper.is_set():
            stopper.wait(self._next_poll_delay())
            if stopper.is_set():
                break
            self._refresh_flag_configuration()

    def _refresh_flag_configuration(self) -> None:
        """Refresh the stored configuration, preserving it on any failed refresh.

        A failed refresh must leave the configuration *and* the stored ETag
        untouched. Advancing the ETag without applying a configuration pins the
        provider to state it never received, after which the server answers 304
        forever and the stale configuration becomes permanent.
        """
        with self._lock:
            etag = self._etag
        try:
            # The flag list applies to every fetch. Sending it only at startup
            # would silently widen the configuration on the first poll.
            response = self._api.retrieve_flag_configuration(
                etag=etag, flags=self._options.evaluation_flag_list
            )
        except Exception:
            logger.exception("Failed to refresh flag configuration")
            self._record_refresh_failure("failed to fetch the flag configuration")
            return

        if response is None:
            # 304 Not Modified: nothing changed, so touch nothing. The server
            # answered, so this is a successful refresh, not a failure.
            self._record_refresh_success(changed_flags=[])
            return
        # A missing flag map never reaches here: the API rejects that response as
        # malformed, which is a failed fetch above. An empty one is a relay proxy
        # that legitimately serves no flags, and is applied as-is — treating it
        # as a failure would drive a correctly-configured provider to STALE and
        # answer PROVIDER_NOT_READY instead of FLAG_NOT_FOUND, forever.
        enrichment = response.evaluation_context_enrichment or {}
        with self._lock:
            changed_flags = changed_flag_keys(self._flags or {}, response.flags)
            if enrichment != self._evaluation_context_enrichment:
                # Enrichment feeds every evaluation, so a change to it changes
                # the effective configuration even when no flag differs.
                changed_flags = changed_flags or ["evaluationContextEnrichment"]
            self._flags = response.flags
            self._evaluation_context_enrichment = enrichment
            self._etag = response.etag
        self._record_refresh_success(changed_flags=changed_flags)

    def _record_refresh_success(self, changed_flags: list[str]) -> None:
        """Reset the failure streak, announce recovery, and announce real changes.

        The configuration-changed event fires only when the content actually
        differs: an ETag tells us a response was fresh, not that it was
        different, and a provider that cannot tell those apart emits on every
        poll.
        """
        with self._lock:
            was_stale = self._stale
            self._consecutive_refresh_failures = 0
            self._stale = False
        if was_stale and self._on_ready is not None:
            self._on_ready()
        if changed_flags and self._on_configuration_changed is not None:
            self._on_configuration_changed(changed_flags)

    def _record_refresh_failure(self, message: str) -> None:
        """Count a failed refresh, reporting stale once the streak is long enough."""
        with self._lock:
            self._consecutive_refresh_failures += 1
            became_stale = (
                self._consecutive_refresh_failures >= _STALE_AFTER_CONSECUTIVE_FAILURES
                and not self._stale
            )
            if became_stale:
                self._stale = True
        if became_stale and self._on_stale is not None:
            self._on_stale(message)

    def shutdown(self) -> None:
        """Stop the polling thread, dispose WASM, and release resources."""
        self._stop_polling()
        self._wasm.dispose()
        with self._lock:
            self._flags = None
            self._etag = None
            self._evaluation_context_enrichment = {}

    # ------------------------------------------------------------------
    # Generic resolver
    # ------------------------------------------------------------------

    @staticmethod
    def _build_eval_context(
        evaluation_context: Optional[EvaluationContext],
    ) -> dict[str, Any]:
        """Convert an OpenFeature EvaluationContext to a flat dict for the WASM input."""
        ctx: dict[str, Any] = {}
        if evaluation_context is None:
            return ctx
        if evaluation_context.targeting_key:
            ctx["targetingKey"] = evaluation_context.targeting_key
        if evaluation_context.attributes:
            ctx.update(evaluation_context.attributes)
        return ctx

    @staticmethod
    def _error_for_code(
        flag_key: str, error_code: str, details: Optional[str]
    ) -> Exception:
        """Map a raw engine error code onto the SDK exception that represents it.

        An unrecognised code maps to GeneralError rather than escaping as an
        unmapped language-level exception.
        """
        if error_code == EngineErrorCode.FLAG_NOT_FOUND:
            return FlagNotFoundError(details or f"Flag '{flag_key}' not found")
        if error_code == EngineErrorCode.TYPE_MISMATCH:
            return TypeMismatchError(details or f"Type mismatch for flag '{flag_key}'")
        if error_code == EngineErrorCode.TARGETING_KEY_MISSING:
            return TargetingKeyMissingError(details or "Targeting key missing")
        if error_code == EngineErrorCode.INVALID_CONTEXT:
            return InvalidContextError(details or "Invalid context")
        if error_code == EngineErrorCode.PARSE_ERROR:
            return ParseError(details or f"Error parsing flag '{flag_key}'")
        return GeneralError(
            details or f"Error evaluating flag '{flag_key}': {error_code}"
        )

    def _fallback_provider(self):
        """The OFREP client used for remote fallback, built on first use."""
        with self._fallback_lock:
            if self._ofrep_fallback is None:
                self._ofrep_fallback = build_ofrep_provider(self._options)
            return self._ofrep_fallback

    def _evaluate_remotely(
        self,
        flag_key: str,
        default_value: T,
        evaluation_context: Optional[EvaluationContext],
        remote_resolver: str,
        raw_error_code: str,
        local_error: Exception,
    ) -> FlagResolutionDetails[T]:
        """Re-evaluate through the relay proxy after a local engine failure.

        Every qualifying failure falls back, with no suppression or back-off, so
        a persistently malformed flag turns each evaluation into a network round
        trip. That is intentional but expensive, which is why each one is logged
        and the result is marked.
        """
        logger.warning(
            "In-process evaluation of flag '%s' failed with %s; "
            "falling back to remote evaluation",
            flag_key,
            raw_error_code,
        )
        try:
            provider = self._fallback_provider()
            details = getattr(provider, remote_resolver)(
                flag_key, default_value, evaluation_context
            )
        except Exception:
            # The in-process failure is the root cause, so it is what the caller
            # sees; the remote failure is logged rather than substituted.
            logger.warning(
                "Remote fallback for flag '%s' also failed", flag_key, exc_info=True
            )
            raise local_error from None

        flag_metadata = dict(details.flag_metadata or {})
        flag_metadata[_EVALUATED_REMOTELY_KEY] = True
        return FlagResolutionDetails(
            value=details.value,
            reason=details.reason,
            variant=details.variant,
            error_code=details.error_code,
            error_message=details.error_message,
            flag_metadata=flag_metadata,
        )

    def _resolve_generic(
        self,
        flag_key: str,
        default_value: T,
        expected_type: Union[Type[T], tuple],
        evaluation_context: Optional[EvaluationContext],
        remote_resolver: str,
    ) -> FlagResolutionDetails[T]:
        with self._lock:
            flags = self._flags
            enrichment = dict(self._evaluation_context_enrichment)

        if flags is None:
            # No configuration has been loaded yet. Reporting FLAG_NOT_FOUND
            # here would misattribute an infrastructure failure to the caller's
            # flag key.
            raise ProviderNotReadyError("flag configuration has not been loaded yet")

        flag = flags.get(flag_key)
        if flag is None:
            raise FlagNotFoundError(
                f"Flag '{flag_key}' not found in local configuration"
            )

        wasm_input = WasmInput(
            flagKey=flag_key,
            flag=flag,
            evalContext=self._build_eval_context(evaluation_context),
            flagContext=WasmFlagContext(
                defaultSdkValue=default_value,
                evaluationContextEnrichment=enrichment,
            ),
        )

        try:
            response = self._wasm.evaluate(wasm_input)
        except (
            WasmEvaluationTrapError,
            WasmInvalidResultError,
            WasmNotLoadedError,
            WasmPoolTimeoutError,
            PydanticSerializationError,
        ) as exc:
            # An engine fault is a GENERAL-class failure, so it falls back to the
            # relay proxy. A trapped instance has already been discarded and
            # rebuilt by the WASM host; we deliberately do not retry locally on
            # the fresh instance first.
            #
            # PydanticSerializationError covers inputs the host cannot even
            # serialize for the module (circular references, pydantic's own
            # ~255-level nesting cap, unserializable attribute values).
            return self._evaluate_remotely(
                flag_key,
                default_value,
                evaluation_context,
                remote_resolver,
                raw_error_code=EngineErrorCode.GENERAL,
                local_error=GeneralError(
                    f"WASM evaluation failed for flag '{flag_key}': {exc}"
                ),
            )

        if response.errorCode:
            # Tested against the raw engine code, before it is mapped onto the
            # SDK's error enumeration: the mapping is lossy, and several engine
            # codes collapse onto GeneralError.
            if response.errorCode in _FALLBACK_ERROR_CODES:
                return self._evaluate_remotely(
                    flag_key,
                    default_value,
                    evaluation_context,
                    remote_resolver,
                    raw_error_code=response.errorCode,
                    local_error=self._error_for_code(
                        flag_key, response.errorCode, response.errorDetails
                    ),
                )
            raise self._error_for_code(
                flag_key, response.errorCode, response.errorDetails
            )

        value = response.value
        if value is not None and not matches_type(value, expected_type):
            raise TypeMismatchError(
                f"Flag '{flag_key}' returned type {type(value).__name__!r}, "
                f"expected {expected_type}"
            )

        resolved_value = value if value is not None else default_value
        flag_metadata = dict(response.metadata or {})
        # The engine reports the flag version top-level rather than inside its
        # metadata block. Surface it alongside the rest, but never over a key the
        # flag itself defined.
        if response.version and _VERSION_KEY not in flag_metadata:
            flag_metadata[_VERSION_KEY] = response.version
        return FlagResolutionDetails(
            value=resolved_value,
            reason=response.reason or Reason.DEFAULT,
            variant=response.variationType,
            flag_metadata=flag_metadata,
        )

    # ------------------------------------------------------------------
    # Sync resolve methods
    # ------------------------------------------------------------------

    def resolve_boolean_details(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        return self._resolve_generic(
            flag_key,
            default_value,
            bool,
            evaluation_context,
            "resolve_boolean_details",
        )

    def resolve_string_details(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        return self._resolve_generic(
            flag_key,
            default_value,
            str,
            evaluation_context,
            "resolve_string_details",
        )

    def resolve_integer_details(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        return self._resolve_generic(
            flag_key,
            default_value,
            int,
            evaluation_context,
            "resolve_integer_details",
        )

    def resolve_float_details(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        return self._resolve_generic(
            flag_key,
            default_value,
            (float, int),
            evaluation_context,
            "resolve_float_details",
        )

    def resolve_object_details(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[list, dict]]:
        return self._resolve_generic(
            flag_key,
            default_value,
            (dict, list),
            evaluation_context,
            "resolve_object_details",
        )

    # ------------------------------------------------------------------
    # Async resolve methods (delegate to sync counterparts)
    # ------------------------------------------------------------------

    async def resolve_boolean_details_async(
        self,
        flag_key: str,
        default_value: bool,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[bool]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_boolean_details, flag_key, default_value, evaluation_context
        )

    async def resolve_string_details_async(
        self,
        flag_key: str,
        default_value: str,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[str]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_string_details, flag_key, default_value, evaluation_context
        )

    async def resolve_integer_details_async(
        self,
        flag_key: str,
        default_value: int,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[int]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_integer_details, flag_key, default_value, evaluation_context
        )

    async def resolve_float_details_async(
        self,
        flag_key: str,
        default_value: float,
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[float]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_float_details, flag_key, default_value, evaluation_context
        )

    async def resolve_object_details_async(
        self,
        flag_key: str,
        default_value: Union[dict, list],
        evaluation_context: Optional[EvaluationContext] = None,
    ) -> FlagResolutionDetails[Union[dict, list]]:
        """Resolve via WASM (async, runs sync evaluation in thread)."""
        return await asyncio.to_thread(
            self.resolve_object_details, flag_key, default_value, evaluation_context
        )

    # ------------------------------------------------------------------
    # Tracking
    # ------------------------------------------------------------------

    def is_flag_trackable(self, flag_key: str) -> bool:
        with self._lock:
            flags = self._flags
        flag = flags.get(flag_key) if flags is not None else None
        if flag is None:
            logger.warning(
                "Flag with key %s not found when checking if trackable", flag_key
            )
            # we default to trackable for unknown flags to be sure they are visible in the exporters
            return True

        # The engine's own default is true, so a flag whose configuration omits
        # trackEvents is trackable. Defaulting to false here silently drops every
        # event for such a flag.
        track_events = flag.get("trackEvents", True)
        if isinstance(track_events, bool):
            return track_events
        return bool(track_events)
