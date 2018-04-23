package sdk_test

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

func TestDefaultValues(t *testing.T) {
	// Redirecting standard output to a file
	f, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)
	back := os.Stdout
	defer func() {
		os.Stdout = back
	}()
	os.Stdout = f

	// Given an integration builder without parameters
	integration, err := sdk.NewIntegration("integration", "4.0").Build()

	// The Build method does not return any error
	assert.NoError(t, err)

	// And the data is correctly set (including defaults)
	assert.Equal(t, "integration", integration.Name)
	assert.Equal(t, "4.0", integration.IntegrationVersion)
	assert.Equal(t, "2", integration.ProtocolVersion)
	assert.Equal(t, 0, len(integration.Data))

	// And when publishing the payload
	assert.NoError(t, integration.Publish())
	f.Close()

	// It is redirected to standard output, and non-prettified
	payload, err := ioutil.ReadFile(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "{\"name\":\"integration\",\"protocol_version\":\"2\",\"integration_version\":\"4.0\",\"data\":[]}",
		string(payload))
}

func TestIntegrationBuilder(t *testing.T) {
	// Redirecting standard output to a file
	output, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)

	// Needed for initialising os.Args + flags (emulating).
	os.Args = []string{"cmd", "--pretty"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	// Given an integration builder with all the parameters set
	var arguments args.DefaultArgumentList
	integration, err := sdk.NewIntegration("integration", "7.0").
		ParsedArguments(&arguments).
		Writer(output).
		Build()

	// The Build method does not return any error
	assert.NoError(t, err)

	// And the data is correctly set
	assert.Equal(t, "integration", integration.Name)
	assert.Equal(t, "7.0", integration.IntegrationVersion)
	assert.Equal(t, "2", integration.ProtocolVersion)
	assert.Equal(t, 0, len(integration.Data))

	// And when publishing the payload
	integration.Publish()
	output.Close()

	// The output works as specified
	payload, err := ioutil.ReadFile(output.Name())
	assert.NoError(t, err)
	assert.NotEqual(t, 0, len(payload))

	// And also is prettified
	assert.True(t, arguments.Pretty)
	assert.Contains(t, payload, uint8('\n'))
}

func TestWrongArguments(t *testing.T) {
	var d interface{} = struct{}{}

	arguments := []struct {
		arg interface{}
	}{
		{struct{ thing string }{"abcd"}},
		{1234},
		{"hello"},
		{[]struct{ x string }{{"hello"}, {"my friend"}}},
		{d},
	}
	for _, arg := range arguments {
		_, err := sdk.NewIntegration("integration", "7.0").ParsedArguments(arg).Build()
		assert.Error(t, err)
	}
}

func TestDefaultStorer(t *testing.T) {
	// Redirecting standard output to a file
	output, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)

	// Given an integration with the default cache
	integration, err := sdk.NewIntegration("cool-integration", "1.0").Writer(output).Build()
	assert.NoError(t, err)

	// And some values
	integration.Storer.Set("hello", 12.33)

	// When publishing the data
	assert.NoError(t, integration.Publish())

	// The data has been cached
	c, err := persist.NewStorer(persist.DefaultPath("cool-integration"), log.NewStdErr(false))
	assert.NoError(t, err)

	v, ts, ok := c.Get("hello")
	assert.True(t, ok)
	assert.NotEqual(t, 0, ts)
	assert.InDelta(t, 12.33, v, 0.1)
}

func TestNoStorer(t *testing.T) {
	// Redirecting standard output to a file
	output, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)

	// Given an integration with the no cache
	integration, err := sdk.NewIntegration("cool-integration", "1.0").Writer(output).NoStorer().Build()
	assert.NoError(t, err)

	// The built integration cache is nil
	assert.Nil(t, integration.Storer)

	// And the data can be published anyway
	assert.NoError(t, integration.Publish())
}

func TestCustomStorer(t *testing.T) {

	// Redirecting standard output to a file
	output, err := ioutil.TempFile("", "stdout")
	assert.NoError(t, err)

	// Given an integration with a custom cache
	customStorer := fakeStorer{}
	integration, err := sdk.NewIntegration("cool-integration", "1.0").Writer(output).Storer(&customStorer).Build()
	assert.NoError(t, err)

	// When publishing the data
	assert.NoError(t, integration.Publish())

	// The data has been cached
	assert.True(t, customStorer.saved)
}

type fakeStorer struct {
	saved bool
}

func (m *fakeStorer) Save() error {
	m.saved = true
	return nil
}

func (fakeStorer) Get(name string) (float64, int64, bool) {
	return 0, 0, false
}

func (fakeStorer) Set(name string, value float64) int64 {
	return 0
}
