package integration

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"

	"io/ioutil"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/metric"
	"github.com/stretchr/testify/assert"
)

var (
	integrationName    = "TestIntegration"
	integrationVersion = "1.0"
)

func TestCreation(t *testing.T) {
	i := newNoLoggerNoWriter(t)

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

func TestDefaultIntegrationWritesToStdout(t *testing.T) {
	// capture stdout to file
	f, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)
	back := os.Stdout
	defer func() {
		os.Stdout = back
	}()
	os.Stdout = f

	i, err := New("integration", "4.0")

	assert.NoError(t, err)
	assert.Equal(t, "integration", i.Name)
	assert.Equal(t, "4.0", i.IntegrationVersion)
	assert.Equal(t, "2", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Entities))

	assert.NoError(t, i.Publish())

	// integration published metadata to stdout
	f.Close()
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"integration","protocol_version":"2","integration_version":"4.0","data":[]}`, string(payload))
}

func TestIntegration_DefaultEntity(t *testing.T) {
	i := newNoLoggerNoWriter(t)

	e1 := i.DefaultEntity()
	e2 := i.DefaultEntity()
	assert.Equal(t, e1, e2)
}

func TestDefaultArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	var al argumentList
	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
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

func TestCustomArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	os.Args = []string{"cmd", "--pretty", "--verbose", "--all"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	var al argumentList
	_, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

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
					`{"metrics":[{"event_type":"EventTypeForEntityThree","metricOne":99,"metricThree":"test","metricTwo":88}],"inventory":{},` +
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

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(w))
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
	i := newNoLoggerNoWriter(t)

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

func TestIntegration_LoggerReturnsDefaultLogger(t *testing.T) {
	i, err := New(integrationName, integrationVersion)
	assert.NoError(t, err)

	assert.Equal(t, i.logger, i.Logger())
}

func TestIntegration_LoggerReturnsInjectedInstance(t *testing.T) {
	l := log.NewStdErr(false)

	i, err := New(integrationName, integrationVersion, Logger(l))
	assert.NoError(t, err)

	assert.Equal(t, l, i.Logger())
}

func newNoLoggerNoWriter(t *testing.T) *Integration {
	i, err := New(integrationName, integrationVersion, Logger(log.Discard), Writer(ioutil.Discard))
	assert.NoError(t, err)

	return i
}

type testWritter struct {
	testFunc func([]byte)
}

func (w testWritter) Write(p []byte) (n int, err error) {
	w.testFunc(p)

	return len(p), err
}
