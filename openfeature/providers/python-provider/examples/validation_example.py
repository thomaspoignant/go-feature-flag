#!/usr/bin/env python3
"""
Example demonstrating Pydantic validation in the GO Feature Flag provider.
"""

from gofeatureflag_python_provider.provider_options import GoFeatureFlagOptions
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider
from gofeatureflag_python_provider.exception import InvalidOptionsException
from gofeatureflag_python_provider.model import EvaluationType


def demonstrate_validation():
    """Demonstrate various validation scenarios."""

    print("🦊 GO Feature Flag Provider - Pydantic Validation Examples\n")

    # Example 1: Valid options
    print("✅ Example 1: Valid options")
    try:
        valid_options = GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            evaluation_type=EvaluationType.REMOTE,
            api_key="my-api-key",
            data_flush_interval=30000,
        )
        print(f"   Created valid options: {valid_options}")

        # Create provider with valid options
        provider = GoFeatureFlagProvider(valid_options)
        print("   Provider created successfully!")

    except Exception as e:
        print(f"   Error: {e}")

    print()

    # Example 2: Invalid endpoint
    print("❌ Example 2: Invalid endpoint")
    try:
        invalid_options = GoFeatureFlagOptions(
            endpoint="not-a-valid-url", evaluation_type=EvaluationType.IN_PROCESS
        )
        print(f"   Created options: {invalid_options}")

    except Exception as e:
        print(f"   Validation error: {e}")

    print()

    # Example 3: Invalid numeric values
    print("❌ Example 3: Invalid numeric values")
    try:
        invalid_options = GoFeatureFlagOptions(
            endpoint="http://localhost:1031",
            data_flush_interval=-1000,  # Negative value
            max_pending_events=0,  # Zero value
        )
        print(f"   Created options: {invalid_options}")

    except Exception as e:
        print(f"   Validation error: {e}")

    print()

    # Example 4: Invalid API key
    print("❌ Example 4: Invalid API key")
    try:
        invalid_options = GoFeatureFlagOptions(
            endpoint="http://localhost:1031", api_key=""  # Empty string
        )
        print(f"   Created options: {invalid_options}")

    except Exception as e:
        print(f"   Validation error: {e}")

    print()

    # Example 5: Creating provider with dict options
    print("✅ Example 5: Creating provider with dict options")
    try:
        options_dict = {
            "endpoint": "http://localhost:1031",
            "evaluation_type": EvaluationType.REMOTE,
            "api_key": "my-api-key",
        }

        provider = GoFeatureFlagProvider(options_dict)
        print("   Provider created successfully from dict!")

    except Exception as e:
        print(f"   Error: {e}")

    print()

    # Example 6: Invalid options type
    print("❌ Example 6: Invalid options type")
    try:
        provider = GoFeatureFlagProvider("invalid options")
        print("   Provider created successfully!")

    except Exception as e:
        print(f"   Error: {e}")


if __name__ == "__main__":
    demonstrate_validation()
