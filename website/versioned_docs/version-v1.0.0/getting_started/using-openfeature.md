---
sidebar_position: 20
description: Deploy the relay proxy and use the OpenFeature SDKs 
---
# Using Open Feature SDKs

:::note
OpenFeature provides a shared, standardized feature flagging client - _an SDK_ - which can be plugged into various 3rd-party feature flagging providers.
Whether you're using an open-source system or a commercial product, whether it's self-hosted or cloud-hosted, OpenFeature provides a consistent, unified API for developers to use feature flagging in their applications.  
_[Documentation](https://docs.openfeature.dev)_
:::

GO Feature Flag believe in **OpenSource** and **standardization**, this is the reason why we decided not implementing any custom SDK and rely only on **Open Feature**.

To be compatible with Open Feature, **GO Feature Flag** is providing a lightweight self-hosted API server *(called [relay proxy](../category/use-the-relay-proxy))* that is using the GO Feature Flag module internally.  
When the **relay proxy** is running in your infrastructure, you can use the **Open Feature SDKs** in combination with **GO Feature Flag providers** to evaluate your flags. 

This schema is an overview on how **Open Feature** is working, you can have more information about all the concepts in the **[Open Feature documentation](https://docs.openfeature.dev)**.
![](/docs/openfeature/concepts.png)

## Create a feature flag configuration

Create a new `YAML` file containing your first flag configuration.

```yaml title="flag-config.yaml"
# 20% of the users will use the variation "my-new-feature"
test-flag:
  variations:
    my-new-feature: true
    my-old-feature: false
  defaultRule:
    percentage:
      my-new-feature: 20
      my-old-feature: 80
```

This flag split the usage of this flag, 20% will use the variation `my-new-feature` and 80% the variation `my-old-feature`.

## Create a relay proxy configuration file

Create a new `YAML` file containing the configuration of your relay proxy.

```yaml title="goff-proxy.yaml"
server:
  mode: http
  port: 1031
pollingInterval: 1000
startWithRetrieverError: false
retriever:
  kind: file
  path: /goff/flag-config.yaml
exporter:
  kind: log
```

## Install the relay proxy

And we will run the **relay proxy** locally to make the API available.  
The default port will be `1031`.

```shell
# Launch the container
docker run \
  -p 1031:1031 \
  -v $(pwd)/flag-config.yaml:/goff/flag-config.yaml \
  -v $(pwd)/goff-proxy.yaml:/goff/goff-proxy.yaml \
  thomaspoignant/go-feature-flag-relay-proxy:latest

```

_If you don't want to use docker to install the **relay proxy** you can follow the [documentation](../relay_proxy/install_relay_proxy.md)_.

## Use Open Feature SDK

_In this example we are using the javascript SDK, but it is still relevant for all the languages_.

### Install dependencies

```shell
npm i @openfeature/server-sdk @openfeature/go-feature-flag-provider
```

### Init your Open Feature client

In your app initialization your have to create a client using the Open Feature SDK and initialize it.

```javascript
const {OpenFeature} = require("@openfeature/server-sdk");
const {GoFeatureFlagProvider} = require("@openfeature/go-feature-flag-provider");


// init Open Feature SDK with GO Feature Flag provider
const goFeatureFlagProvider = new GoFeatureFlagProvider({
  endpoint: 'http://localhost:1031/' // DNS of your instance of relay proxy
});
OpenFeature.setProvider(goFeatureFlagProvider);
const featureFlagClient = OpenFeature.getClient('my-app')
```

### Evaluate your flag

Now you can evaluate your flags anywhere in your code using this client.

```javascript
// Context of your flag evaluation.
// With GO Feature Flag you MUST have a targetingKey that is a unique identifier of the user.
const userContext = {
  targetingKey: '1d1b9238-2591-4a47-94cf-d2bc080892f1', // user unique identifier (mandatory)
  firstname: 'john',
  lastname: 'doe',
  email: 'john.doe@gofeatureflag.org',
  admin: true, // this field is used in the targeting rule of the flag "flag-only-for-admin"
  // ...
};

const adminFlag = await featureFlagClient.getBooleanValue('flag-only-for-admin', false, userContext);
if (adminFlag) {
   // flag "flag-only-for-admin" is true for the user
  console.log("new feature");
} else {
  // flag "flag-only-for-admin" is false for the user
}
```
