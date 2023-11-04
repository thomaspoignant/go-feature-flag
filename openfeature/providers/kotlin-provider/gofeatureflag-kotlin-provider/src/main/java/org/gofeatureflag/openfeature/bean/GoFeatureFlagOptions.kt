package org.gofeatureflag.openfeature.bean

data class GoFeatureFlagOptions(
    /**
     * (mandatory) endpoint contains the DNS of your GO Feature Flag relay proxy
     * example: https://mydomain.com/gofeatureflagproxy/
     */
    val endpoint: String,

    /**
     * (optional) timeout in millisecond we are waiting when calling the
     * go-feature-flag relay proxy API.
     * Default: 10000 ms
     */
    val timeout: Long = 10000,

    /**
     * (optional) maxIdleConnections is the maximum number of connexions in the connexion pool.
     * Default: 1000
     */
    val maxIdleConnections: Int = 1000,

    /**
     * (optional) keepAliveDuration is the time in millisecond we keep the connexion open.
     * Default: 7200000 (2 hours)
     */
    val keepAliveDuration: Long = 7200000,

    /**
     * (optional) apiKey, if the relay proxy is configured to authenticate the requests, you should provide
     * an API Key to the provider.
     * Please ask the administrator of the relay proxy to provide an API Key.
     * (This feature is available only if you are using GO Feature Flag relay proxy v1.7.0 or above)
     * Default: null
     */
    val apiKey: String? = null,

    /** (optional) retryDelay is the time in millisecond we wait before retrying to connect to the relay proxy.
     * Default: 1000 ms
     */
    val retryDelay: Long = 1000,
)
