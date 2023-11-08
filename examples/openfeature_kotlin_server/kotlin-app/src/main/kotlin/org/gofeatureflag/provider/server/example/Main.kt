package org.gofeatureflag.provider.server.example

import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProvider
import dev.openfeature.contrib.providers.gofeatureflag.GoFeatureFlagProviderOptions
import dev.openfeature.sdk.EvaluationContext
import dev.openfeature.sdk.MutableContext
import dev.openfeature.sdk.OpenFeatureAPI


fun main() {
    val options = GoFeatureFlagProviderOptions.builder().endpoint("http://gofeatureflag:1031/").build()
    val provider = GoFeatureFlagProvider(options)

    OpenFeatureAPI.getInstance().provider = provider
    // wait for the provider to be ready
    val api = OpenFeatureAPI.getInstance()
    val featureFlagClient = api.client


    var it = 0
    while (true) {
        val ctx = evaluationContexts[it % evaluationContexts.size]
        val flag = "my-new-feature"
        val myNewFeature = featureFlagClient.getBooleanValue(flag, false, ctx)
        if (myNewFeature) {
            // the new feature is available
            println("✅ ${flag} is available for id ${ctx.targetingKey}");
        } else {
            // apply the old feature
            println("❌ ${flag} is not available for id ${ctx.targetingKey}");
        }

        Thread.sleep(1000)
        it++
    }
}

val evaluationContexts: Array<EvaluationContext> = arrayOf(
    MutableContext("1d1b9238-2591-4a47-94cf-d2bc080892f1").add("email", "user1@gofeatureflag.org"),
    MutableContext("fa0f8cfa-02a8-4424-b201-f4dca70a3819").add("email", "user2@gofeatureflag.org"),
    MutableContext("401ee3dd-81f0-4f49-9bf0-a95eb1f1d0d6").add("email", "user3@gofeatureflag.org"),
    MutableContext("9799dce2-9621-4137-8a95-6033cdeeddc5").add("email", "user4@gofeatureflag.org"),
    MutableContext("628836b5-8d64-4ba5-8043-11e1cf811e58").add("email", "user5@gofeatureflag.org"),
    MutableContext("b3576da1-1b5b-4b94-b98c-ac3eb92dd53f").add("email", "user6@gofeatureflag.org"),
)
