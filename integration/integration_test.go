package integration

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	"io/ioutil"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/stretchr/testify/assert"
)

func TestIntegration_DefaultEntity(t *testing.T) {
	i, err := newTestingIntegration()
	assert.NoError(t, err)

	e1 := i.DefaultEntity()
	e2 := i.DefaultEntity()
	assert.Equal(t, e1, e2)
}

func TestBuilder_Build(t *testing.T) {
	i, err := newTestingIntegration()
	assert.NoError(t, err)

	if i.Name != "TestIntegration" {
		t.Error()
	}
	if i.IntegrationVersion != "1.0" {
		t.Error()
	}
	if i.ProtocolVersion != "2" {
		t.Error()
	}
	if len(i.Entities) != 0 {
		t.Error()
	}
}

func TestBuilder_BuildWithDefaultArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	// Needed for initialising os.Args + flags (emulating).
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	var al argumentList
	i, err := NewBuilder("TestIntegration", "1.0").Logger(log.Discard).Writer(ioutil.Discard).ParsedArguments(&al).Build()
	if err != nil {
		t.Fatal(err)
	}
	if i.Name != "TestIntegration" {
		t.Error()
	}
	if i.IntegrationVersion != "1.0" {
		t.Error()
	}
	if i.ProtocolVersion != "2" {
		t.Error()
	}
	if len(i.Entities) != 0 {
		t.Error()
	}
	if !al.All {
		t.Error()
	}
	if al.Pretty {
		t.Error()
	}
	if al.Verbose {
		t.Error()
	}
}

func TestBuilder_BuildWithCustomArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	// Needed for initialising os.Args + flags (emulating).
	os.Args = []string{"cmd", "--pretty", "--verbose", "--all"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	var al argumentList
	_, err := NewBuilder("TestIntegration", "1.0").Logger(log.Discard).Writer(ioutil.Discard).ParsedArguments(&al).Build()
	if err != nil {
		t.Fatal(err)
	}
	if !al.All {
		t.Error()
	}
	if !al.Pretty {
		t.Error()
	}
	if !al.Verbose {
		t.Error()
	}
}

func TestIntegration_Publish(t *testing.T) {
	w := testWritter{
		func(p []byte) {
			expectedOutputRaw := []byte(
				`{"name":"TestIntegration","protocol_version":"2","integration_version":"1.0","data":[` +
					`{"entity":{"name":"EntityOne","type":"test"},"metrics":[{"event_type":"EventTypeForEntityOne","metricOne":99,"metricThree":"test","metricTwo":88}],"inventory":{},` +
					`"events":[{"summary":"evnt1sum","category":"evnt1cat"},{"summary":"evnt2sum","category":"evnt2cat"}]},` +
					`{"entity":{"name":"EntityTwo","type":"test"},"metrics":[{"event_type":"EventTypeForEntityTwo","metricOne":99,"metricThree":"test","metricTwo":88}],"inventory":{},` +
					`"events":[]},` +
					`{"entity":{},"metrics":[{"event_type":"EventTypeForEntityThree","metricOne":99,"metricThree":"test","metricTwo":88}],"inventory":{},` +
					`"events":[{"summary":"evnt3sum","category":"evnt3cat"}]}]}`)
			expectedOutput := new(Integration)
			err := json.Unmarshal(expectedOutputRaw, expectedOutput)
			if err != nil {
				t.Fatal("error unmarshaling expected output raw test data sample")
			}

			integration := new(Integration)
			err = json.Unmarshal(p, integration)
			if err != nil {
				t.Error("error unmarshaling integration output", err)
			}

			if !reflect.DeepEqual(expectedOutput, integration) {
				t.Errorf("output does not match the expectations.\nGot:\n%v\nExpected:\n%v", string(p), string(expectedOutputRaw))
			}
		},
	}

	i, err := NewBuilder("TestIntegration", "1.0").Logger(log.Discard).Writer(w).Build()
	if err != nil {
		t.Fatal(err)
	}

	e, err := i.Entity("EntityOne", "test")
	if err != nil {
		t.Fatal(err)
	}

	ms, err := e.NewMetricSet("EventTypeForEntityOne")
	assert.NoError(t, err)
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e.AddEvent(metric.NewEvent("evnt1sum", "evnt1cat"))
	e.AddEvent(metric.NewEvent("evnt2sum", "evnt2cat"))

	e, err = i.Entity("EntityTwo", "test")
	if err != nil {
		t.Fatal(err)
	}

	ms, err = e.NewMetricSet("EventTypeForEntityTwo")
	assert.NoError(t, err)
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e, err = i.Entity("", "")
	if err != nil {
		t.Fatal(err)
	}

	ms, err = e.NewMetricSet("EventTypeForEntityThree")
	assert.NoError(t, err)
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e.AddEvent(metric.NewEvent("evnt3sum", "evnt3cat"))

	i.Publish()
}

func TestIntegration_EntityReturnsExistingEntity(t *testing.T) {
	i, err := newTestingIntegration()
	assert.NoError(t, err)

	e1, err := i.Entity("Entity1", "test")
	if err != nil {
		t.Fail()
	}

	e2, err := i.Entity("Entity1", "test")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, e1, e2)

	if len(i.Entities) > 1 {
		t.Error()
	}
}

// NOTE: This test does nothing as test but when running with -race flag we can detect data races.
// See Lock and Unlock on Entity method.
func TestIntegration_EntityHasNoDataRace(t *testing.T) {
	in, err := NewBuilder("TestIntegration", "1.0").Logger(log.Discard).Writer(ioutil.Discard).Synchronized().Build()
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		go func(i int) {
			in.Entity(fmt.Sprintf("entity%v", i), "test")
		}(i)
	}
}

type testWritter struct {
	testFunc func([]byte)
}

func (w testWritter) Write(p []byte) (n int, err error) {
	w.testFunc(p)

	return len(p), err
}

func newTestingIntegration() (*Integration, error) {
	return NewBuilder("TestIntegration", "1.0").Logger(log.Discard).Writer(ioutil.Discard).Build()
}
