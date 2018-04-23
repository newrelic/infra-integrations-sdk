package integration

import (
	"encoding/json"
	"io"
	"sync"

	"bytes"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Name string `json:"name"`
	Type string `json:"type"`
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

// EntityData defines all the data related to a particular event from an entity.
type EntityData struct {
	storer    persist.Storer
	Entity    Entity       `json:"entity"`
	Metrics   []metric.Set `json:"metrics"`
	Inventory Inventory    `json:"inventory"`
	Events    []Event      `json:"events"`
}

// NewEntityData creates a new EntityData with default values initialised.
// TODO: do it private
func NewEntityData(entityName, entityType string) (EntityData, error) {
	// If one of the attributes is defined, both Name and Type are needed.
	if entityName == "" && entityType != "" || entityName != "" && entityType == "" {
		return EntityData{}, errors.New("entity name and type are required when defining one")
	}

	d := EntityData{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []metric.Set{},
		Inventory: make(Inventory),
		Events:    []Event{},
	}

	// Entity data is optional. When not specified, data from the integration is reported for the agent's own entity.
	if entityName != "" && entityType != "" {
		d.Entity = Entity{
			Name: entityName,
			Type: entityType,
		}
	}

	return d, nil
}

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	locker             sync.Locker
	Storer             persist.Storer `json:"-"`
	Name               string         `json:"name"`
	ProtocolVersion    string         `json:"protocol_version"`
	IntegrationVersion string         `json:"integration_version"`
	Data               []*EntityData  `json:"data"`
	prettyOutput       bool
	writer             io.Writer
}

// Entity method creates or retrieves an already created EntityData.
func (i *Integration) Entity(entityName, entityType string) (*EntityData, error) {
	i.locker.Lock()
	defer i.locker.Unlock()
	for _, e := range i.Data {
		if e.Entity.Name == entityName && e.Entity.Type == entityType {
			return e, nil
		}
	}

	d, err := NewEntityData(entityName, entityType)
	if err != nil {
		return nil, err
	}

	i.Data = append(i.Data, &d)

	return &d, nil
}

// NewMetricSet returns a new instance of Set with its sample attached to
// the IntegrationData.
func (d *EntityData) NewMetricSet(eventType string) *metric.Set {
	d.Metrics = append(d.Metrics, *metric.NewSet(eventType, d.storer))

	return metric.NewSet(eventType, d.storer)
}

// AddNotificationEvent method adds a new Event with default event category.
func (d *EntityData) AddNotificationEvent(summary string) error {
	return d.AddEvent(Event{Summary: summary, Category: defaultEventCategory})
}

// AddEvent method adds a new Event.
func (d *EntityData) AddEvent(e Event) error {
	if e.Summary == "" {
		return errors.New("summary of the event cannot be empty")
	}

	d.Events = append(d.Events, e)
	return nil
}

// Publish runs all necessary tasks before publishing the data. Currently, it
// stores the Storer, prints the JSON representation of the integration using a writer (stdout by default)
// and re-initializes the integration object (allowing re-use it during the
// execution of your code).
func (i *Integration) Publish() error {
	if i.Storer != nil {
		if err := i.Storer.Save(); err != nil {
			return err
		}
	}

	output, err := i.toJSON(i.prettyOutput)
	if err != nil {
		return err
	}

	_, err = i.writer.Write(output)
	defer i.Clear()

	return err
}

// Clear re-initializes the Inventory, Metrics and Events for this integration.
// Used after publishing so the object can be reused.
func (i *Integration) Clear() {
	i.locker.Lock()
	defer i.locker.Unlock()
	i.Data = []*EntityData{} // empty array preferred instead of null on marshaling.
}

// MarshalJSON serializes integration to JSON, fulfilling Marshaler interface.
func (i *Integration) MarshalJSON() (output []byte, err error) {
	output, err = json.Marshal(*i)
	if err != nil {
		err = errors.Wrap(err, "error marshalling to JSON")
	}

	return
}

// toJSON serializes integration as JSON. If the pretty attribute is
// set to true, the JSON will be indented for easy reading.
func (i *Integration) toJSON(pretty bool) (output []byte, err error) {
	output, err = i.MarshalJSON()
	if !pretty {
		return
	}

	var buf bytes.Buffer
	err = json.Indent(&buf, output, "", "\t")
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
