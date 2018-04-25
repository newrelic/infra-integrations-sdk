package metric

const (
	// NotificationEventCategory category for notification events.
	NotificationEventCategory = "notifications"
)

// Event is the data type to represent arbitrary, one-off messages for key
// activities on a system.
type Event struct {
	Summary  string `json:"summary"`
	Category string `json:"category,omitempty"`
}

// NewEvent creates a new event.
func NewEvent(summary, category string) *Event {
	return &Event{
		Summary:  summary,
		Category: category,
	}
}

// NewNotification creates a new notification event.
func NewNotification(summary string) *Event {
	return NewEvent(summary, NotificationEventCategory)
}
