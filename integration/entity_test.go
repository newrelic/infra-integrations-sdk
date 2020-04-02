package integration

import (
	"testing"
	"time"

	"strconv"
	"sync"

	"encoding/json"

	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Entity_NewEntityInitializesCorrectly(t *testing.T) {

	e, err := newEntity("name", "type", "displayName", persist.NewInMemoryStore())

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "displayName", e.Metadata.DisplayName)
	assert.Equal(t, "type", e.Metadata.EntityType)
	assert.Empty(t, e.Metadata.GetTags())
	assert.Empty(t, e.Events)
	assert.Empty(t, e.Metrics)
	assert.NotNil(t, e.Inventory)
	assert.Empty(t, e.Inventory.Items())

}

func Test_Entity_EntityAddTag(t *testing.T) {
	e, err := newEntity("name", "type", "", persist.NewInMemoryStore())
	assert.NoError(t, err)

	e.AddTag("key1", "val1")
	assert.Len(t, e.Tags(), 1, "tags should have been added to the entity")

}

func Test_Entity_NewEntityWithTags(t *testing.T) {
	e, err := newEntity("name", "type", "displayName", persist.NewInMemoryStore())
	assert.NoError(t, err)

	e.AddTag("env", "prod")
	e.AddTag("srv", "auth")

	assert.Len(t, e.Metadata.Tags, 2)
	assert.Equal(t, e.Metadata.GetTag("env"), "prod")
	assert.Equal(t, e.Metadata.GetTag("srv"), "auth")
}

func Test_Entity_AddTagReplacesExisting(t *testing.T) {
	e, err := newEntity("name", "type", "displayName", persist.NewInMemoryStore())
	assert.NoError(t, err)

	e.AddTag("env", "prod")
	assert.Len(t, e.Metadata.Tags, 1)
	assert.Equal(t, e.Metadata.GetTag("env"), "prod")

	e.AddTag("env", "staging")

	assert.Len(t, e.Metadata.Tags, 1)
	assert.Equal(t, e.Metadata.GetTag("env"), "staging")
}

func Test_Entity_NameAndTypeCannotBeEmpty(t *testing.T) {
	_, err := newEntity("", "", "", nil)

	assert.Error(t, err)
}

func Test_Entity_AddNotificationEvent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", "", persist.NewInMemoryStore())
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

func Test_Entity_AddEventWithAttributes(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	require.NoError(t, err)

	ev := event.New(time.Now(), "TestSummary", "TestCategory")
	ev.AddAttribute("attrKey", "attrVal")
	err = en.AddEvent(ev)
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

func Test_Entity_AddNotificationWithEmptySummaryFails(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.NewNotification(""))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func Test_Entity_AddEventThrowsNoError(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.New(time.Now(), "TestSummary", "TestCategory"))
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

func Test_Entity_AddEventReturnsNoError(t *testing.T) {
	en, err := newEntity("Entity1", "displayname", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New(time.Now(), "TestSummary", ""))
	assert.NoError(t, err)

	err = en.AddEvent(event.New(time.Now(), "TestSummary", ""))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func Test_Entity_AddEventWithEmptySummaryReturnsError(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	err = en.AddEvent(event.New(time.Now(), "", "TestCategory"))
	assert.Error(t, err)
	assert.Len(t, en.Events, 0)
}

func Test_Entity_AddInventoryConcurrent(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1", persist.NewInMemoryStore())
	assert.NoError(t, err)

	itemsAmount := 100
	wg := sync.WaitGroup{}
	wg.Add(itemsAmount)
	for i := 0; i < itemsAmount; i++ {
		go func(j int) {
			assert.NoError(t, en.AddInventoryItem(strconv.Itoa(j), "foo", "bar"))
			wg.Done()
		}(i)
	}

	wg.Wait()
	assert.Len(t, en.Inventory.Items(), itemsAmount)
}

func Test_Entity_IsAnonymousEntity(t *testing.T) {
	e := newAnonymousEntity(persist.NewInMemoryStore())

	assert.Empty(t, e.Metadata, "default entity should have no identifier")
	assert.True(t, e.isAnonymousEntity())
}

func Test_Entity_AnonymousEntityIsProperlySerialized(t *testing.T) {
	e := newAnonymousEntity(persist.NewInMemoryStore())
	j, err := json.Marshal(e)

	assert.NoError(t, err)
	assert.Equal(t, `{"common":{},"metrics":[],"inventory":{},"events":[]}`, string(j))
}

func Test_Entity_EntitiesWithSameMetadataAreSameAs(t *testing.T) {
	e1, err := newEntity("entity", "type", "", persist.NewInMemoryStore())
	assert.NoError(t, err)
	e1.AddTag("env", "prod")

	e2, err := newEntity("entity", "type", "", persist.NewInMemoryStore())
	assert.NoError(t, err)
	e2.AddTag("env", "prod")

	e3, err := newEntity("entity", "otherType", "ns", persist.NewInMemoryStore())
	assert.NoError(t, err)

	assert.True(t, e1.SameAs(e2))
	assert.False(t, e1.SameAs(e3))
}
