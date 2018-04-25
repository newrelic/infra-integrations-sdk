package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"

	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/pkg/errors"
)

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	locker             sync.Locker
	storer             persist.Storer
	Name               string    `json:"name"`
	ProtocolVersion    string    `json:"protocol_version"`
	IntegrationVersion string    `json:"integration_version"`
	Data               []*Entity `json:"data"`
	prettyOutput       bool
	writer             io.Writer
}

// New creates new integration with sane default values.
func New(name, version string) (*Integration, error) {
	return NewBuilder(name, version).Build()
}

// NewWithArgs creates new integration with given arguments and sane default values.
func NewWithArgs(name, version string, args interface{}) (*Integration, error) {
	return NewBuilder(name, version).ParsedArguments(args).Build()
}

// Entity method creates or retrieves an already created EntityData.
func (i *Integration) Entity(entityName, entityType string) (e *Entity, err error) {
	i.locker.Lock()
	defer i.locker.Unlock()

	for _, e = range i.Data {
		if e.Metadata.Name == entityName && e.Metadata.Type == entityType {
			return e, nil
		}
	}

	e, err = newEntity(entityName, entityType, i.storer)
	if err != nil {
		return nil, err
	}

	i.Data = append(i.Data, e)

	return e, nil
}

// Publish runs all necessary tasks before publishing the data. Currently, it
// stores the Storer, prints the JSON representation of the integration using a writer (stdout by default)
// and re-initializes the integration object (allowing re-use it during the
// execution of your code).
func (i *Integration) Publish() error {
	if i.storer != nil {
		if err := i.storer.Save(); err != nil {
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
	i.Data = []*Entity{} // empty array preferred instead of null on marshaling.
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
