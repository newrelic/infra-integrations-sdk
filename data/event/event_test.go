package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEvent(t *testing.T) {
	e := New("summary", "category")

	assert.Equal(t, e.Summary, "summary")
	assert.Equal(t, e.Category, "category")
}

func TestNewNotification(t *testing.T) {
	n := NewNotification("summary")
	assert.Equal(t, n.Summary, "summary")
}

func TestNewAttributes(t *testing.T) {
	e := NewWithAttributes(
		"summary",
		"category",
		map[string]interface{}{"attrKey": "attrVal"},
	)

	assert.Equal(t, e.Summary, "summary")
	assert.Equal(t, e.Category, "category")
	assert.Equal(t, e.Attributes["attrKey"], "attrVal")
}
