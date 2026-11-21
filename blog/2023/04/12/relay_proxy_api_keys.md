# Limiting Access to the relay proxy with API Keys

April 12, 2023 ·

<!-- -->

3 min read

[![Thomas Poignant](https://github.com/thomaspoignant.png)](https://github.com/thomaspoignant)

[Thomas Poignant](https://github.com/thomaspoignant)

Creator of GO Feature Flag

![security](/assets/images/20945235-2b2db0a23e427482e0e64519f65a3400.jpg)

GO Feature Flag Relay Proxy is a powerful tool that enables you to evaluate your feature flags without deploying code changes. It allows you to test and roll out new features to your users gradually, and to monitor the performance of your code in real-time.

With the recent release of version `v1.7.0`, GO Feature Flag Relay Proxy introduces a new feature that enhances the security of your feature flag evaluations. This new feature allows you to control who can access your relay proxy by configuring API Keys.

## Limiting Access with API Keys[​](#limiting-access-with-api-keys "Direct link to Limiting Access with API Keys")

With the release of GO Feature Flag Relay Proxy version `v1.7.0`, you can now limit access to the relay proxy by configuring API Keys. This feature allows you to restrict access to only known clients, adding an additional layer of security to your feature flag evaluations.

To enable this feature, you need to add a list of API Keys to your configuration. Only requests with these API Keys will be accepted. This means that any requests without a valid API Key will be rejected.

In your **relay proxy** configuration file it will look like this:

```
# ...

# List of authorized API keys.

# Each request will need to provide one of authorized key inside a X-API-Key header with format` X-API-Key: <api-key>`.

# There will be no authorization when this config is not configured.

apiKeys:

  - 6147795c-4438-44f1-8ab2-18cb3fa9591f # apiKey for client 1

  - 3e3fb534-c8fd-42e9-8f2e-c80276b1685c # apiKey for client 2

  - 6603f9f8-2871-4149-81e8-8a362a8ae266 # apiKey for client 3

  - 0fc30ffe-b9bc-4398-a9f1-6f8b8486e8ce # apiKey for client 4

  - 500d222e-2654-41a4-bd31-e24a629cb78a # apiKey for client 5



# ...
```

## Using Client API Keys with Openfeature Providers[​](#using-client-api-keys-with-openfeature-providers "Direct link to Using Client API Keys with Openfeature Providers")

If you are using Open Feature and our providers, we have updated them to handle the `apiKey` for you. You can now specify your `apiKey` directly when you are initializing your provider. This will authenticate every request to the relay proxy with this apiKey.

To use this feature, you need to specify the client API Key in your Open Feature provider initialization. This can be done by adding the `apiKey` like this:

* GO
* .NET
* JAVA
* JS / TS

```
provider, err := gofeatureflag.NewProvider(gofeatureflag.ProviderOptions{

  //...

  APIKey:   "6147795c-4438-44f1-8ab2-18cb3fa9591f",

})
```

```
GoFeatureFlagProviderOptions options = new GoFeatureFlagProviderOptions{

    // ...

    ApiKey = "6147795c-4438-44f1-8ab2-18cb3fa9591f"

  }



GoFeatureFlagProvider authenticatedProvider = new GoFeatureFlagProvider(options);
```

```
GoFeatureFlagProviderOptions options =

      GoFeatureFlagProviderOptions.builder()

      .apiKey("6147795c-4438-44f1-8ab2-18cb3fa9591f")

      .endpoint(relayProxyAuthenticatedEndpoint)

      .build();



GoFeatureFlagProvider provider = new GoFeatureFlagProvider(options);
```

```
const goFeatureFlagProvider = new GoFeatureFlagProvider({

  // ...

  apiKey: '6147795c-4438-44f1-8ab2-18cb3fa9591f'

});
```

## Conclusion[​](#conclusion "Direct link to Conclusion")

The new feature of GO Feature Flag Relay Proxy 1.7.0 allows you to limit access to your relay proxy by configuring API Keys.

This adds an extra layer of security to your feature flag evaluations, ensuring that only known clients can access your relay proxy.

This feature is easy to configure and can be set up in a matter of minutes, giving you peace of mind knowing that your feature flags are secure.

**Tags:**

* [feature flag](https://gofeatureflag.org/blog/tags/feature-flag)
* [relay-proxy](https://gofeatureflag.org/blog/tags/relay-proxy)

[Edit this page](https://github.com/thomaspoignant/go-feature-flag/tree/main/website/blog/2023-04-12-relay_proxy_api_keys/index.mdx)
