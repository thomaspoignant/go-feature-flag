"""
Shared construction of the OFREP client used for remote evaluation.

Both the remote evaluator and the in-process evaluator's remote fallback build
their client here, so authentication, headers and timeout are identical on both
paths by construction rather than duplicated and left to drift apart.
"""

from typing import Callable, Optional

from openfeature.contrib.provider.ofrep import OFREPProvider

from gofeatureflag_python_provider.options import GoFeatureFlagOptions

# Used when no timeout is configured. Matches the flag-configuration client.
DEFAULT_TIMEOUT_SECONDS = 10.0


def build_ofrep_provider(options: GoFeatureFlagOptions):
    """
    Build an OFREPProvider configured from the provider options.

    The timeout is passed explicitly: left unset, the OFREP client applies its
    own default rather than the one the caller configured.

    :param options: Provider options (endpoint, optional api_key, timeout).
    :return: A configured OFREPProvider.
    """
    headers_factory: Optional[Callable[[], dict[str, str]]] = None
    if options.api_key or options.custom_headers:
        api_key = options.api_key
        # Custom headers first, so a configured api_key always wins over a
        # custom Authorization header rather than being silently dropped.
        custom_headers = dict(options.custom_headers or {})

        def _headers() -> dict[str, str]:
            headers = dict(custom_headers)
            headers["Content-Type"] = "application/json"
            if api_key:
                headers["X-API-Key"] = api_key
            return headers

        headers_factory = _headers

    timeout_seconds = (
        options.timeout / 1000.0
        if options.timeout is not None
        else DEFAULT_TIMEOUT_SECONDS
    )
    return OFREPProvider(
        base_url=str(options.endpoint).rstrip("/"),
        headers_factory=headers_factory,
        timeout=timeout_seconds,
    )
