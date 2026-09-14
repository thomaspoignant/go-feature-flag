from __future__ import annotations

import json
import logging
import tempfile
import time
from functools import lru_cache
from pathlib import Path
from typing import Iterator

import pytest
import requests
import yaml

PACKAGE_LOGGER_NAME = "gofeatureflag_python_provider"

_TESTS_DIR = Path(__file__).parent

# The relay proxy image the container tests evaluate against. Tracking latest is
# deliberate: a wire-format change in a new release should surface here rather
# than in a user's application.
RELAY_PROXY_IMAGE = "gofeatureflag/go-feature-flag:latest"
RELAY_PROXY_PORT = 1031
_HEALTH_TIMEOUT_SECONDS = 60
_HEALTH_POLL_INTERVAL_SECONDS = 0.5


@pytest.fixture(autouse=True)
def reset_package_logger_level():
    """Restore the package logger's level between tests.

    GoFeatureFlagProvider.__init__ applies options.log_level to the package-root
    logger only when nothing has configured it yet (level is NOTSET). That check
    makes the level sticky for the rest of the process, so without this reset the
    first provider a test constructs would silently decide the level for every
    test that runs after it.
    """
    logger = logging.getLogger(PACKAGE_LOGGER_NAME)
    previous = logger.level
    logger.setLevel(logging.NOTSET)
    yield
    logger.setLevel(previous)


@lru_cache(maxsize=1)
def docker_is_available() -> bool:
    """Whether a Docker daemon can be reached.

    Container-backed tests skip rather than fail without one: CI always has a
    daemon, a contributor's laptop may not, and `pytest` here runs unfiltered.
    Cached because it is evaluated at collection time by every such module.
    """
    try:
        from testcontainers.core.docker_client import DockerClient

        DockerClient().client.ping()
    except Exception:
        return False
    return True


def _write_relay_proxy_config(directory: Path) -> None:
    """Materialize a relay proxy configuration derived from the mocked fixture.

    mock_responses/config/valid-all-types.json is shaped exactly like a
    /v1/flag/configuration response, so the flag map nested inside it is what the
    file retriever wants. Generating the container's flag file from that fixture
    rather than hand-writing a second copy is the point of this setup: the
    container serves byte-identical definitions to the ones the mocked tests
    replay, so the mock cannot drift away from what the real engine does without
    a container test failing.
    """
    source = json.loads(
        (_TESTS_DIR / "mock_responses" / "config" / "valid-all-types.json").read_text()
    )
    (directory / "flags.json").write_text(json.dumps(source["flags"]))

    config = {
        "server": {"mode": "http", "port": RELAY_PROXY_PORT},
        "pollingInterval": 1000,
        "startWithRetrieverError": False,
        "retriever": {
            "kind": "file",
            "path": "/goff/flags.json",
            "fileFormat": "json",
        },
        "exporter": {"kind": "log"},
        # The relay proxy serves this back on /v1/flag/configuration, so it has to
        # come from the same fixture as the flags for the two to stay consistent.
        "evaluationContextEnrichment": source.get("evaluationContextEnrichment") or {},
    }
    # /goff/ is one of the relay proxy's default configuration locations, so
    # mounting the directory is enough — no command line or env var needed.
    (directory / "goff-proxy.yaml").write_text(yaml.safe_dump(config))


def _wait_until_healthy(endpoint: str, container) -> None:
    """Block until the relay proxy answers /health, or give up with its logs.

    /health lives on the API port here because monitoringPort defaults to 0.
    Polling it mirrors what openfeature/provider_tests/integration_tests.sh does
    and keeps this independent of the testcontainers wait-strategy API.
    """
    deadline = time.monotonic() + _HEALTH_TIMEOUT_SECONDS
    while time.monotonic() < deadline:
        try:
            if requests.get(f"{endpoint}/health", timeout=1).status_code == 200:
                return
        except requests.RequestException:
            pass
        time.sleep(_HEALTH_POLL_INTERVAL_SECONDS)

    logs = container.get_logs()[0].decode("utf-8", errors="replace")
    raise RuntimeError(
        f"relay proxy at {endpoint} was not healthy after "
        f"{_HEALTH_TIMEOUT_SECONDS}s. Container logs:\n{logs}"
    )


@pytest.fixture(scope="session")
def relay_proxy() -> Iterator[str]:
    """A real relay proxy in a container; yields its base URL.

    Session-scoped: booting one costs a few seconds, and every test using it only
    reads flags, so there is no state to isolate between them.
    """
    from testcontainers.core.container import DockerContainer

    with tempfile.TemporaryDirectory() as tmp:
        directory = Path(tmp)
        _write_relay_proxy_config(directory)

        container = (
            DockerContainer(RELAY_PROXY_IMAGE)
            .with_exposed_ports(RELAY_PROXY_PORT)
            .with_volume_mapping(str(directory), "/goff", "ro")
        )
        with container:
            # A random host port, so this never collides with a relay proxy a
            # developer already has running on 1031.
            endpoint = (
                f"http://{container.get_container_host_ip()}:"
                f"{container.get_exposed_port(RELAY_PROXY_PORT)}"
            )
            _wait_until_healthy(endpoint, container)
            yield endpoint
