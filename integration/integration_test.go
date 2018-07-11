package integration

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"

	"io/ioutil"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/data/event"
	"github.com/newrelic/infra-integrations-sdk/data/metric"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/stretchr/testify/assert"
)

var (
	integrationName    = "TestIntegration"
	integrationVersion = "1.0"
)

func TestCreation(t *testing.T) {
	i := newTestIntegration(t)

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

	i, err := New("integration", "4.0", InMemoryStore())

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
	i := newTestIntegration(t)

	e1 := i.LocalEntity()
	e2 := i.LocalEntity()
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
	if !al.All() {
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

	os.Args = []string{"cmd", "--pretty", "--verbose"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	var al argumentList
	_, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	if !al.All() {
		t.Error()
	}
	if !al.Pretty {
		t.Error()
	}
	if !al.Verbose {
		t.Error()
	}
}

func TestVerboseLog(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	// Given an integration set in verbose mode
	os.Args = []string{"cmd", "--verbose"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Whose log messages are written in the standard error
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	var al argumentList
	i, err := New("TestIntegration", "1.0", Args(&al))
	assert.NoError(t, err)

	// When logging a debug message
	i.logger.Debugf("hello everybody")
	assert.NoError(t, w.Close())

	// The message is correctly submitted to the standard error
	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.Contains(t, stdErrBytes.String(), "hello everybody")
}

func TestNonVerboseLog(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	// Given an integration set in non-verbose mode
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Whose log messages are written in the standard error
	r, w, err := os.Pipe()
	assert.NoError(t, err)
	back := os.Stderr
	os.Stderr = w
	defer func() {
		os.Stderr = back
	}()
	defer r.Close()

	var al argumentList
	i, err := New("TestIntegration", "1.0", Args(&al))
	assert.NoError(t, err)

	// When logging info, error and debug messages
	i.logger.Debugf("this is a debug")
	i.logger.Infof("this is an info")
	i.logger.Errorf("this is an error")
	assert.NoError(t, w.Close())

	// The standard error shows all the message levels but debug
	stdErrBytes := new(bytes.Buffer)
	stdErrBytes.ReadFrom(r)
	assert.Contains(t, stdErrBytes.String(), "this is an error")
	assert.Contains(t, stdErrBytes.String(), "this is an info")
	assert.NotContains(t, stdErrBytes.String(), "this is a debug")
}

func TestIntegration_Publish(t *testing.T) {
	w := testWriter{
		func(p []byte) {
			expectedOutputRaw := []byte(
				`{"name":"TestIntegration","protocol_version":"2","integration_version":"1.0","data":[` +
					`{"entity":{"name":"EntityOne","type":"test"},"metrics":[{"event_type":"EventTypeForEntityOne","metricOne":99,"metricThree":"test","metricTwo":88}],
					     "inventory":{"key1":{"field1":123,"field2":"hello"},"key2":{"field3":"world"}},` +
					`"events":[{"summary":"evnt1sum","category":"evnt1cat"},{"summary":"evnt2sum","category":"evnt2cat"}]},` +
					`{"entity":{"name":"EntityTwo","type":"test"},"metrics":[{"event_type":"EventTypeForEntityTwo","metricOne":99,"metricThree":"test","metricTwo":88}],"inventory":{},` +
					`"events":[]},` +
					`{"metrics":[{"event_type":"EventTypeForEntityThree","metricOne":99,"metricThree":"test","metricTwo":88}],
						"inventory":{"inv":{"key":"val"}},` +
					`"events":[{"summary":"evnt3sum","category":"evnt3cat"}]}]}`)
			expectedOutput := new(Integration)
			err := json.Unmarshal(expectedOutputRaw, expectedOutput)
			assert.NoError(t, err, "error unmarshaling expected output raw test data sample")

			integration := new(Integration)
			err = json.Unmarshal(p, integration)
			if err != nil {
				assert.NoError(t, err, "error unmarshaling integration output")
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

	ms := e.NewMetricSet("EventTypeForEntityOne")
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e.AddEvent(event.New("evnt1sum", "evnt1cat"))
	e.AddEvent(event.New("evnt2sum", "evnt2cat"))

	assert.NoError(t, e.SetInventoryItem("key1", "field1", 123))
	assert.NoError(t, e.SetInventoryItem("key1", "field2", "hello"))
	assert.NoError(t, e.SetInventoryItem("key2", "field3", "world"))

	e, err = i.Entity("EntityTwo", "test")
	if err != nil {
		t.Fatal(err)
	}

	ms = e.NewMetricSet("EventTypeForEntityTwo")
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e, err = i.Entity("", "")
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, e.SetInventoryItem("inv", "key", "value"))

	ms = e.NewMetricSet("EventTypeForEntityThree")
	ms.SetMetric("metricOne", 99, metric.GAUGE)
	ms.SetMetric("metricTwo", 88, metric.GAUGE)
	ms.SetMetric("metricThree", "test", metric.ATTRIBUTE)

	e.AddEvent(event.New("evnt3sum", "evnt3cat"))

	i.Publish()
}

func TestIntegration_EntityReturnsExistingEntity(t *testing.T) {
	i := newTestIntegration(t)

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

func newTestIntegration(t *testing.T) *Integration {
	i, err := New(integrationName, integrationVersion, Logger(log.Discard), Writer(ioutil.Discard), InMemoryStore())
	assert.NoError(t, err)

	return i
}

func TestIntegration_CreateLocalAndRemoteEntities(t *testing.T) {
	i, err := New(integrationName, integrationVersion)
	assert.NoError(t, err)

	local := i.LocalEntity()
	assert.NotEqual(t, local, nil)

	remote, err := i.Entity("Entity1", "test")
	assert.NoError(t, err)
	assert.NotEqual(t, remote, nil)
}

type testWriter struct {
	testFunc func([]byte)
}

func (w testWriter) Write(p []byte) (n int, err error) {
	w.testFunc(p)

	return len(p), err
}
