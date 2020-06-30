package eventstoredb

import "github.com/google/uuid"

// Event represents an event document
type Event struct {
	EventID   string      `json:"eventId"`
	EventType string      `json:"eventType"`
	Data      interface{} `json:"data"`
	Metadata  interface{} `json:"metadata"`
}

// NewEvent constructs a new Event type.
func NewEvent(eventType string, data, metadata interface{}) *Event {
	return &Event{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Data:      data,
		Metadata:  metadata,
	}
}
