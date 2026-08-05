import logging

import pytest

PACKAGE_LOGGER_NAME = "gofeatureflag_python_provider"


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
