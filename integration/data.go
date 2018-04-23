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

type EntityMetadata struct {
	Name string `json:"name"`
	Type string `json:"type""`
}

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

// NewEntity creates a new EntityData with default values initialised.
// TODO: do it private
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

// NewMetricSet returns a new instance of Set with its sample attached to
// the IntegrationData.
func (d *Entity) NewMetricSet(eventType string) *metric.Set {
	d.Metrics = append(d.Metrics, *metric.NewSet(eventType, d.storer))

	return metric.NewSet(eventType, d.storer)
}

// AddNotificationEvent method adds a new Event with default event category.
func (d *Entity) AddNotificationEvent(summary string) error {
	return d.AddEvent(Event{Summary: summary, Category: defaultEventCategory})
}

// AddEvent method adds a new Event.
func (d *Entity) AddEvent(e Event) error {
	if e.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	d.Events = append(d.Events, e)
	return nil
}
