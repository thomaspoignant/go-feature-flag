package org.gofeatureflag.openfeature.controller

import kotlinx.coroutines.runBlocking
import org.gofeatureflag.openfeature.bean.Event
import java.util.Collections
import java.util.Timer
import java.util.TimerTask

class DataCollectorManager(
    private val goffApi: GoFeatureFlagApi,
    private val flushIntervalMs: Long
) {
    private val featureEventList = Collections.synchronizedList(mutableListOf<Event>())
    private var timer: Timer? = null

    fun addEvent(featureEvent: Event) {
        featureEventList.add(featureEvent)
    }

    fun start() {
        val task: TimerTask = object : TimerTask() {
            override fun run() {
                runBlocking {
                    try {
                        this@DataCollectorManager.sendToCollector()
                    } catch (e: Throwable) {
                        // do nothing
                    }
                }
            }
        }
        val timer = Timer()
        timer.schedule(task, flushIntervalMs, flushIntervalMs)
        this.timer = timer
    }

    fun stop() {
        this.timer?.cancel()
    }

    private suspend fun sendToCollector() {
        try {
            val events = featureEventList.toList()
            this.goffApi.postEventsToDataCollector(events)
            featureEventList.clear()
        } catch (e: Exception) {
            throw e
        }
    }
}