package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/data/metric"

	"github.com/newrelic/infra-integrations-sdk/v4/args"
	"github.com/newrelic/infra-integrations-sdk/v4/log"
)

// Custom attribute keys:
const (
	CustomAttrPrefix = "NRI_"
)

// NR infrastructure agent protocol version
const (
	protocolVersion = "4"
)

// Metadata describes the integration
type Metadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Integration defines the format of the output JSON that integrations will return for protocol 2.
type Integration struct {
	ProtocolVersion string    `json:"protocol_version"`
	Metadata        Metadata  `json:"integration"`
	Entities        []*Entity `json:"data"`
	// HostEntity is an "entity" that serves as dumping ground for metrics not associated with a specific entity
	HostEntity   *Entity `json:"-"` //skip json serializing
	locker       sync.Locker
	prettyOutput bool
	writer       io.Writer
	logger       log.Logger
	args         interface{}
}

// New creates new integration with sane default values.
func New(name, version string, opts ...Option) (i *Integration, err error) {

	if name == "" {
		err = errors.New("integration name cannot be empty")
		return
	}

	if version == "" {
		err = errors.New("integration version cannot be empty")
		return
	}

	i = &Integration{

		ProtocolVersion: protocolVersion,
		Metadata:        Metadata{name, version},
		Entities:        []*Entity{},
		writer:          os.Stdout,
		locker:          &sync.Mutex{},
	}

	for _, opt := range opts {
		if err = opt(i); err != nil {
			err = fmt.Errorf("error applying option to integration. %s", err)
			return
		}
	}

	// arguments
	if err = i.checkArguments(); err != nil {
		return
	}
	if err = args.SetupArgs(i.args); err != nil {
		return
	}
	defaultArgs := args.GetDefaultArgs(i.args)
	i.prettyOutput = defaultArgs.Pretty

	if defaultArgs.Verbose {
		log.SetupLogging(defaultArgs.Verbose)
	}

	// Setting default values, if not set yet
	if i.logger == nil {
		i.logger = log.NewStdErr(defaultArgs.Verbose)
	}

	i.HostEntity = newHostEntity()

	return
}

// NewEntity method creates a new (uniquely named) Entity.
// The `name` of the Entity must be unique for the account otherwise it will cause conflicts
func (i *Integration) NewEntity(name string, entityType string, displayName string) (e *Entity, err error) {
	i.locker.Lock()
	defer i.locker.Unlock()

	e, err = newEntity(name, entityType, displayName)
	if err != nil {
		return nil, err
	}

	err = i.addDefaultAttributes(e)

	return e, err
}

// AddEntity adds an entity to the list of entities. No check for "duplicates" is performed
func (i *Integration) AddEntity(e *Entity) {
	i.Entities = append(i.Entities, e)
}

// Publish writes the data to output (stdout) and resets the integration "object"
func (i *Integration) Publish() error {
	defer i.Clear()

	// add the host entity to the list of entities to be serialized, if not empty
	if notEmpty(i.HostEntity) {
		i.Entities = append(i.Entities, i.HostEntity)
	}
	output, err := i.toJSON(i.prettyOutput)
	if err != nil {
		return err
	}
	output = append(output, []byte{'\n'}...)
	_, err = i.writer.Write(output)

	return err
}

// Clear re-initializes the Inventory, Metrics and Events for this integration.
// Used after publishing so the object can be reused.
func (i *Integration) Clear() {
	i.locker.Lock()
	defer i.locker.Unlock()
	i.Entities = []*Entity{} // empty array preferred instead of null on marshaling.
	// reset the host entity
	i.HostEntity = newHostEntity()
}

// MarshalJSON serializes integration to JSON, fulfilling Marshaler interface.
func (i *Integration) MarshalJSON() (output []byte, err error) {
	output, err = json.Marshal(*i)
	if err != nil {
		err = fmt.Errorf("error marshalling to JSON: %s", err)
	}

	return
}

// toJSON serializes integration as JSON. If the pretty attribute is
// set to true, the JSON will be indented for easy reading.
func (i *Integration) toJSON(pretty bool) (output []byte, err error) {
	if !pretty {
		return i.MarshalJSON()
	}

	return json.MarshalIndent(*i, "", "\t")
}

// Logger returns the integration logger instance.
func (i *Integration) Logger() log.Logger {
	return i.logger
}

// FindEntity finds ad return an entity by name. returns false if entity does not exist in the integration
func (i *Integration) FindEntity(name string) (*Entity, bool) {
	if i.Entities == nil {
		return &Entity{}, false
	}
	for _, e := range i.Entities {
		// nil metadata usually means the "host" entity, so keep searching
		if e == nil || e.Metadata == nil {
			continue
		}

		if e.Metadata.Name == name {
			return e, true
		}
	}
	return &Entity{}, false
}

// Gauge creates a metric of type gauge
func Gauge(timestamp time.Time, metricName string, value float64) (metric.Metric, error) {
	return metric.NewGauge(timestamp, metricName, value)
}

// Count creates a metric of type count
func Count(timestamp time.Time, metricName string, value float64) (metric.Metric, error) {
	return metric.NewCount(timestamp, metricName, value)
}

// Summary creates a metric of type summary
func Summary(timestamp time.Time, metricName string, count float64,
	average float64, sum float64, min float64, max float64) (metric.Metric, error) {
	return metric.NewSummary(timestamp, metricName, count, average, sum, min, max)
}

// CumulativeCount creates metric of type cumulative count
func CumulativeCount(timestamp time.Time, metricName string, value float64) (metric.Metric, error) {
	return metric.NewCumulativeCount(timestamp, metricName, value)
}

// Rate creates a metric of type rate
func Rate(timestamp time.Time, metricName string, value float64) (metric.Metric, error) {
	return metric.NewRate(timestamp, metricName, value)
}

// CumulativeRate creates a metric of type cumulative rate
func CumulativeRate(timestamp time.Time, metricName string, value float64) (metric.Metric, error) {
	return metric.NewCumulativeRate(timestamp, metricName, value)
}

// PrometheusHistogram creates a metric of type prometheus histogram
func PrometheusHistogram(timestamp time.Time, metricName string, sampleCount uint64, sampleSum float64) (*metric.PrometheusHistogram, error) {
	return metric.NewPrometheusHistogram(timestamp, metricName, sampleCount, sampleSum)
}

// PrometheusSummary creates a metric of type prometheus summary
func PrometheusSummary(timestamp time.Time, metricName string, sampleCount uint64, sampleSum float64) (*metric.PrometheusSummary, error) {
	return metric.NewPrometheusSummary(timestamp, metricName, sampleCount, sampleSum)
}

// -- private
// is entity empty?
func notEmpty(entity *Entity) bool {
	return len(entity.Events) > 0 || len(entity.Metrics) > 0 || entity.Inventory.Len() > 0
}

func (i *Integration) checkArguments() error {
	if i.args == nil {
		i.args = new(struct{})
		return nil
	}
	val := reflect.ValueOf(i.args)

	if val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Struct {
		return nil
	}

	return errors.New("arguments must be a pointer to a struct (or nil)")
}

func (i *Integration) addDefaultAttributes(e *Entity) error {
	defaultArgs := args.GetDefaultArgs(i.args)

	// get env vars values for "custom" prefixed vars (NRIA_) and add them as attributes to the entity
	if defaultArgs.Metadata {
		for _, element := range os.Environ() {
			variable := strings.Split(element, "=")
			prefix := fmt.Sprintf("%s%s_", CustomAttrPrefix, strings.ToUpper(i.Metadata.Name))
			if strings.HasPrefix(variable[0], prefix) {
				err := e.AddTag(strings.TrimPrefix(variable[0], prefix), variable[1])
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
