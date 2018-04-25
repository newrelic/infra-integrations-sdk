package integration

import (
	"testing"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestNewEntity(t *testing.T) {
	e, err := NewEntity("name", "type", persist.NewInMemoryStore())

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "type", e.Metadata.Type)
}

func TestAddNotificationEvent(t *testing.T) {
	en, err := NewEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(metric.NewNotification("TestSummary"))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 1)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "notifications" {
		t.Error("malformed event")
	}
}

func TestAddNotificationWithEmptySummaryFails(t *testing.T) {
	en, err := NewEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(metric.NewNotification(""))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestAddEvent_Entity(t *testing.T) {
	en, err := NewEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(metric.NewEvent("TestSummary", "TestCategory"))
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "TestCategory" {
		t.Error("event malformed")
	}

	if len(en.Events) != 1 {
		t.Error("not expected length of events")
	}
}

func TestAddEvent(t *testing.T) {
	en, err := NewEntity("Entity1", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(metric.NewEvent("TestSummary", ""))
	assert.NoError(t, err)

	err = en.AddEvent(metric.NewEvent("TestSummary", ""))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func TestAddEvent_Entity_EmptySummary_Error(t *testing.T) {
	en, err := NewEntity("Entity1", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(metric.NewEvent("", "TestCategory"))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)

}
