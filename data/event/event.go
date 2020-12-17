package event

import (
	"fmt"
	"time"

	err "github.com/newrelic/infra-integrations-sdk/v4/data/errors"
	agentEventPkg "github.com/newrelic/infrastructure-agent/pkg/event"
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

// Events stores events
type Events []*Event

// Event defines the structure of an event
type Event struct {
	Timestamp int64  `json:"timestamp"`
	Summary   string `json:"summary"`
	Category  string `json:"category,omitempty"`
	// Attributes are optional, they represent additional information that
	// can be attached to an event.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// New creates a new event.
func New(timestamp time.Time, summary, category string) (*Event, error) {
	if len(summary) == 0 {
		return nil, err.ParameterCannotBeEmpty("summary")
	}
	return &Event{
		Timestamp:  timestamp.Unix(),
		Summary:    summary,
		Category:   category,
		Attributes: make(map[string]interface{}),
	}, nil
}

// NewNotification creates a new notification event.
func NewNotification(summary string) (*Event, error) {
	return New(time.Now(), summary, NotificationEventCategory)
}

// AddAttribute adds an attribute to the Event
func (e *Event) AddAttribute(key string, value interface{}) error {
	if len(key) == 0 {
		return err.ParameterCannotBeEmpty("key")
	}

	if agentEventPkg.IsReserved(key) {
		return fmt.Errorf("attribute '%s' is reserved", key)
	}
	e.Attributes[key] = value
	return nil
}
