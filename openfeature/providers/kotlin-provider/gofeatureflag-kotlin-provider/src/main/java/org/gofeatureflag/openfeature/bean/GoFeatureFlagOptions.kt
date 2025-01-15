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

    /**
     * (optional) polling interval in millisecond to refresh the flags
     * Default: 300000 (5 minutes)
     */
    val pollingIntervalInMillis: Long = 300000,

    /**
     * (optional) interval time we publish statistics collection data to the proxy.
     * The parameter is used only if the cache is enabled, otherwise the collection of the data is done directly
     * when calling the evaluation API.
     * default: 1000 ms
     */
    val flushIntervalMs: Long = 300000,

    /**
     *  (optional) exporter metadata is a set of key-value that will be added to the metadata when calling the
     *  exporter API. All those informations will be added to the event produce by the exporter.
     *
     * ‼️Important: If you are using a GO Feature Flag relay proxy before version v1.41.0, the information
     * of this field will not be added to your feature events.
     */
    val exporterMetadata: Map<String, Any> = emptyMap(),
)

