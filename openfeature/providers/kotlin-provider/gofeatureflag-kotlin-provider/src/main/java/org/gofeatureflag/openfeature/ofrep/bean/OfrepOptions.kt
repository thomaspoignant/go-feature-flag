package org.gofeatureflag.openfeature.ofrep.bean

import okhttp3.Headers


data class OfrepOptions(
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
     * (optional) headers to add to the OFREP calls
     * Default: empty
     */
    val headers: Headers? = null,

    /**
     * (optional) polling interval in millisecond to refresh the flags
     * Default: 300000 (5 minutes)
     */
    val pollingIntervalInMillis: Long = 300000
)
