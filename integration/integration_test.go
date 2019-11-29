package integration

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/newrelic/infra-integrations-sdk/data/attribute"
	"github.com/newrelic/infra-integrations-sdk/data/event"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
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

	assert.Equal(t, "TestIntegration", i.Name)
	assert.Equal(t, "1.0", i.IntegrationVersion)
	assert.Equal(t, "3", i.ProtocolVersion)
	assert.Len(t, i.Entities, 0)
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

func TestIntegration_LocalEntity(t *testing.T) {
	i := newTestIntegration(t)

	e1 := i.LocalEntity()
	e2 := i.LocalEntity()
	assert.Equal(t, e1, e2)
}

func TestIntegration_Entity(t *testing.T) {
	i := newTestIntegration(t)

	e1, err := i.Entity("name", "ns")
	assert.NoError(t, err)
	e2, err := i.Entity("name2", "ns")
	assert.NoError(t, err)
	e3, err := i.Entity("name", "ns")
	assert.NoError(t, err)

	assert.NotEqual(t, e1, e2, "Different names create different entities")
	assert.Equal(t, e1, e3, "Same namespace & name create/retrieve same entity")
}

func TestIntegration_Entity_WithIDAttrs(t *testing.T) {
	i := newTestIntegration(t)

	idAttr := NewIDAttribute("k", "v")

	e1, err := i.Entity("name", "ns", idAttr)
	assert.NoError(t, err)
	e2, err := i.Entity("name", "ns")
	assert.NoError(t, err)
	e3, err := i.Entity("name", "ns", idAttr)
	assert.NoError(t, err)

	assert.NotEqual(t, e1, e2, "Different id-attributes create different entities")
	assert.Equal(t, e1, e3, "Same namespace, name and id-attributes create/retrieve same entity")
}

func TestIntegration_EntityReportedBy(t *testing.T) {
	i := newTestIntegration(t)

	e, err := i.EntityReportedBy("reporting:entity:key", "name", "ns")
	assert.NoError(t, err)

	reportingEntityExists := false
	for _, a := range e.customAttributes {
		if a.Key == AttrReportingEntity {
			reportingEntityExists = true
			assert.Equal(t, a.Value, "reporting:entity:key")
		}
	}
	require.True(t, reportingEntityExists)
}

func TestIntegration_EntityReportedVia(t *testing.T) {
	i := newTestIntegration(t)

	e, err := i.EntityReportedVia("reporting.endpoint:123", "name", "ns")
	assert.NoError(t, err)

	reportingEndpointExists := false
	for _, a := range e.customAttributes {
		if a.Key == AttrReportingEndpoint {
			reportingEndpointExists = true
			assert.Equal(t, a.Value, "reporting.endpoint:123")
		}
	}
	require.True(t, reportingEndpointExists)
}

func TestDefaultArguments(t *testing.T) {
	al := sdk_args.DefaultArgumentList{}

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	assert.Equal(t, "TestIntegration", i.Name)
	assert.Equal(t, "1.0", i.IntegrationVersion)
	assert.Equal(t, "3", i.ProtocolVersion)
	assert.Len(t, i.Entities, 0)
	assert.True(t, al.All())
	assert.False(t, al.Pretty)
	assert.False(t, al.Verbose)
}

func TestClusterAndServiceArgumentsAreAddedToMetadata(t *testing.T) {
	al := sdk_args.DefaultArgumentList{}

	os.Setenv("NRI_CLUSTER", "foo")
	os.Setenv("NRI_SERVICE", "bar")
	defer os.Clearenv()
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)

	i, err := New("TestIntegration", "1.0", Logger(log.Discard), Writer(ioutil.Discard), Args(&al))
	assert.NoError(t, err)

	e, err := i.Entity("name", "ns")
	assert.NoError(t, err)

	assert.Equal(t, []attribute.Attribute{
		{
			Key:   CustomAttrCluster,
			Value: "foo",
		},
		{
			Key:   CustomAttrService,
			Value: "bar",
		},
	}, e.customAttributes)
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
					"type": "test",
					"id_attributes": [
					  {
						"Key":"env",
						"Value":"prod"
					  }
					]
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
					  "category": "evnt1cat",
						"attributes": {
							"attr1": "attr1Val",
							"attr2": 42
						}
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
					"type": "test",
					"id_attributes": []
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

	e, err := i.Entity("EntityOne", "test", NewIDAttribute("env", "prod"))
	assert.NoError(t, err)
	ms := e.NewMetricSet("EventTypeForEntityOne")
	assert.NoError(t, ms.SetMetric("metricOne", 1, metric.GAUGE))
	assert.NoError(t, ms.SetMetric("metricTwo", "test", metric.ATTRIBUTE))
	assert.NoError(t, ms.SetMetric("metricBool", true, metric.GAUGE))

	assert.NoError(t, e.AddEvent(event.NewWithAttributes(
		"evnt1sum",
		"evnt1cat",
		map[string]interface{}{
			"attr1": "attr1Val",
			"attr2": 42,
		},
	)))
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
