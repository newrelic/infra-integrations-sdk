package integration

import (
	"testing"

	"strconv"
	"sync"

	"encoding/json"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
)

func TestNewEntity(t *testing.T) {
	e, err := newEntity("name", "type", persist.NewInMemoryStore())

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "type", e.Metadata.Namespace)
}

func TestEntitiesRequireNameAndType(t *testing.T) {
	_, err := newEntity("", "", nil)

	assert.Error(t, err)
}

func TestAddNotificationEvent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.NewNotification("TestSummary"))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 1)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "notifications" {
		t.Error("malformed event")
	}
}

func TestAddNotificationWithEmptySummaryFails(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.NewNotification(""))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestAddEvent_Entity(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.New("TestSummary", "TestCategory"))
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
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func TestAddEvent_Entity_EmptySummary_Error(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New("", "TestCategory"))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestEntity_AddInventoryConcurrent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	itemsAmount := 100
	wg := sync.WaitGroup{}
	wg.Add(itemsAmount)
	for i := 0; i < itemsAmount; i++ {
		go func(j int) {
			assert.NoError(t, en.SetInventoryItem(strconv.Itoa(j), "foo", "bar"))
			wg.Done()
		}(i)
	}

	wg.Wait()
	assert.Len(t, en.Inventory.Items(), itemsAmount)
}

func TestEntity_DefaultEntityIsNotSerialized(t *testing.T) {
	e := newLocalEntity(persist.NewInMemoryStore())
	j, err := json.Marshal(e)

	assert.NoError(t, err)
	assert.Equal(t, `{"metrics":[],"inventory":{},"events":[]}`, string(j))
}

func TestEntity_IsDefaultEntity(t *testing.T) {
	e := newLocalEntity(persist.NewInMemoryStore())

	assert.Empty(t, e.Metadata, "default entity should have no identifier")
	assert.True(t, e.isLocalEntity())
}
