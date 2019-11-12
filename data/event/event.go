package event

import "github.com/newrelic/infra-integrations-sdk/data/attribute"

const (
	// NotificationEventCategory category for notification events.
	NotificationEventCategory = "notifications"
)

// Event is the data type to represent arbitrary, one-off messages for key
// activities on a system. Ex:
//
// Event{
//   Category: "gear",
//   Summary:  "gear has been changed",
//   Attributes: map[string]interface{}{
//     "oldGear":      3,
//     "newGear":      4,
//     "transmission": "manual",
//   },
// }
type Event struct {
	Summary  string `json:"summary"`
	Category string `json:"category,omitempty"`
	// Attributes are optional, they represent additional information that
	// can be attached to an event.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// New creates a new event.
func New(summary, category string) *Event {
	return &Event{
		Summary:    summary,
		Category:   category,
		Attributes: make(map[string]interface{}),
	}
}

// NewNotification creates a new notification event.
func NewNotification(summary string) *Event {
	return New(summary, NotificationEventCategory)
}

// NewWithAttributes creates a new event with the given attributes
func NewWithAttributes(summary, category string, attributes map[string]interface{}) *Event {
	e := New(summary, category)
	e.Attributes = attributes
	return e
}

func (e *Event) setAttribute(key string, val interface{}) {
	e.Attributes[key] = val
}

// AddCustomAttributes add customAttributes to the Event
func AddCustomAttributes(e *Event, customAttributes []attribute.Attribute) {
	for _, attr := range customAttributes {
		e.setAttribute(attr.Key, attr.Value)
	}
}
