package org.gofeatureflag.openfeature.utils

import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.Value
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class EvaluationContextUtilTest {

    @Test
    fun isAnonymousUserReturnsTrueWhenContextIsNull() {
        val result = EvaluationContextUtil.isAnonymousUser(null)
        assertTrue(result)
    }

    @Test
    fun isAnonymousUserReturnsTrueWhenAnonymousFieldIsTrue() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.Boolean(true))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertTrue(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenAnonymousFieldIsFalse() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.Boolean(false))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenAnonymousFieldIsNotPresent() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("email" to Value.String("test@example.com"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenAnonymousFieldIsNull() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("null"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenContextHasNoAttributes() {
        val context = ImmutableContext(targetingKey = "user123")

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenAnonymousFieldIsStringTrue() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("true"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }

    @Test
    fun isAnonymousUserReturnsFalseWhenAnonymousFieldIsStringFalse() {
        val context = ImmutableContext(
            targetingKey = "user123",
            attributes = mapOf("anonymous" to Value.String("false"))
        )

        val result = EvaluationContextUtil.isAnonymousUser(context)
        assertFalse(result)
    }
}
