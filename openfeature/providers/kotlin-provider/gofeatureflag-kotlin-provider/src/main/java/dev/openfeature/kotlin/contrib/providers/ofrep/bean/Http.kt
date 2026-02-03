package dev.openfeature.kotlin.contrib.providers.ofrep.bean

import io.ktor.client.HttpClient
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.engine.okhttp.OkHttp
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json
import okhttp3.ConnectionPool
import java.util.concurrent.TimeUnit

internal fun defaultHttpEngine(options: OfrepOptions): HttpClientEngine {
    return OkHttp.create {
        config {
            retryOnConnectionFailure(true)
            connectionPool(
                ConnectionPool(
                    options.maxIdleConnections,
                    options.keepAliveDuration.inWholeMilliseconds,
                    TimeUnit.MILLISECONDS
                )
            )
        }
    }
}

internal fun createHttpClient(options: OfrepOptions): HttpClient {
    val httpEngine = options.httpClientEngine ?: defaultHttpEngine(options)
    return HttpClient(httpEngine) {
        install(ContentNegotiation) {
            json()
        }
    }
}
