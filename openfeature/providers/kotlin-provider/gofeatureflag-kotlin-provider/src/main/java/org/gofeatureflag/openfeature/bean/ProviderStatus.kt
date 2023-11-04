package org.gofeatureflag.openfeature.bean

enum class ProviderStatus {
    /**
     * The provider has not been initialized and cannot yet evaluate flags.
     */
    NOT_READY,

    /**
     * The provider is ready to resolve flags.
     */
    READY,

    /**
     * The provider is in an error state and unable to evaluate flags.
     */
    ERROR,

    /**
     * The provider's cached state is no longer valid and may not be up-to-date with the source of truth.
     */
    STALE
}