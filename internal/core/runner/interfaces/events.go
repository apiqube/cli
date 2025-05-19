package interfaces

import "time"

type EventEmitter interface {
	Emit(event Event)
	Subscribe(handler EventHandler)
	Unsubscribe(handler EventHandler)
}

type Event struct {
	Type      string // e.g. "case.started", "case.finished", "plan.error"
	Timestamp time.Time
	Payload   map[string]any
}

type EventHandler func(event Event)
