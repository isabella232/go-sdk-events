package ldevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldlog"
)

type eventsOutbox struct {
	events           []Event
	summarizer       eventSummarizer
	capacity         int
	capacityExceeded bool
	droppedEvents    int
	loggers          ldlog.Loggers
}

func newEventsOutbox(capacity int, loggers ldlog.Loggers) *eventsOutbox {
	return &eventsOutbox{
		events:     make([]Event, 0, capacity),
		summarizer: newEventSummarizer(),
		capacity:   capacity,
		loggers:    loggers,
	}
}

func (b *eventsOutbox) addEvent(event Event) {
	if len(b.events) >= b.capacity {
		if !b.capacityExceeded {
			b.capacityExceeded = true
			b.loggers.Warn("Exceeded event queue capacity. Increase capacity to avoid dropping events.")
		}
		b.droppedEvents++
		return
	}
	b.capacityExceeded = false
	b.events = append(b.events, event)
}

func (b *eventsOutbox) addToSummary(event FeatureRequestEvent) {
	b.summarizer.summarizeEvent(event)
}

func (b *eventsOutbox) getPayload() flushPayload {
	var copied []Event
	if len(b.events) > 0 {
		copied = make([]Event, len(b.events))
		copy(copied, b.events)
	}
	return flushPayload{
		events:  copied,
		summary: b.summarizer.snapshot(),
	}
}

func (b *eventsOutbox) clear() {
	for i := range b.events {
		b.events[i] = nil
	}
	b.events = b.events[0:0]
	b.summarizer.reset()
}
