"""Service-level exceptions for the GO Feature Flag API (mirroring JS SDK)."""


class GoFeatureFlagServiceError(Exception):
    """Base exception for GO Feature Flag API service errors."""

    pass


class FlagConfigurationUnavailableError(GoFeatureFlagServiceError):
    """Flag configuration endpoint is not reachable or returned an error."""

    pass


class DataCollectorError(GoFeatureFlagServiceError):
    """Failed to send events to the data collector."""

    pass


class UnauthorizedError(GoFeatureFlagServiceError):
    """Authentication or authorization failed (401/403)."""

    pass
