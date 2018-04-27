package event

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

// New creates a new event.
func New(summary, category string) *Event {
	return &Event{
		Summary:  summary,
		Category: category,
	}
}

// NewNotification creates a new notification event.
func NewNotification(summary string) *Event {
	return New(summary, NotificationEventCategory)
}
