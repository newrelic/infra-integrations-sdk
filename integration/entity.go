package integration

import (
	"encoding/json"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Metadata  EntityMetadata `json:"entity"`
	Metrics   []metric.Set   `json:"metrics"`
	Inventory Inventory      `json:"inventory"`
	Events    []Event        `json:"events"`
	storer    persist.Storer
}

// EntityMetadata stores entity Metadata
type EntityMetadata struct {
	Name string `json:"name"`
	Type string `json:"type""`
}

// MarshalJSON implements json.Marshaler
func (e *Entity) MarshalJSON() ([]byte, error) {
	return json.Marshal(*e)
}

type inventoryItem map[string]interface{}

// Inventory is the data type for inventory data produced by an integration data
// source and emitted to the agent's inventory data store.
type Inventory map[string]inventoryItem

// SetItem stores a value into the inventory data structure.
func (i Inventory) SetItem(key string, field string, value interface{}) {
	if _, ok := i[key]; ok {
		i[key][field] = value
	} else {
		i[key] = inventoryItem{field: value}
	}

}

const defaultEventCategory = "notifications"

// Event is the data type to represent arbitrary, one-off messages for key
// activities on a system.
type Event struct {
	Summary  string `json:"summary"`
	Category string `json:"category,omitempty"`
}

// NewEvent creates a new event.
func NewEvent(summary, category string) *Event {
	return &Event{
		Summary:  summary,
		Category: category,
	}
}

// NewNotification creates a new notification event.
func NewNotification(summary string) *Event {
	return NewEvent(summary, defaultEventCategory)
}

// NewEntity creates a new remote-entity.
func NewEntity(entityName, entityType string) (Entity, error) {
	// If one of the attributes is defined, both Name and Type are needed.
	if entityName == "" && entityType != "" || entityName != "" && entityType == "" {
		return Entity{}, errors.New("entity name and type are required when defining one")
	}

	d := Entity{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []metric.Set{},
		Inventory: make(Inventory),
		Events:    []Event{},
	}

	// Entity data is optional. When not specified, data from the integration is reported for the agent's own entity.
	if entityName != "" && entityType != "" {
		d.Metadata = EntityMetadata{
			Name: entityName,
			Type: entityType,
		}
	}

	return d, nil
}

// NewMetricSet returns a new instance of Set with its sample attached to the integration.
func (e *Entity) NewMetricSet(eventType string) *metric.Set {
	e.Metrics = append(e.Metrics, *metric.NewSet(eventType, e.storer))

	return metric.NewSet(eventType, e.storer)
}

// AddNotificationEvent method adds a new Event with default event category.
func (e *Entity) AddNotificationEvent(summary string) error {
	return e.AddEvent(Event{Summary: summary, Category: defaultEventCategory})
}

// AddEvent method adds a new Event.
func (e *Entity) AddEvent(event Event) error {
	if event.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	e.Events = append(e.Events, event)
	return nil
}
