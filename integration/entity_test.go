package integration

import (
	"encoding/json"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/newrelic/infra-integrations-sdk/v4/data/event"
)

func Test_Entity_NewEntityInitializesCorrectly(t *testing.T) {

	e, err := newEntity("name", "type", "displayName")

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "displayName", e.Metadata.DisplayName)
	assert.Equal(t, "type", e.Metadata.EntityType)
	assert.Empty(t, e.Metadata.Metadata)
	assert.Empty(t, e.Events)
	assert.Empty(t, e.Metrics)
	assert.NotNil(t, e.Inventory)
	assert.Empty(t, e.Inventory.Items())

}

func Test_Entity_EntityAddTag(t *testing.T) {
	e, err := newEntity("name", "type", "")
	assert.NoError(t, err)

	_ = e.AddTag("key1", "val1")
	assert.Len(t, e.GetMetadata(), 1, "tags should have been added to the entity")

}

func Test_Entity_EntityCannotAddTagWithEmptyName(t *testing.T) {
	e, err := newEntity("name", "type", "")
	assert.NoError(t, err)

	err = e.AddTag("", "val1")
	assert.Error(t, err)
	assert.Len(t, e.GetMetadata(), 0, "tags should NOT have been added to the entity")
}

func Test_Entity_AddTagReplacesExisting(t *testing.T) {
	e, err := newEntity("name", "type", "displayName")
	assert.NoError(t, err)

	_ = e.AddTag("env", "prod")
	assert.Len(t, e.Metadata.Metadata, 1)
	assert.Equal(t, e.Metadata.GetTag("env"), "prod")

	_ = e.AddTag("env", "staging")

	assert.Len(t, e.Metadata.Metadata, 1)
	assert.Equal(t, e.Metadata.GetTag("env"), "staging")
}

func Test_Entity_NameAndTypeCannotBeEmpty(t *testing.T) {
	_, err := newEntity("", "", "")

	assert.Error(t, err)
}

func Test_Entity_AddNotificationEvent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", "")
	if err != nil {
		t.Fatal(err)
	}
	_ = en.AddTag("clusterName", "my-cluster-name")

	ev, _ := event.NewNotification("TestSummary")
	en.AddEvent(ev)
	assert.NoError(t, err)

	assert.Len(t, en.Events, 1)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "notifications" {
		t.Error("malformed event")
	}
}

func Test_Entity_AddEventWithAttributes(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1")
	require.NoError(t, err)

	ev, _ := event.New(time.Now(), "TestSummary", "TestCategory")
	_ = ev.AddAttribute("attrKey", "attrVal")
	en.AddEvent(ev)
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
	en, err := newEntity("Entity1", "displayName", "Type1")
	if err != nil {
		t.Fatal(err)
	}

	ev, err := event.NewNotification("")
	assert.Error(t, err)
	assert.Nil(t, ev)
	assert.Len(t, en.Events, 0)
}

func Test_Entity_AddEventThrowsNoError(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1")
	if err != nil {
		t.Fatal(err)
	}

	ev, err := event.New(time.Now(), "TestSummary", "TestCategory")
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}
	en.AddEvent(ev)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "TestCategory" {
		t.Error("event malformed")
	}

	if len(en.Events) != 1 {
		t.Error("not expected length of events")
	}
}

func Test_Entity_AddEventReturnsNoError(t *testing.T) {
	en, err := newEntity("Entity1", "displayname", "Type1")
	assert.NoError(t, err)

	ev, _ := event.New(time.Now(), "TestSummary", "")
	en.AddEvent(ev)
	assert.NoError(t, err)

	ev, _ = event.New(time.Now(), "TestSummary", "")
	en.AddEvent(ev)
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func Test_Entity_AddEventWithEmptySummaryReturnsError(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1")
	assert.NoError(t, err)

	ev, err := event.New(time.Now(), "", "TestCategory")
	assert.Error(t, err)
	assert.Nil(t, ev)
	assert.Len(t, en.Events, 0)
}

func Test_Entity_AddInventoryConcurrent(t *testing.T) {
	en, err := newEntity("Entity1", "displayName", "Type1")
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
	e := newHostEntity()

	assert.Empty(t, e.Metadata, "default entity should have no identifier")
	assert.True(t, e.isHostEntity())
}

func Test_Entity_AnonymousEntityIsProperlySerialized(t *testing.T) {
	e := newHostEntity()
	j, err := json.Marshal(e)

	assert.NoError(t, err)
	assert.Equal(t, `{"common":{},"metrics":[],"inventory":{},"events":[]}`, string(j))
}

func Test_Entity_EntitiesWithSameMetadataAreSameAs(t *testing.T) {
	e1, err := newEntity("entity", "type", "")
	assert.NoError(t, err)
	_ = e1.AddTag("env", "prod")

	e2, err := newEntity("entity", "type", "")
	assert.NoError(t, err)
	_ = e2.AddTag("env", "prod")

	e3, err := newEntity("entity", "otherType", "ns")
	assert.NoError(t, err)

	assert.True(t, e1.SameAs(e2))
	assert.False(t, e1.SameAs(e3))
}

func TestEntity_AddCommonDimension(t *testing.T) {
	tests := []struct {
		name      string
		commons   metric.Dimensions
		expeceted *Entity
	}{
		{"empty", nil, newHostEntity()},
		{"one entry", metric.Dimensions{"k": "v"}, &Entity{
			CommonDimensions: metric.Dimensions{"k": "v"},
			Metadata:         nil,
			Metrics:          metric.Metrics{},
			Inventory:        inventory.New(),
			Events:           event.Events{},
			lock:             &sync.Mutex{},
		}},
		{"two entries", metric.Dimensions{"k1": "v1", "k2": "v2"}, &Entity{
			CommonDimensions: metric.Dimensions{"k1": "v1", "k2": "v2"},
			Metadata:         nil,
			Metrics:          metric.Metrics{},
			Inventory:        inventory.New(),
			Events:           event.Events{},
			lock:             &sync.Mutex{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got := newHostEntity()
			for k, v := range tt.commons {
				got.AddCommonDimension(k, v)
			}

			assert.Equal(t, tt.expeceted, got)
		})
	}
}
