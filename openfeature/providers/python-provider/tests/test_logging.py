"""
Tests for the provider's logging contract.

Every module logs through ``logging.getLogger(__name__)``, so all records land
under the ``gofeatureflag_python_provider`` logger and stay attributable to the
module that emitted them. The provider sets that logger's level from
``options.log_level``, but only when the embedding application has not configured
it — the level is process-wide state a library must not seize.
"""

from __future__ import annotations

import importlib
import logging
import pkgutil

import gofeatureflag_python_provider
from gofeatureflag_python_provider.options import EvaluationType, GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider

PACKAGE_LOGGER_NAME = "gofeatureflag_python_provider"


def _options(**kwargs):
    return GoFeatureFlagOptions(
        endpoint="http://localhost:1031",
        evaluation_type=EvaluationType.REMOTE,
        **kwargs,
    )


def _module_level_loggers():
    """Every module-level Logger in the package, as (module name, logger) pairs."""
    found = []
    for module_info in pkgutil.walk_packages(
        gofeatureflag_python_provider.__path__,
        prefix=f"{PACKAGE_LOGGER_NAME}.",
    ):
        module = importlib.import_module(module_info.name)
        for attribute in vars(module).values():
            if isinstance(attribute, logging.Logger):
                found.append((module.__name__, attribute))
    return found


def test_every_module_logger_is_named_after_its_module():
    """No module may hardcode a logger name.

    services/api.py used to call getLogger("gofeatureflag_python_provider"),
    which emitted records indistinguishable from the package root and from every
    other module in %(name)s.
    """
    loggers = _module_level_loggers()
    assert loggers, "expected the package to define at least one module logger"

    mismatched = [
        (module_name, logger.name)
        for module_name, logger in loggers
        if logger.name != module_name
    ]
    assert mismatched == []


def test_every_module_logger_lives_under_the_package_logger():
    """The provider sets the level on the package root and relies on inheritance."""
    for module_name, logger in _module_level_loggers():
        assert logger.name.startswith(f"{PACKAGE_LOGGER_NAME}."), module_name


def test_package_logger_has_a_null_handler():
    """A library must not make Python fall back to 'no handlers could be found'."""
    handlers = logging.getLogger(PACKAGE_LOGGER_NAME).handlers
    assert any(isinstance(handler, logging.NullHandler) for handler in handlers)


def test_log_level_is_applied_when_the_application_configured_nothing():
    GoFeatureFlagProvider(options=_options(log_level="DEBUG"))

    assert logging.getLogger(PACKAGE_LOGGER_NAME).level == logging.DEBUG


def test_log_level_does_not_override_an_application_configured_level():
    logging.getLogger(PACKAGE_LOGGER_NAME).setLevel(logging.ERROR)

    GoFeatureFlagProvider(options=_options(log_level="DEBUG"))

    assert logging.getLogger(PACKAGE_LOGGER_NAME).level == logging.ERROR


def test_first_provider_wins_when_two_disagree():
    """The level is process-wide, so it cannot follow whichever provider was built last."""
    GoFeatureFlagProvider(options=_options(log_level="DEBUG"))
    GoFeatureFlagProvider(options=_options(log_level="ERROR"))

    assert logging.getLogger(PACKAGE_LOGGER_NAME).level == logging.DEBUG


def test_get_log_level_int_accepts_an_int():
    assert _options(log_level=logging.DEBUG).get_log_level_int() == logging.DEBUG


def test_get_log_level_int_accepts_a_level_name_in_any_case():
    assert _options(log_level="debug").get_log_level_int() == logging.DEBUG
    assert _options(log_level="INFO").get_log_level_int() == logging.INFO
    assert _options(log_level="Error").get_log_level_int() == logging.ERROR


def test_get_log_level_int_defaults_to_warning():
    assert _options().get_log_level_int() == logging.WARNING
    assert _options(log_level="not-a-level").get_log_level_int() == logging.WARNING
