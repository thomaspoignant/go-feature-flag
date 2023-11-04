package org.gofeatureflag.openfeature

import dev.openfeature.sdk.ImmutableContext
import dev.openfeature.sdk.OpenFeatureAPI
import dev.openfeature.sdk.Value
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.junit.Assert.assertEquals
import org.junit.Test


/**
 * Example local unit test, which will execute on the development machine (host).
 *
 * See [testing documentation](http://d.android.com/tools/testing).
 */
class ExampleUnitTest {
    @Test
    fun addition_isCorrect() {
        assertEquals(4, 2 + 2)
    }

    @Test
    fun xxx() {
        val evaluationContext = ImmutableContext(
            targetingKey = "0a23d9a5-0a8f-42c9-9f5f-4de3afd6cf99",
            attributes = mutableMapOf(
                "region" to Value.String("us-east-1"),
//                "email" to Value.String("john.doe@gofeatureflag.org")
            )
        )
        OpenFeatureAPI.setProvider(
            GoFeatureFlagProvider(
                options = GoFeatureFlagOptions(
                    endpoint = "http://localhost:1031"
                )
            ), evaluationContext
        )

        OpenFeatureAPI.setEvaluationContext(evaluationContext)
        val cli = OpenFeatureAPI.getClient("cli")
        println(cli.getObjectDetails("object_key", Value.Structure(mapOf())))
        if (cli.getBooleanValue("my-flag", false)) {
            println("my-flag is true")
        }

        OpenFeatureAPI.shutdown()

        Thread.sleep(90000)
        println(cli.getObjectDetails("object_key", Value.Structure(mapOf())))

    }
}