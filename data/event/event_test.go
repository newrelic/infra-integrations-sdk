package event

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
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

func TestNewEventsWithAttributes(t *testing.T) {
	e := NewWithAttributes(
		"summary",
		"category",
		map[string]interface{}{"attrKey": "attrVal"},
	)

	assert.Equal(t, e.Summary, "summary")
	assert.Equal(t, e.Category, "category")
	assert.Equal(t, e.Attributes["attrKey"], "attrVal")
}

func TestEventsAddCustomAttributes(t *testing.T) {
	e := &Event{
		Summary:    "summary",
		Category:   "category",
		Attributes: map[string]interface{}{"attrKey": "attrVal"},
	}

	a := attribute.Attributes{attribute.Attr("clusterName", "my-cluster")}

	AddCustomAttributes(e, a)

	assert.Equal(t, e.Summary, "summary")
	assert.Equal(t, e.Category, "category")
	assert.Equal(t, e.Attributes, map[string]interface{}{"attrKey": "attrVal", "clusterName": "my-cluster"})
}
