package sse

import (
	"strconv"
	"time"
)

// EventLog holds all of previous events
type EventLog []*Event

// Add event to eventlog
func (e *EventLog) Add(ev *Event) {
	if !ev.hasContent() {
		return
	}

	ev.ID = []byte(e.currentindex())
	ev.timestamp = time.Now()
	*e = append(*e, ev)
}

// Cutting event to eventlog
func (e *EventLog) Cutting(ev *Event, size int) {
	e.Add(ev)
	if len(*e) <= size {
		return
	}

	ne := make(EventLog, 0, size)
	ne = append(ne, (*e)[len(*e)-size:]...)
	*e = ne
}

// Clear events from eventlog
func (e *EventLog) Clear() {
	*e = nil
}

// Replay events to a subscriber
func (e *EventLog) Replay(s *Subscriber) {
	for i := 0; i < len(*e); i++ {
		id, _ := strconv.Atoi(string((*e)[i].ID))
		if id >= s.eventid {
			s.connection <- (*e)[i]
		}
	}
}

func (e *EventLog) currentindex() string {
	return strconv.Itoa(len(*e))
}
