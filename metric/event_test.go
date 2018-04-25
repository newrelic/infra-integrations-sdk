package metric

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestNewEvent(t *testing.T) {
	e := NewEvent("summary", "category")

	assert.Equal(t, e.Summary, "summary")
	assert.Equal(t, e.Category, "category")
}

func TestNewNotification(t *testing.T) {
	n := NewNotification("summary")
	assert.Equal(t, n.Summary, "summary")
}
