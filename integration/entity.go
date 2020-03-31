package integration

import (
	"errors"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/data/metadata"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Common    *Common              `json:"common"`
	Metadata  *metadata.Metadata   `json:"entity,omitempty"`
	Metrics   []*metric.Set        `json:"metrics"`
	Inventory *inventory.Inventory `json:"inventory"`
	Events    []*event.Event       `json:"events"`
	storer    persist.Storer
	lock      sync.Locker
}

// Common is a common set of attributes
type Common struct{}

// SameAs return true when is same entity
func (e *Entity) SameAs(b *Entity) bool {
	if e.Metadata == nil || b.Metadata == nil {
		return false
	}

	return e.Metadata.EqualsTo(b.Metadata)
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string, nameSpacingAttributes ...attribute.Attribute) *metric.Set {

	s := metric.NewSet(eventType, e.storer, nameSpacingAttributes...)

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Metrics = append(e.Metrics, s)
	return s
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event *event.Event) error {
	if event.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	e.lock.Lock()
	defer e.lock.Unlock()
	e.Events = append(e.Events, event)
	return nil
}

// SetInventoryItem method sets the inventory item (only one allowed).
func (e *Entity) SetInventoryItem(key string, field string, value interface{}) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Inventory.SetItem(key, field, value)
}

// Tags return the Entity tags
func (e *Entity) Tags() metadata.Tags {
	return e.Metadata.Tags
}

// AddTags adds tags to the entity.
func (e *Entity) AddTags(tags ...metadata.Tag) {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, a := range tags {
		e.AddTag(a.Key, a.Value)
	}
}

// AddTag adds a new tag to the entity
func (e *Entity) AddTag(key string, value string) {
	e.Metadata.AddTag(key, value)
}

// Name is the unique entity identifier within a New Relic customer account.
func (e *Entity) Name() string {
	return e.Metadata.Name
}

//--- private

// newAnonymousEntity creates a entity without metadata.
func newAnonymousEntity(storer persist.Storer, addHostnameToMetadata bool) *Entity {
	return &Entity{
		Common:   &Common{},
		Metadata: nil,
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []*metric.Set{},
		Inventory: inventory.New(),
		Events:    []*event.Event{},
		storer:    storer,
		lock:      &sync.Mutex{},
	}
}

// isAnonymousEntity returns true if entity has no metadata
func (e *Entity) isAnonymousEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// newEntity creates a new entity with with metadata.
func newEntity(
	name,
	displayName string,
	entityType string,
	storer persist.Storer,
	tags ...metadata.Tag,
) (*Entity, error) {

	if name == "" || entityType == "" {
		return nil, errors.New("entity name and type cannot be empty")
	}

	if len(tags) == 0 {
		tags = metadata.Tags{}
	}

	d := &Entity{
		// empty array or object preferred instead of null on marshaling.
		Common: &Common{},
		Metadata: &metadata.Metadata{
			Name:        name,
			DisplayName: displayName,
			EntityType:  entityType,
			Tags:        tags,
		},
		Metrics:   []*metric.Set{},
		Inventory: inventory.New(),
		Events:    []*event.Event{},
		storer:    storer,
		lock:      &sync.Mutex{},
	}

	return d, nil
}
