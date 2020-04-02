package event

import (
	"time"
)

const (
	// NotificationEventCategory category for notification events.
	NotificationEventCategory = "notifications"
)

// Event is the data type to represent arbitrary, one-off messages for key
// activities on a system. Ex:
//
// Event{
//   Timestamp: 12312323,
//   Category: "gear",
//   Summary:  "gear has been changed",
//   Attributes: map[string]interface{}{
//     "oldGear":      3,
//     "newGear":      4,
//     "transmission": "manual",
//   },
// }
type Event struct {
	Timestamp int64  `json:"timestamp"`
	Summary   string `json:"summary"`
	Category  string `json:"category,omitempty"`
	// Attributes are optional, they represent additional information that
	// can be attached to an event.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// New creates a new event.
func New(timestamp time.Time, summary, category string) *Event {
	return &Event{
		Timestamp:  timestamp.Unix(),
		Summary:    summary,
		Category:   category,
		Attributes: make(map[string]interface{}),
	}
}

// NewNotification creates a new notification event.
func NewNotification(summary string) *Event {
	return New(time.Now(), summary, NotificationEventCategory)
}

// AddAttribute adds an attribute to the Event
func (e *Event) AddAttribute(key string, value interface{}) {
	// TODO validate value type (bool, number, string)
	e.setAttribute(key, value)
}

func (e *Event) setAttribute(key string, val interface{}) {
	// TODO validate value type (bool, number, string)
	e.Attributes[key] = val
}
