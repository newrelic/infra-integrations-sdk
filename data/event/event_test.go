package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Event_NewEvent(t *testing.T) {
	now := time.Now()
	e := New(now,"summary", "category")

	assert.Equal(t, now.Unix(), e.Timestamp)
	assert.Equal(t, "summary", e.Summary)
	assert.Equal(t, "category", e.Category)
}

func Test_Event_NewNotification(t *testing.T) {
	n := NewNotification("summary")
	assert.Equal(t, "summary",  n.Summary)
}

func Test_Event_NewEventsWithAttributes(t *testing.T) {
	now := time.Now()
	e := New(now, "summary", "category")
	e.AddAttribute("attrKey", "attrVal")

	assert.Equal(t, now.Unix(), e.Timestamp)
	assert.Equal(t,"summary", e.Summary)
	assert.Equal(t, "category", e.Category)
	assert.Equal(t, "attrVal", e.Attributes["attrKey"])
}
