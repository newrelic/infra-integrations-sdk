package sdk

import (
	"encoding/json"
	"fmt"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/cache"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
)

// Inventory is the data type for inventory data produced by a plugin data
// source and emitted to the agent's inventory data store
type Inventory map[string]interface{}

// Event is the data type for single shot events
type Event map[string]interface{}

// Integration defines the format of the output JSON that plugins will return
type Integration struct {
	Name               string               `json:"name"`
	ProtocolVersion    string               `json:"protocol_version"`
	IntegrationVersion string               `json:"integration_version"`
	Metrics            []*metric.MetricSet  `json:"metrics"`
	Inventory          map[string]Inventory `json:"inventory"`
	Events             []Event              `json:"events"`
	prettyOutput       bool
}

// NewIntegration initializes a new instance of integration data
func NewIntegration(name string, version string, arguments interface{}) (*Integration, error) {
	err := args.SetupArgs(arguments)
	if err != nil {
		return nil, err
	}
	defaultArgs := args.GetDefaultArgs(arguments)

	log.SetupLogging(defaultArgs.Verbose)

	// Avoid working with an uninitialized or in error state cache
	if err = cache.Status(); err != nil {
		return nil, err
	}

	integration := &Integration{
		Name:               name,
		ProtocolVersion:    "1",
		IntegrationVersion: version,
		Inventory:          make(map[string]Inventory),
		Metrics:            make([]*metric.MetricSet, 0),
		Events:             make([]Event, 0),
		prettyOutput:       defaultArgs.Pretty,
	}

	return integration, nil
}

// NewMetricSet returns a new instance of MetricSet with its sample attached to the IntegrationData
func (integration *Integration) NewMetricSet(eventType string, provider string) *metric.MetricSet {
	ms := metric.NewMetricSet(eventType, provider)
	integration.Metrics = append(integration.Metrics, &ms)
	return &ms
}

// Publish will run any tasks before publishing the data. In this case, it will
// store the cache and print the JSON repreentation of the integration to stdout
func (integration *Integration) Publish() error {
	if err := cache.Save(); err != nil {
		return err
	}

	output, err := integration.toJSON(integration.prettyOutput)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
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
		return "", fmt.Errorf("Error marshalling to JSON: %s", err)
	}

	if string(output) == "null" {
		return "[]", nil
	}

	return string(output), nil
}
