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
    private val eventList = Collections.synchronizedList(mutableListOf<Event>())
    private var timer: Timer? = null

    fun addEvent(event: Event) {
        eventList.add(event)
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
            val events = eventList.toList()
            this.goffApi.postEventsToDataCollector(events)
            eventList.clear()
        } catch (e: Exception) {
            throw e
        }
    }
}