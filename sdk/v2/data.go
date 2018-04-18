package v2

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/newrelic/infra-integrations-sdk/sdk/v1"
	"github.com/pkg/errors"
)

// Entity is the producer of the data. Entity could be a host, a container, a pod, or whatever unit of meaning.
type Entity struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Inventory is the data type for inventory data produced by an integration data
// source and emitted to the agent's inventory data store.
type Inventory v1.Inventory

// Event is the data type to represent arbitrary, one-off messages for key
// activities on a system.
type Event v1.Event

// EntityData defines all the data related to a particular event from an entity.
type EntityData struct {
	Entity    Entity       `json:"entity"`
	Metrics   []metric.Set `json:"metrics"`
	Inventory Inventory    `json:"inventory"`
	Events    []Event      `json:"events"`
}

// NewEntityData creates a new EntityData with default values initialised.
func NewEntityData(entityName, entityType string) (EntityData, error) {
	// If one of the attributes is defined, both Name and Type are needed.
	if entityName == "" && entityType != "" || entityName != "" && entityType == "" {
		return EntityData{}, errors.New("entity name and type are required when defining one")
	}

	d := EntityData{
		// empty array or object preferred instead of null on marshaling.
		Metrics:   []metric.Set{},
		Inventory: Inventory{},
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
func (integration *Integration) Entity(entityName, entityType string) (*EntityData, error) {
	integration.locker.Lock()
	defer integration.locker.Unlock()
	for _, e := range integration.Data {
		if e.Entity.Name == entityName && e.Entity.Type == entityType {
			return e, nil
		}
	}

	d, err := NewEntityData(entityName, entityType)
	if err != nil {
		return nil, err
	}

	integration.Data = append(integration.Data, &d)

	return &d, nil
}

// NewMetricSet returns a new instance of Set with its sample attached to
// the IntegrationData.
func (d *EntityData) NewMetricSet(eventType string) metric.Set {
	ms := metric.NewSet(eventType)
	d.Metrics = append(d.Metrics, ms)

	return ms
}

// AddNotificationEvent method adds a new Event with default event category.
func (d *EntityData) AddNotificationEvent(summary string) error {
	return d.AddEvent(Event{Summary: summary, Category: v1.DefaultEventCategory})
}

// AddEvent method adds a new Event.
func (d *EntityData) AddEvent(e Event) error {
	if e.Summary == "" {
		return fmt.Errorf("summary of the event cannot be empty")
	}

	d.Events = append(d.Events, e)
	return nil
}

// Publish runs all necessary tasks before publishing the data. Currently, it
// stores the Storer, prints the JSON representation of the integration using a writer (stdout by default)
// and re-initializes the integration object (allowing re-use it during the
// execution of your code).
func (integration *Integration) Publish() error {
	if integration.Storer != nil {
		if err := integration.Storer.Save(); err != nil {
			return err
		}
	}

	output, err := integration.toJSON(integration.prettyOutput)
	if err != nil {
		return err
	}

	fmt.Fprint(integration.writer, output)
	integration.Clear()

	return nil
}

// Clear re-initializes the Inventory, Metrics and Events for this integration.
// Used after publishing so the object can be reused.
func (integration *Integration) Clear() {
	integration.locker.Lock()
	defer integration.locker.Unlock()
	integration.Data = []*EntityData{} // empty array preferred instead of null on marshaling.
}

// toJSON returns the integration as a JSON string. If the pretty attribute is
// set to true, the JSON will be idented for easy reading.
func (integration *Integration) toJSON(pretty bool) (string, error) {
	var output []byte
	var err error

	if pretty {
		output, err = json.MarshalIndent(integration, "", "\t")
	} else {
		output, err = json.Marshal(integration)
	}

	if err != nil {
		return "", fmt.Errorf("error marshalling to JSON: %s", err)
	}

	if string(output) == "null" {
		return "[]", nil
	}

	return string(output), nil
}
