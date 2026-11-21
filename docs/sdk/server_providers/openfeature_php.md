# PHP

[![Packagist - Version](https://img.shields.io/packagist/v/open-feature/go-feature-flag-provider?logo=php\&color=blue\&style=flat-square)](https://packagist.org/packages/open-feature/go-feature-flag-provider) [![Packagist - Downloads](https://img.shields.io/packagist/dt/open-feature/go-feature-flag-provider?logo=php\&style=flat-square)](https://packagist.org/packages/open-feature/go-feature-flag-provider)

In conjunction with the [OpenFeature SDK](https://openfeature.dev/docs/reference/concepts/provider) you will be able to evaluate your feature flags in your PHP applications.

### Functionalities:[​](#functionalities "Direct link to Functionalities:")

* Manage the integration of the OpenFeature PHP SDK and GO Feature Flag relay-proxy.

## Dependency Setup[​](#dependency-setup "Direct link to Dependency Setup")

### Composer[​](#composer "Direct link to Composer")

```
composer require open-feature/go-feature-flag-provider
```

## Getting started[​](#getting-started "Direct link to Getting started")

### Initialize the provider[​](#initialize-the-provider "Direct link to Initialize the provider")

The `GoFeatureFlagProvider` takes a config object as parameter to be initialized.

The constructor of the config object has the following options:

| **Option**      | **Description**                                                                                                  |
| --------------- | ---------------------------------------------------------------------------------------------------------------- |
| `endpoint`      | **(mandatory)** The URL to access to the relay-proxy.<br />*(example: `https://relay.proxy.gofeatureflag.org/`)* |
| `apiKey`        | The token used to call the relay proxy.                                                                          |
| `customHeaders` | Any headers you want to add to call the relay-proxy.                                                             |
| `httpclient`    | The HTTP Client to use (if you want to use a custom one). *It has to be a `PSR-7` compliant implementation.*     |

The only required option to create a `GoFeatureFlagProvider` is the URL *(`endpoint`)* to your GO Feature Flag relay-proxy instance.

```
use OpenFeature\Providers\GoFeatureFlag\config\Config;

use OpenFeature\Providers\GoFeatureFlag\GoFeatureFlagProvider;

use OpenFeature\implementation\flags\MutableEvaluationContext;

use OpenFeature\implementation\flags\Attributes;

use OpenFeature\OpenFeatureAPI;



$config = new Config('https://gofeatureflag.org', 'my-api-key');

$provider = new GoFeatureFlagProvider($config);



$api = OpenFeatureAPI::getInstance();

$api->setProvider($provider);

$client = $api->getClient();

$evaluationContext = new MutableEvaluationContext(

      "214b796a-807b-4697-b3a3-42de0ec10a37", 

      new Attributes(["email" => 'contact@gofeatureflag.org'])

  );



$value = $client->getBooleanDetails('my-flag-key', false, $evaluationContext);

if ($value) {

    echo "The flag is enabled";

} else {

    echo "The flag is disabled";

}
```

The evaluation context is the way for the client to specify contextual data that GO Feature Flag uses to evaluate the feature flags, it allows to define rules on the flag.

The `targeting_key` is mandatory for GO Feature Flag to evaluate the feature flag, it could be the id of a user, a session ID or anything you find relevant to use as identifier during the evaluation.

### Evaluate a feature flag[​](#evaluate-a-feature-flag "Direct link to Evaluate a feature flag")

The client is used to retrieve values for the current `EvaluationContext`. For example, retrieving a boolean value for the flag **"my-flag"**:

```
$value = $client->getBooleanDetails('integer_key', false, $evaluationContext);

if ($value) {

  echo "The flag is enabled";

} else {

  echo "The flag is disabled";

}
```

GO Feature Flag supports different all OpenFeature supported types of feature flags, it means that you can use all the accessor directly

```
// Bool

$client->getBooleanDetails('my-flag-key', false, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));

$client->getBooleanValue('my-flag-key', false, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));



// String

$client->getStringDetails('my-flag-key', "default", new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));

$client->getStringValue('my-flag-key', "default", new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));



// Integer

$client->getIntegerDetails('my-flag-key', 1, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));

$client->getIntegerValue('my-flag-key', 1, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));



// Float

$client->getFloatDetails('my-flag-key', 1.1, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));

$client->getFloatValue('my-flag-key', 1.1, new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));



// Object

$client->getObjectDetails('my-flag-key', ["default" => true], new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));

$client->getObjectValue('my-flag-key', ["default" => true], new MutableEvaluationContext("214b796a-807b-4697-b3a3-42de0ec10a37"));
```

## Contributing[​](#contributing "Direct link to Contributing")

This project welcomes contributions from the community. If you're interested in contributing, see the [contributors' guide](https://github.com/thomaspoignant/go-feature-flag/blob/main/CONTRIBUTING.md) for some helpful tips.

### PHP Versioning[​](#php-versioning "Direct link to PHP Versioning")

This library targets PHP version 8.0 and newer. As long as you have any compatible version of PHP on your system you should be able to utilize the OpenFeature SDK.

This package also has a .tool-versions file for use with PHP version managers like asdf.

### Installation and Dependencies[​](#installation-and-dependencies "Direct link to Installation and Dependencies")

Install dependencies with `composer install`, it will update the `composer.lock` with the most recent compatible versions.

We value having as few runtime dependencies as possible. The addition of any dependencies requires careful consideration and review.

## Features status[​](#features-status "Direct link to Features status")

| Status | Feature                      | Description                                                                                                                                                                |
| ------ | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **     | Remote Evaluation            | The provider is calling the remote server to evaluate the feature flags.                                                                                                   |
| **     | Tracking Flag Evaluation     | The provider is tracking all the evaluations of your feature flags and you can export them using an exporter.                                                              |
| **     | In process Evaluation        | The provider is able to evaluate the feature flags in process, it loads the configuration from the relay-proxy and perform evaluation inside the SDK.                      |
| **     | Tracking Custom Events       | The provider is tracking custom events through the track() function of your SDK. All those events are send to the exporter for you to forward them where you want.         |
| **     | Configuration Change Updates | The provider is able to update the configuration based on the configuration, it means that the provider is able to react to any feature flag change on your configuration. |
| **     | Provider Events Reactions    | You can add an event handler to the provider to react to the provider events.                                                                                              |

**Implemented**In-progress**Not implemented yet
