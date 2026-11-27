# Python

[![PyPI - Version](https://img.shields.io/pypi/v/gofeatureflag-python-provider?color=blue\&style=flat-square)](https://pypi.org/project/gofeatureflag-python-provider/) ![PyPI - Downloads](https://img.shields.io/pypi/dm/gofeatureflag-python-provider?style=flat-square)

The Python provider for GO Feature Flag allows you to evaluate feature flags in your Python application. It supports two evaluation modes:

* **In-Process** *(default)*: flag configuration is fetched from the relay proxy and cached locally; flags are evaluated using a bundled WASM module with no per-evaluation network call.
* **Remote**: each flag evaluation calls the GO Feature Flag relay proxy using the OFREP API.

## Install dependencies[​](#install-dependencies "Direct link to Install dependencies")

The first thing we will do is install the **Open Feature SDK** and the **GO Feature Flag provider**.

```
pip install gofeatureflag-python-provider
```

## Initialize your Open Feature client[​](#initialize-your-open-feature-client "Direct link to Initialize your Open Feature client")

To evaluate the flags you need to have an Open Feature configured in your app.

* In-Process (default)
* Remote

In-Process evaluation fetches the flag configuration at startup and refreshes it on a configurable polling interval. Flags are evaluated locally using a bundled WASM module — no network call is made per evaluation.

```
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider

from gofeatureflag_python_provider.options import GoFeatureFlagOptions, EvaluationType

from openfeature import api

from openfeature.evaluation_context import EvaluationContext



goff_provider = GoFeatureFlagProvider(

    options=GoFeatureFlagOptions(

        endpoint="https://gofeatureflag.org/",

        evaluation_type=EvaluationType.INPROCESS,  # default

    )

)

api.set_provider(goff_provider)

client = api.get_client(domain="test-client")
```

Remote evaluation sends each flag evaluation as an HTTP request to the relay proxy.

```
from gofeatureflag_python_provider.provider import GoFeatureFlagProvider

from gofeatureflag_python_provider.options import GoFeatureFlagOptions, EvaluationType

from openfeature import api

from openfeature.evaluation_context import EvaluationContext



goff_provider = GoFeatureFlagProvider(

    options=GoFeatureFlagOptions(

        endpoint="https://gofeatureflag.org/",

        evaluation_type=EvaluationType.REMOTE,

    )

)

api.set_provider(goff_provider)

client = api.get_client(domain="test-client")
```

## Evaluate your flag[​](#evaluate-your-flag "Direct link to Evaluate your flag")

This code block explains how you can create an `EvaluationContext` and use it to evaluate your flag.

note

In this example, we are evaluating a `boolean` flag, but other types are also available.

**Refer to the [Open Feature documentation](https://docs.openfeature.dev/docs/reference/concepts/evaluation-api#basic-evaluation) to know more about it.**

```
# Context of your flag evaluation.

# With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.

evaluation_ctx = EvaluationContext(

    targeting_key="d45e303a-38c2-11ed-a261-0242ac120002",

    attributes={

        "email": "john.doe@gofeatureflag.org",

        "firstname": "john",

        "lastname": "doe",

        "anonymous": False,

        "professional": True,

        "rate": 3.14,

        "age": 30,

        "company_info": {"name": "my_company", "size": 120},

        "labels": ["pro", "beta"],

    },

)



admin_flag = client.get_boolean_value(

    flag_key="flag-only-for-admin",

    default_value=False,

    evaluation_context=evaluation_ctx,

)



if admin_flag:

    # flag "flag-only-for-admin" is true for the user

    pass

else:

    # flag "flag-only-for-admin" is false for the user

    pass
```

## Configure the provider[​](#configure-the-provider "Direct link to Configure the provider")

You can configure the provider with several options to customize its behavior:

| Option                                  | Mandatory | Default        | Description                                                                                      |
| --------------------------------------- | --------- | -------------- | ------------------------------------------------------------------------------------------------ |
| **`endpoint`**                          | `true`    | —              | URL of the GO Feature Flag relay proxy *(e.g. `http://localhost:1031`)*                          |
| **`evaluation_type`**                   | `false`   | `INPROCESS`    | Evaluation mode: `EvaluationType.INPROCESS` or `EvaluationType.REMOTE`                           |
| **`api_key`**                           | `false`   | `None`         | API key for authenticated relay proxy requests                                                   |
| **`flag_config_poll_interval_seconds`** | `false`   | `120`          | Polling interval (seconds) to refresh flag configuration *(in-process only)*                     |
| **`wasm_file_path`**                    | `false`   | `None`         | Path to a custom WASM/WASI evaluation binary *(in-process only; uses bundled binary by default)* |
| **`wasm_pool_size`**                    | `false`   | CPU core count | Pool size for concurrent WASM evaluation instances *(in-process only)*                           |
| **`data_flush_interval`**               | `false`   | `60000`        | Interval (ms) to flush usage data to the relay proxy                                             |
| **`max_pending_events`**                | `false`   | `10000`        | Maximum buffered events before a forced flush                                                    |
| **`disable_data_collection`**           | `false`   | `False`        | Set to `True` to disable usage analytics                                                         |
| **`exporter_metadata`**                 | `false`   | `{}`           | Static metadata attached to evaluation events                                                    |
| **`log_level`**                         | `false`   | `"WARNING"`    | Logging level: `"DEBUG"`, `"INFO"`, `"WARNING"`, or `"ERROR"`                                    |
| **`urllib3_pool_manager`**              | `false`   | `None`         | Custom `urllib3.PoolManager` HTTP client                                                         |

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | In process Evaluation        | The provider is able to evaluate the feature flags in process, it loads the configuration from the relay-proxy and perform evaluation inside the SDK.                      |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |

**Implemented**In-progress**Not implemented yet

## Contribute to the provider[​](#contribute-to-the-provider "Direct link to Contribute to the provider")

You can find the source of the provider in the [`repository`](https://github.com/thomaspoignant/go-feature-flag/tree/main/openfeature/providers/python-provider).
