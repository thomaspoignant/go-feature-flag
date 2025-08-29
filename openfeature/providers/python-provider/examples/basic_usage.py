#!/usr/bin/env python3
"""
Basic usage example for the GoFeatureFlag provider.
"""

import asyncio
import logging
from openfeature import api
from gofeatureflag_python_provider import (
    GoFeatureFlagProvider,
    GoFeatureFlagOptions,
    EvaluationType,
)


async def main():
    """Main function demonstrating basic usage."""

    # Set up logging
    logging.basicConfig(level=logging.INFO)
    logger = logging.getLogger(__name__)

    # Create provider options
    options = GoFeatureFlagOptions(
        endpoint="https://your-go-feature-flag-relay.com",
        evaluation_type=EvaluationType.REMOTE,  # Use remote evaluation
        timeout=5000,  # 5 seconds timeout
        api_key="your-api-key",  # Optional: if you have authentication
    )

    # Create the provider
    provider = GoFeatureFlagProvider(options, logger)

    # Initialize the provider
    await provider.initialize()

    try:
        # Set the provider in the OpenFeature API
        api.set_provider(provider)

        # Create a client
        client = api.get_client()

        # Example: Evaluate a boolean flag
        boolean_result = client.get_boolean_details("my-feature-flag", False)
        print(f"Boolean flag result: {boolean_result.value}")
        print(f"Reason: {boolean_result.reason}")

        # Example: Evaluate a string flag
        string_result = client.get_string_details("my-string-flag", "default-value")
        print(f"String flag result: {string_result.value}")
        print(f"Reason: {string_result.reason}")

        # Example: Evaluate a number flag
        number_result = client.get_float_details("my-number-flag", 42.0)
        print(f"Number flag result: {number_result.value}")
        print(f"Reason: {number_result.reason}")

        # Example: Track a custom event
        provider.track("user_action", None, {"action": "button_click"})

    finally:
        # Clean up
        provider.shutdown()


if __name__ == "__main__":
    asyncio.run(main())
