"""Test fixtures for the GO Feature Flag Python provider tests."""

import pytest
import requests


def is_responsive(url):
    """Check if GOFF is responsive."""
    try:
        response = requests.get(url=url, timeout=10000)
        if response.status_code == 200:
            return True
    except requests.exceptions.ConnectionError:
        return False


@pytest.fixture(scope="session")
def goff(docker_ip, docker_services):
    """Ensure that HTTP service is up and responsive."""

    # `port_for` takes a container port and returns the corresponding host port
    port = docker_services.port_for("goff", 1031)
    url = f"http://{docker_ip}:{port}"
    docker_services.wait_until_responsive(
        timeout=30.0,
        pause=0.1,
        check=lambda: is_responsive(f"{url}/health"),
    )
    yield url
