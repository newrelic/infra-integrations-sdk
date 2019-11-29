package integration

import (
	"testing"

	"strconv"
	"sync"

	"encoding/json"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEntity(t *testing.T) {
	e, err := newEntity("name", "type", persist.NewInMemoryStore(), false)

	assert.NoError(t, err)
	assert.Equal(t, "name", e.Metadata.Name)
	assert.Equal(t, "type", e.Metadata.Namespace)
	assert.False(t, e.AddHostname)
	assert.Empty(t, e.Metadata.IDAttrs)
}

func TestNewEntityWithIdentifierAttributes(t *testing.T) {
	attr1 := NewIDAttribute("env", "prod")
	attr2 := NewIDAttribute("srv", "auth")
	e, err := newEntity(
		"name",
		"type",
		persist.NewInMemoryStore(),
		true,
		attr1,
		attr2,
	)

	assert.NoError(t, err)
	assert.True(t, e.AddHostname)
	assert.Len(t, e.Metadata.IDAttrs, 2)
	assert.Equal(t, e.Metadata.IDAttrs[0], attr1)
	assert.Equal(t, e.Metadata.IDAttrs[1], attr2)
}

func TestNewEntityWithOneIdentifierAttribute(t *testing.T) {
	attr1 := NewIDAttribute("env", "prod")
	e, err := newEntity(
		"name",
		"type",
		persist.NewInMemoryStore(),
		true,
		attr1,
	)

	assert.NoError(t, err)
	assert.True(t, e.AddHostname)
	assert.Len(t, e.Metadata.IDAttrs, 1)
	assert.Equal(t, e.Metadata.IDAttrs[0], attr1)
}

func TestEntity_AddAttributes(t *testing.T) {
	idAttr := NewIDAttribute("env", "prod")
	e, err := newEntity(
		"name",
		"type",
		persist.NewInMemoryStore(),
		false,
		idAttr,
	)
	assert.NoError(t, err)

	e.AddAttributes(attribute.Attr("key1", "val1"), attribute.Attr("key2", "val2"))

	assert.Len(t, e.customAttributes, 2, "attributes should have been added to the entity")

	ms := e.NewMetricSet("event-type")

	assert.Equal(t, "val1", ms.Metrics["key1"])
	assert.Equal(t, "val2", ms.Metrics["key2"])
}

func TestEntitiesRequireNameAndType(t *testing.T) {
	_, err := newEntity("", "", nil, false)

	assert.Error(t, err)
}

func TestAddNotificationEvent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
	if err != nil {
		t.Fatal(err)
	}

	en.customAttributes = attribute.Attributes{attribute.Attr("clusterName", "my-cluster-name")}

	err = en.AddEvent(event.NewNotification("TestSummary"))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 1)

	if en.Events[0].Summary != "TestSummary" || en.Events[0].Category != "notifications" {
		t.Error("malformed event")
	}
}

func TestAddEventWithAttributes(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
	require.NoError(t, err)

	en.customAttributes = attribute.Attributes{attribute.Attr("clusterName", "my-cluster-name")}
	attrs := map[string]interface{}{"attrKey": "attrVal"}
	err = en.AddEvent(event.NewWithAttributes("TestSummary", "TestCategory", attrs))
	assert.NoError(t, err)

	require.Len(t, en.Events, 1)

	assert.Equal(t, "TestSummary", en.Events[0].Summary)
	assert.Equal(t, "TestCategory", en.Events[0].Category)

	expectedAttrs := map[string]interface{}{
		"attrKey":     "attrVal",
		"clusterName": "my-cluster-name",
	}
	assert.Equal(t, expectedAttrs, en.Events[0].Attributes)
}

func TestAddNotificationWithEmptySummaryFails(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
	if err != nil {
		t.Fatal(err)
	}

	err = en.AddEvent(event.NewNotification(""))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestAddEvent_Entity(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
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
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	err = en.AddEvent(event.New("TestSummary", ""))
	assert.NoError(t, err)

	assert.Len(t, en.Events, 2)
}

func TestAddEvent_Entity_EmptySummary_Error(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
	assert.NoError(t, err)

	err = en.AddEvent(event.New("", "TestCategory"))
	assert.Error(t, err)

	assert.Len(t, en.Events, 0)
}

func TestEntity_AddInventoryConcurrent(t *testing.T) {
	en, err := newEntity("Entity1", "Type1", persist.NewInMemoryStore(), false)
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
	e := newLocalEntity(persist.NewInMemoryStore(), false)
	j, err := json.Marshal(e)

	assert.NoError(t, err)
	assert.Equal(t, `{"metrics":[],"inventory":{},"events":[]}`, string(j))
}

func TestEntity_IsDefaultEntity(t *testing.T) {
	e := newLocalEntity(persist.NewInMemoryStore(), false)

	assert.Empty(t, e.Metadata, "default entity should have no identifier")
	assert.True(t, e.isLocalEntity())
}

func TestEntity_Key(t *testing.T) {
	e, err := newEntity("entity", "ns", persist.NewInMemoryStore(), false)
	assert.NoError(t, err)

	k, err := e.Key()
	assert.NoError(t, err)
	assert.Equal(t, "ns:entity", k.String())
}

func TestEntity_Key_WithIDAttrs(t *testing.T) {
	attr1 := NewIDAttribute("env", "prod")
	attr2 := NewIDAttribute("srv", "auth")
	e, err := newEntity("entity", "ns", persist.NewInMemoryStore(), false, attr1, attr2)
	assert.NoError(t, err)

	k, err := e.Key()
	assert.NoError(t, err)
	assert.Equal(t, "ns:entity:env=prod:srv=auth", k.String())
}

func TestEntity_SameAs(t *testing.T) {
	attr := NewIDAttribute("env", "prod")
	e1, err := newEntity("entity", "ns", persist.NewInMemoryStore(), false, attr)
	assert.NoError(t, err)

	e2, err := newEntity("entity", "ns", persist.NewInMemoryStore(), false, attr)
	assert.NoError(t, err)

	e3, err := newEntity("entity", "ns", persist.NewInMemoryStore(), false)
	assert.NoError(t, err)

	assert.True(t, e1.SameAs(e2))
	assert.False(t, e1.SameAs(e3))
}
