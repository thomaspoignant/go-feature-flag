import atexit
import time
import pytest
from unittest.mock import patch, MagicMock

from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.options import GoFeatureFlagOptions
from openfeature import api


@pytest.fixture(autouse=True)
def clean_up_providers():
    api.clear_providers()
    yield
    api.clear_providers()


@patch.object(
    GoFeatureFlagProvider,
    "shutdown",
    wraps=GoFeatureFlagProvider.shutdown,
    autospec=True,
)
def test_graceful_exit_runs(mock_shutdown: MagicMock):
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(endpoint="https://gofeatureflag.org/"),
    )
    api.set_provider(goff_provider)

    atexit._run_exitfuncs()
    mock_shutdown.assert_called_once()


@patch.object(
    GoFeatureFlagProvider,
    "shutdown",
    wraps=GoFeatureFlagProvider.shutdown,
    autospec=True,
)
def test_graceful_exit_skipped_without_openfeature_api(mock_shutdown: MagicMock):
    _ = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(endpoint="https://gofeatureflag.org/"),
    )

    atexit._run_exitfuncs()
    mock_shutdown.assert_not_called()


@patch.object(
    GoFeatureFlagProvider,
    "shutdown",
    wraps=GoFeatureFlagProvider.shutdown,
    autospec=True,
)
def test_both_graceful_exit_and_manual_cleanup(mock_shutdown: MagicMock):
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(endpoint="https://gofeatureflag.org/"),
    )
    api.set_provider(goff_provider)

    api.shutdown()
    mock_shutdown.assert_called_once()

    atexit._run_exitfuncs()
    mock_shutdown.assert_called()
    assert mock_shutdown.call_count == 2


@patch.object(
    GoFeatureFlagProvider,
    "shutdown",
    wraps=GoFeatureFlagProvider.shutdown,
    autospec=True,
)
def test_graceful_exit_interrupts_polling_cycle(mock_shutdown: MagicMock):
    goff_provider = GoFeatureFlagProvider(
        options=GoFeatureFlagOptions(
            endpoint="https://gofeatureflag.org/",
            data_flush_interval=1000000,  # extremely long to ensure polling is in wait state
        ),
    )
    api.set_provider(goff_provider)

    start_time = time.time()

    atexit._run_exitfuncs()
    mock_shutdown.assert_called_once()

    elapsed_time = time.time() - start_time
    assert elapsed_time < 1, "Graceful exit took too long"
