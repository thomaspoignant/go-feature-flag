"""
Evaluator implementations for GO Feature Flag provider.

Selects remote (relay proxy) or inprocess (local/WASM) evaluation based on options.
"""

from gofeatureflag_python_provider.evaluator.abstract_evaluator import AbstractEvaluator
from gofeatureflag_python_provider.evaluator.inprocess_evaluator import (
    InProcessEvaluator,
)
from gofeatureflag_python_provider.evaluator.remote_evaluator import RemoteEvaluator

__all__ = [
    "AbstractEvaluator",
    "InProcessEvaluator",
    "RemoteEvaluator",
]
