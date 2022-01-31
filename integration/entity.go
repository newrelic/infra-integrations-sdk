package integration

import (
	"errors"
	"sync"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/data/event"
	"github.com/newrelic/infra-integrations-sdk/v4/data/inventory"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metadata"
	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	CommonDimensions Common               `json:"common"` // dimensions common to every entity metric
	Metadata         *metadata.Metadata   `json:"entity,omitempty"`
	Metrics          metric.Metrics       `json:"metrics"`
	Inventory        *inventory.Inventory `json:"inventory"`
	Events           event.Events         `json:"events"`
	lock             sync.Locker
}

// Common is the producer of the common dimensions/attributes.
type Common struct {
	Timestamp *int64 `json:"timestamp,omitempty"`
	Interval  *int64 `json:"interval.ms,omitempty"`
	// Attributes are optional, they represent additional information that
	// can be attached to an event.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SameAs return true when is same entity
func (e *Entity) SameAs(b *Entity) bool {
	if e.Metadata == nil || b.Metadata == nil {
		return false
	}

	return e.Metadata.EqualsTo(b.Metadata)
}

// AddMetric adds a new metric to the entity metrics list
func (e *Entity) AddMetric(metric metric.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Metrics = append(e.Metrics, metric)
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event *event.Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.Events = append(e.Events, event)
}

// AddInventoryItem method sets the inventory item (only one allowed).
func (e *Entity) AddInventoryItem(key string, field string, value interface{}) error {
	if len(key) == 0 || len(field) == 0 {
		return errors.New("key or field cannot be empty")
	}
	e.lock.Lock()
	defer e.lock.Unlock()
	return e.Inventory.SetItem(key, field, value)
}

// AddCommonDimension adds a new dimension to every metric within the entity.
func (e *Entity) AddCommonDimension(key string, value string) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.CommonDimensions.Attributes[key] = value
}

// AddCommonTimestamp adds a new common timestamp to the entity.
func (e *Entity) AddCommonTimestamp(timestamp time.Time) {
	e.lock.Lock()
	defer e.lock.Unlock()

	t := timestamp.Unix()
	e.CommonDimensions.Timestamp = &t
}

// AddCommonInterval adds a common interval in milliseconds from a time.Duration (nanoseconds).
func (e *Entity) AddCommonInterval(timestamp time.Duration) {
	e.lock.Lock()
	defer e.lock.Unlock()

	t := timestamp.Milliseconds()
	e.CommonDimensions.Interval = &t
}

// GetMetadata returns all the Entity's metadata
func (e *Entity) GetMetadata() metadata.Map {
	return e.Metadata.Metadata
}

// AddTag adds a new tag to the entity
func (e *Entity) AddTag(key string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	e.Metadata.AddTag(key, value)
	return nil
}

// AddMetadata adds a new metadata to the entity
func (e *Entity) AddMetadata(key string, value interface{}) error {
	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}
	e.Metadata.AddMetadata(key, value)
	return nil
}

// Name is the unique entity identifier within a New Relic customer account.
func (e *Entity) Name() string {
	return e.Metadata.Name
}

//--- private

// newHostEntity creates a entity without metadata.
func newHostEntity() *Entity {
	return &Entity{
		CommonDimensions: Common{
			Attributes: make(map[string]interface{}),
		},
		Metadata:  nil,
		Metrics:   metric.Metrics{},
		Inventory: inventory.New(),
		Events:    event.Events{},
		lock:      &sync.Mutex{},
	}
}

// isHostEntity returns true if entity has no metadata
func (e *Entity) isHostEntity() bool {
	return e.Metadata == nil || e.Metadata.Name == ""
}

// newEntity creates a new entity with with metadata.
func newEntity(name, entityType string, displayName string) (*Entity, error) {

	if name == "" || entityType == "" {
		return nil, errors.New("entity name and type cannot be empty")
	}

	e := newHostEntity()
	e.Metadata = metadata.New(name, entityType, displayName)

	return e, nil
}
