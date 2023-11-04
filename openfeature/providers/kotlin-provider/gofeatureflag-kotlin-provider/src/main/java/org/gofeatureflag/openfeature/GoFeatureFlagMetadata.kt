package org.gofeatureflag.openfeature

import dev.openfeature.sdk.ProviderMetadata
import java.security.Provider

class GoFeatureFlagMetadata() : ProviderMetadata {
    override val name: String
        get() = "GoFeatureFlagProvider"
}