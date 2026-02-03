package org.gofeatureflag.openfeature

import dev.openfeature.kotlin.sdk.ImmutableContext
import dev.openfeature.kotlin.sdk.ImmutableStructure
import dev.openfeature.kotlin.sdk.OpenFeatureAPI
import dev.openfeature.kotlin.sdk.TrackingEventDetails
import dev.openfeature.kotlin.sdk.Value
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import okhttp3.mockwebserver.MockResponse
import okhttp3.mockwebserver.MockWebServer
import okhttp3.mockwebserver.RecordedRequest
import org.gofeatureflag.openfeature.bean.GoFeatureFlagOptions
import org.gofeatureflag.openfeature.bean.Events
import org.gofeatureflag.openfeature.bean.createEventsGson
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test
import java.nio.file.Files
import java.nio.file.Paths
import kotlin.time.Duration.Companion.milliseconds

class GoFeatureFlagProviderTest {
    private var mockWebServer: MockWebServer? = null

    @Before
    fun before() {
        mockWebServer = MockWebServer()
        mockWebServer!!.start(10031)
    }

    @After
    fun after() {
        mockWebServer!!.shutdown()
    }

    @Test
    fun `should call the hook with a valid result`() {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(Files.readAllBytes(Paths.get(jsonFilePath)))
        mockWebServer!!.enqueue(MockResponse().setBody(jsonString).setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        val options =
            GoFeatureFlagOptions(
                endpoint = mockWebServer!!.url("/").toString(),
                flushIntervalMs = 100,
                pollingInterval = 10000.milliseconds
            )

        val provider = GoFeatureFlagProvider(options)
        val ctx = ImmutableContext(targetingKey = "123")
        runBlocking {
            OpenFeatureAPI.setProviderAndWait(
                provider = provider,
                dispatcher = Dispatchers.IO,
                initialContext = ctx
            )
        }

        val client = OpenFeatureAPI.getClient()
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(1000)
        mockWebServer!!.takeRequest()
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()


        val got = createEventsGson().fromJson(recordedRequest.body.readUtf8(), Events::class.java)
        assertEquals(6, got.featureEvents?.size)
        got.featureEvents?.forEach {
            assertEquals("title-flag", it.key)
            assertEquals("123", it.userKey)
            assertEquals("GO Feature Flag", it.value)
            assertEquals(false, it.default)
            assertEquals("PROVIDER_CACHE", it.source)
            assertEquals("default_title", it.variation)
            assertEquals("feature", it.kind)
            assertEquals("user", it.contextKind)
            assertNotNull(it.creationDate)
        }
    }

    @Test
    fun `should call the hook with an error result`() {
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        val options =
            GoFeatureFlagOptions(
                endpoint = mockWebServer!!.url("/").toString(),
                flushIntervalMs = 100,
                pollingInterval = 10000.milliseconds
            )

        val provider = GoFeatureFlagProvider(options)
        val ctx = ImmutableContext(targetingKey = "123")
        runBlocking {
            OpenFeatureAPI.setProviderAndWait(
                provider = provider,
                dispatcher = Dispatchers.IO,
                initialContext = ctx
            )
        }

        val client = OpenFeatureAPI.getClient()
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(500)
        mockWebServer!!.takeRequest()
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()


        val got = createEventsGson().fromJson(recordedRequest.body.readUtf8(), Events::class.java)
        assertEquals(6, got.featureEvents?.size)
        got.featureEvents?.forEach {
            assertEquals("title-flag", it.key)
            assertEquals("123", it.userKey)
            assertEquals("default", it.value)
            assertEquals(true, it.default)
            assertEquals("PROVIDER_CACHE", it.source)
            assertEquals("SdkDefault", it.variation)
            assertEquals("feature", it.kind)
            assertEquals("user", it.contextKind)
            assertNotNull(it.creationDate)
        }
    }

    @Test
    fun `should call the hook multiple times`() {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(Files.readAllBytes(Paths.get(jsonFilePath)))
        mockWebServer!!.enqueue(MockResponse().setBody(jsonString).setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        val options =
            GoFeatureFlagOptions(
                endpoint = mockWebServer!!.url("/").toString(),
                flushIntervalMs = 100,
                pollingInterval = 10000.milliseconds
            )

        val provider = GoFeatureFlagProvider(options)
        val ctx = ImmutableContext(targetingKey = "123")
        runBlocking {
            OpenFeatureAPI.setProviderAndWait(
                provider = provider,
                dispatcher = Dispatchers.IO,
                initialContext = ctx
            )
        }

        val client = OpenFeatureAPI.getClient()
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(1000)
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(1000)
        mockWebServer!!.takeRequest()
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()
        val got = createEventsGson().fromJson(recordedRequest.body.readUtf8(), Events::class.java)
        assertEquals(6, got.featureEvents?.size)
        val recordedRequest2: RecordedRequest = mockWebServer!!.takeRequest()
        val got2 = createEventsGson().fromJson(recordedRequest2.body.readUtf8(), Events::class.java)
        assertEquals(3, got2.events?.size)
    }

    @Test
    fun `should call the hook and send metadata`() {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(Files.readAllBytes(Paths.get(jsonFilePath)))
        mockWebServer!!.enqueue(MockResponse().setBody(jsonString).setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        val options =
            GoFeatureFlagOptions(
                endpoint = mockWebServer!!.url("/").toString(),
                flushIntervalMs = 100,
                pollingInterval = 10000.milliseconds,
                exporterMetadata = mapOf("device" to "Pixel 4", "appVersion" to "1.0.0")
            )

        val provider = GoFeatureFlagProvider(options)
        val ctx = ImmutableContext(targetingKey = "123")
        runBlocking {
            OpenFeatureAPI.setProviderAndWait(
                provider = provider,
                dispatcher = Dispatchers.IO,
                initialContext = ctx
            )
        }

        val client = OpenFeatureAPI.getClient()
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(1000)
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        Thread.sleep(1000)
        mockWebServer!!.takeRequest()
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()
        val got = createEventsGson().fromJson(recordedRequest.body.readUtf8(), Events::class.java)
        assertEquals(6, got.featureEvents?.size)
        assertEquals("Pixel 4", got.meta["device"])
        assertEquals("1.0.0", got.meta["appVersion"])
        assertEquals("android", got.meta["provider"])
        assertEquals(true, got.meta["openfeature"])
        val recordedRequest2: RecordedRequest = mockWebServer!!.takeRequest()
        val got2 = createEventsGson().fromJson(recordedRequest2.body.readUtf8(), Events::class.java)
        assertEquals(3, got2.events?.size)
        assertEquals("Pixel 4", got2.meta["device"])
        assertEquals("1.0.0", got2.meta["appVersion"])
        assertEquals("android", got2.meta["provider"])
        assertEquals(true, got2.meta["openfeature"])
    }

    @Test
    fun `should collect feature events and tracking events`() {
        val jsonFilePath =
            javaClass.classLoader?.getResource("org.gofeatureflag.openfeature.ofrep/valid_api_response.json")?.file
        val jsonString = String(Files.readAllBytes(Paths.get(jsonFilePath)))
        mockWebServer!!.enqueue(MockResponse().setBody(jsonString).setHeader("Content-Type", "application/json").setResponseCode(200))
        mockWebServer!!.enqueue(MockResponse().setBody("{}").setHeader("Content-Type", "application/json").setResponseCode(200))
        val options =
            GoFeatureFlagOptions(
                endpoint = mockWebServer!!.url("/").toString(),
                flushIntervalMs = 100,
                pollingInterval = 10000.milliseconds,
                exporterMetadata = mapOf("device" to "Pixel 4", "appVersion" to "1.0.0")
            )

        val provider = GoFeatureFlagProvider(options)
        val ctx = ImmutableContext(targetingKey = "123")
        runBlocking {
            OpenFeatureAPI.setProviderAndWait(
                provider = provider,
                dispatcher = Dispatchers.IO,
                initialContext = ctx
            )
        }

        val client = OpenFeatureAPI.getClient()
        // Feature evaluations that will generate FeatureEvents
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")
        client.getStringValue("title-flag", "default")

        // Tracking calls that will generate TrackingEvents
        val trackingDetails = TrackingEventDetails(
            100.0,
            ImmutableStructure(
                "action" to Value.String("click"),
                "element" to Value.String("button")
            )
        )
        provider.track("user-action", ctx, trackingDetails)
        provider.track("user-action", ctx, trackingDetails)
        provider.track("page-view", ctx, null)

        Thread.sleep(1000)
        mockWebServer!!.takeRequest()
        val recordedRequest: RecordedRequest = mockWebServer!!.takeRequest()

        val got = createEventsGson().fromJson(recordedRequest.body.readUtf8(), Events::class.java)

        // Verify we have both feature events and tracking events
        assertEquals(3, got.featureEvents?.size)
        assertEquals(3, got.trackingEvents?.size)
        assertEquals(6, got.events?.size)

        // Verify feature events
        got.featureEvents?.forEach {
            assertEquals("title-flag", it.key)
            assertEquals("123", it.userKey)
            assertEquals("GO Feature Flag", it.value)
            assertEquals(false, it.default)
            assertEquals("PROVIDER_CACHE", it.source)
            assertEquals("default_title", it.variation)
            assertEquals("feature", it.kind)
            assertEquals("user", it.contextKind)
            assertNotNull(it.creationDate)
        }

        // Verify tracking events
        val userActionTrackingEvents = got.trackingEvents?.filter { it.key == "user-action" }
        assertEquals(2, userActionTrackingEvents?.size)
        userActionTrackingEvents?.forEach {
            assertEquals("user-action", it.key)
            assertEquals("123", it.userKey)
            assertEquals("tracking", it.kind)
            assertEquals("user", it.contextKind)
            assertNotNull(it.creationDate)
            // TrackingEventDetails contains value (as Number) and the structure map
            assertEquals(100.0, (it.trackingEventDetails?.get("value") as? Double))
            assertEquals("click", it.trackingEventDetails?.get("action"))
            assertEquals("button", it.trackingEventDetails?.get("element"))
        }

        val pageViewTrackingEvents = got.trackingEvents?.filter { it.key == "page-view" }
        assertEquals(1, pageViewTrackingEvents?.size)
        pageViewTrackingEvents?.first()?.let {
            assertEquals("page-view", it.key)
            assertEquals("123", it.userKey)
            assertEquals("tracking", it.kind)
            assertEquals("user", it.contextKind)
            assertNotNull(it.creationDate)
            assertTrue(it.trackingEventDetails == null || it.trackingEventDetails!!.isEmpty())
        }

        // Verify metadata
        assertEquals("Pixel 4", got.meta["device"])
        assertEquals("1.0.0", got.meta["appVersion"])
        assertEquals("android", got.meta["provider"])
        assertEquals(true, got.meta["openfeature"])
    }
}