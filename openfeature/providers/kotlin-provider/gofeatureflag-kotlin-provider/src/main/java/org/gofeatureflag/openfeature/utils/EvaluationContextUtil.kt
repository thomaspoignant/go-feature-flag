package org.gofeatureflag.openfeature.utils

import dev.openfeature.kotlin.sdk.EvaluationContext

/**
 * EvaluationContextUtil is a utility class that provides methods to work with the evaluation context.
 * It is used to check if the user is anonymous or not.
 */
object EvaluationContextUtil {

    /**
     * anonymousFieldName is the name of the field in the evaluation context that indicates
     * if the user is anonymous.
     */
    private const val ANONYMOUS_FIELD_NAME = "anonymous"

    /**
     * isAnonymousUser is checking if the user in the evaluationContext is anonymous.
     *
     * @param ctx - EvaluationContext from open-feature
     * @return true if the user is anonymous, false otherwise
     */
    fun isAnonymousUser(ctx: EvaluationContext?): Boolean {
        if (ctx == null) {
            return true
        }
        val value = ctx.getValue(ANONYMOUS_FIELD_NAME)
        return value?.asBoolean() == true
    }
}
