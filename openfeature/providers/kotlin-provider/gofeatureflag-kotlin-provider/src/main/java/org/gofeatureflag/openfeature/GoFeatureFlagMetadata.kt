package org.gofeatureflag.openfeature

import dev.openfeature.kotlin.sdk.ProviderMetadata

class GoFeatureFlagMetadata : ProviderMetadata {
    override val name: String
        get() = "GO Feature Flag Provider"
}