package org.gofeatureflag.openfeature.utils

import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.Value
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class EvaluationContextUtilTest {

    @Test
    fun `isAnonymousUser should return true when context is null`() {
        val result = EvaluationContextUtil.isAnonymousUser(null)
        assertTrue(result)
    }

    @Test
    fun `isAnonymousUser should return true when anonymous field is true`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.Boolean(true))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertTrue(result)
    }

    @Test
    fun `isAnonymousUser should return false when anonymous field is false`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.Boolean(false))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun `isAnonymousUser should return false when anonymous field is not present`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("email" to Value.String("test@example.com"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun `isAnonymousUser should return false when anonymous field is null`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("null"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun `isAnonymousUser should return false when context has no attributes`() {
        val context = ImmutableContext(targetingKey = "user123")

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun `isAnonymousUser should return false when anonymous field is string true`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("true"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun `isAnonymousUser should return false when anonymous field is string false`() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("false"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }
}
