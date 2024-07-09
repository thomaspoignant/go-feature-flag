# OpenFeature Provider for iOS

This example is a simple iOS application that demonstrates how to use the OpenFeature Swift SDK to integrate 
OpenFeature into your iOS application.

### How to use this example?

1. Start your relay proxy with the flag configuration in the folder (`flags.goff.yaml`).
2. Replace the `endpoint` in the `openfeature_swift_iosApp.swift` file with your OpenFeature endpoint
_(replace `https://your_domain.io` by the endpoint of your relayproxy)_.
3. Start the iOS application in Xcode.  
4. By clicking on the **Check flag** button, you will call the OpenFeature API to get the flag value.
5. The switch button will be set to on/off dependending on the flag `my-flag` value.

## Want to use the OpenFeature SDK with GO Feature Flag in your iOS application?

Check the [`go-feature-flag/openfeature-swift-provider`](https://github.com/go-feature-flag/openfeature-swift-provider/) repository.

## Demo
https://github.com/thomaspoignant/go-feature-flag/assets/17908063/82b7f946-d501-4e28-9c0a-68e23475ce7d
