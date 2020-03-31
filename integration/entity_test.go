package integration

import (
	"testing"

	"strconv"
	"sync"

	"encoding/json"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metadata"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntity(t *testing.T) {

	e, err := newEntity("name", "displayName", "type", persist.NewInMemoryStore())

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "displayName", e.Metadata.DisplayName)
	assert.Equal(t, "type", e.Metadata.EntityType)
	assert.Empty(t, e.Metadata.Tags)
	assert.Empty(t, e.Events)
	assert.Empty(t, e.Metrics)
	assert.NotNil(t, e.Inventory)
	assert.Empty(t, e.Inventory.Items())

}

func Test_NewEntityWithTags(t *testing.T) {
	attr1 := metadata.NewTag("env", "prod")
	attr2 := metadata.NewTag("srv", "auth")
	e, err := newEntity("name", "displayName", "type", persist.NewInMemoryStore(), attr1, attr2)

	assert.NoError(t, err)
	assert.Len(t, e.Metadata.Tags, 2)
	assert.Equal(t, e.Metadata.Tags[0], attr1)
	assert.Equal(t, e.Metadata.Tags[1], attr2)
}

func TestEntity_AddTags(t *testing.T) {
	tag := metadata.NewTag("env", "prod")
	e, err := newEntity("name", "", "type", persist.NewInMemoryStore(), tag)
	assert.NoError(t, err)

	e.AddTag("key1", "val1")
	assert.Len(t, e.Tags(), 2, "attributes should have been added to the entity")

}

func Test_NameAndTypeCannotBeEmpty(t *testing.T) {
	_, err := newEntity("", "", "", nil)

	assert.Error(t, err)
}

func TestAddNotificationEvent(t *testing.T) {
	en, err := newEntity("Entity1", "", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}
	en.AddTag("clusterName", "my-cluster-name")

	err = en.AddEvent(event.NewNotification("TestSummary"))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 1)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "notifications" {
		t.Error("malformed event")
	}
}

func Test_AddEventWithAttributes(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	require.NoError(t, err)

	attrs := map[string]interface{}{"attrKey": "attrVal"}
	err = en.AddEvent(event.NewWithAttributes("TestSummary", "TestCategory", attrs))
	assert.NoError(t, err)

	require.Len(t, en.Events, 1)

	assert.Equal(t, "TestSummary", en.Events[0].Summary)
	assert.Equal(t, "TestCategory", en.Events[0].Category)

	expectedAttrs := map[string]interface{}{
		"attrKey": "attrVal",
		//"clusterName": "my-cluster-name", TODO: should this be added to the event?
	}
	assert.Equal(t, expectedAttrs, en.Events[0].Attributes)
}

func Test_AddNotificationWithEmptySummaryFails(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.NewNotification(""))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestAddEvent_Entity(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
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
	en, err := newEntity("Entity1", "displayname", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func TestAddEvent_Entity_EmptySummary_Error(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New("", "TestCategory"))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestEntity_AddInventoryConcurrent(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
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
	e := newAnonymousEntity(persist.NewInMemoryStore(), false)
	j, err := json.Marshal(e)

	assert.NoError(t, err)
	assert.Equal(t, `{"common":{},"metrics":[],"inventory":{},"events":[]}`, string(j))
}

func TestEntity_IsDefaultEntity(t *testing.T) {
	e := newAnonymousEntity(persist.NewInMemoryStore(), false)

	assert.Empty(t, e.Metadata, "default entity should have no identifier")
	assert.True(t, e.isAnonymousEntity())
}

func TestEntity_SameAs(t *testing.T) {
	attr := metadata.NewTag("env", "prod")
	e1, err := newEntity("entity", "", "ns", persist.NewInMemoryStore(), attr)
	assert.NoError(t, err)

	e2, err := newEntity("entity", "", "ns", persist.NewInMemoryStore(), attr)
	assert.NoError(t, err)

	e3, err := newEntity("entity", "", "ns", persist.NewInMemoryStore())
	assert.NoError(t, err)

	assert.True(t, e1.SameAs(e2))
	assert.False(t, e1.SameAs(e3))
}
