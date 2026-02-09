package dev.openfeature.kotlin.contrib.providers.ofrep.payloads

internal const val VALID_1_FLAG_IN_PARSE_ERROR_PAYLOAD = """
{
  "flags": [
    {
      "value": true,
      "key": "my-flag",
      "reason": "STATIC",
      "variant": "variantA",
      "metadata": {
        "additionalProp1": true,
        "additionalProp2": true,
        "additionalProp3": true
      }
    },
    {
      "key": "my-other-flag",
      "errorCode": "PARSE_ERROR",
      "errorDetails": "Error details about PARSE_ERROR"
    }
  ]
}
"""
