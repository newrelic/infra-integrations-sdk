package integration

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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
	if i.ProtocolVersion != "3" {
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
	assert.Equal(t, "3", i.ProtocolVersion)
	assert.Equal(t, 0, len(i.Entities))

	assert.NoError(t, i.Publish())

	// integration published metadata to stdout
	f.Close()
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, `{"name":"integration","protocol_version":"3","integration_version":"4.0","data":[]}`+"\n", string(payload))
}

func TestIntegration_DefaultEntity(t *testing.T) {
	i := newTestIntegration(t)

	e1 := i.LocalEntity()
	e2 := i.LocalEntity()
	assert.Equal(t, e1, e2)
}

func TestDefaultArguments(t *testing.T) {
	al := sdk_args.DefaultArgumentList{}

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	if i.Name != "TestIntegration" {
		t.Error()
	}
	if i.IntegrationVersion != "1.0" {
		t.Error()
	}
	if i.ProtocolVersion != "3" {
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

func TestArgumentsSetEntityAttribute(t *testing.T) {
	testCases := []struct {
		envvar    string
		attribute func(*Entity) string
		value     string
	}{
		{"NRI_CLUSTER", func(e *Entity) string { return e.Cluster }, "foo"},
		{"NRI_SERVICE", func(e *Entity) string { return e.Service }, "bar"},
	}

	for _, test := range testCases {
		t.Run(test.envvar, func(t *testing.T) {

			al := sdk_args.DefaultArgumentList{}

			os.Setenv(test.envvar, test.value)
			defer os.Clearenv()
			os.Args = []string{"cmd"}
			flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)

			i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
			assert.NoError(t, err)

			e, err := i.Entity("name", "ns")
			assert.NoError(t, err)

			assert.Equal(t, test.value, test.attribute(e))
		})
	}
}

func TestAddHostnameFlagDecoratesEntities(t *testing.T) {
	al := sdk_args.DefaultArgumentList{}

	os.Args = []string{"cmd", "-nri_add_hostname"}
	flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	assert.True(t, i.LocalEntity().AddHostname)

	e, err := i.Entity("name", "ns")
	assert.NoError(t, err)

	assert.True(t, e.AddHostname)
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
		func(integrationBytes []byte) {
			expectedOutputRaw := []byte(`
			{
			  "name": "TestIntegration",
			  "protocol_version": "3",
			  "integration_version": "1.0",
			  "data": [
				{
				  "entity": {
					"name": "EntityOne",
					"type": "test"
				  },
				  "metrics": [
					{
					  "event_type": "EventTypeForEntityOne",
					  "metricBool": 1,
					  "metricOne": 1,
					  "metricTwo": "test"
					}
				  ],
				  "inventory": {},
				  "events": [
					{
					  "summary": "evnt1sum",
					  "category": "evnt1cat"
					},
					{
					  "summary": "evnt2sum",
					  "category": "evnt2cat"
					}
				  ]
				},
				{
				  "entity": {
					"name": "EntityTwo",
					"type": "test"
				  },
				  "metrics": [
					{
					  "event_type": "EventTypeForEntityTwo",
					  "metricOne": 2
					}
				  ],
				  "inventory": {},
				  "events": []
				},
				{
				  "metrics": [],
                  "inventory":{
			        "inv":{"key":"val"}
                  },
				  "events": []
				}
			  ]
			}`)

			// awful but cannot compare with json.Unmarshal as it's not supported by Integration
			assert.Equal(t, stripBlanks(expectedOutputRaw), stripBlanks(integrationBytes))
		},
	}

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(w))
	assert.NoError(t, err)

	e, err := i.Entity("EntityOne", "test")
	assert.NoError(t, err)
	ms := e.NewMetricSet("EventTypeForEntityOne")
	assert.NoError(t, ms.SetMetric("metricOne", 1, metric.GAUGE))
	assert.NoError(t, ms.SetMetric("metricTwo", "test", metric.ATTRIBUTE))
	assert.NoError(t, ms.SetMetric("metricBool", true, metric.GAUGE))

	assert.NoError(t, e.AddEvent(event.New("evnt1sum", "evnt1cat")))
	assert.NoError(t, e.AddEvent(event.New("evnt2sum", "evnt2cat")))

	e2, err := i.Entity("EntityTwo", "test")
	assert.NoError(t, err)
	ms = e2.NewMetricSet("EventTypeForEntityTwo")
	assert.NoError(t, ms.SetMetric("metricOne", 2, metric.GAUGE))

	e3 := i.LocalEntity()
	assert.NoError(t, e3.SetInventoryItem("inv", "key", "val"))

	assert.NoError(t, i.Publish())
}

func stripBlanks(b []byte) string {
	return strings.Replace(
		strings.Replace(
			strings.Replace(
				string(b),
				" ", "", -1),
			"\n", "", -1),
		"\t", "", -1)
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
