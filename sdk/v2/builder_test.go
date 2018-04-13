package v2_test

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/sdk/v2"
	"github.com/stretchr/testify/assert"
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
	integration, err := v2.NewIntegration("integration", "4.0").Build()

	// The Build method does not return any error
	assert.NoError(t, err)

	// And the data is correctly set (including defaults)
	assert.Equal(t, "integration", integration.Name)
	assert.Equal(t, "4.0", integration.IntegrationVersion)
	assert.Equal(t, "2", integration.ProtocolVersion)
	assert.Equal(t, 0, len(integration.Data))

	// And when publishing the payload
	integration.Publish()
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
	integration, err := v2.NewIntegration("integration", "7.0").
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
		_, err := v2.NewIntegration("integration", "7.0").ParsedArguments(arg).Build()
		assert.Error(t, err)
	}
}
