package args_test

import (
	"flag"
	"os"
	"reflect"
	"testing"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/stretchr/testify/assert"
)

func TestSetupArgsDefault(t *testing.T) {
	type argumentList struct {
		Verbose  bool          `default:"false" help:"Print more information to logs."`
		Pretty   bool          `default:"false" help:"Print pretty formatted JSON."`
		Hostname string        `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int           `default:"3306" help:"Port on which MySQL server is listening."`
		Username string        `help:"Username for accessing the database."`
		Password string        `help:"Passowrd for the given user."`
		Config   sdk_args.JSON `default:"randomstring" help:""`
	}
	var args argumentList

	_ = os.Setenv("HOSTNAME", "")
	defer os.Clearenv()
	os.Args = []string{"cmd"}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	expected := argumentList{
		Verbose: false, Pretty: false, Hostname: "localhost", Port: 3306,
		Username: "", Password: "",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected args list: %+v not equal to result: %+v", expected, args)
	}
}

func TestSetupArgsCommandLine(t *testing.T) {
	type argumentList struct {
		Verbose  bool          `default:"false" help:"Print more information to logs."`
		Pretty   bool          `default:"false" help:"Print pretty formatted JSON."`
		Hostname string        `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int           `default:"3306" help:"Port on which MySQL server is listening."`
		Username string        `help:"Username for accessing the database."`
		Password string        `help:"Passowrd for the given user."`
		Config   sdk_args.JSON `default:"randomstring" help:""`
		List     sdk_args.JSON `default:"randomstring" help:""`
	}
	var args argumentList

	_ = os.Setenv("HOSTNAME", "")
	defer os.Clearenv()
	os.Args = []string{
		"cmd",
		"-verbose",
		"-pretty",
		"-hostname=otherhost",
		"-port=1234",
		"-password=dbpwd",
		"-username=dbuser",
		"-config={\"arg1\": 2}",
		"-list=[\"arg1\", 2]",
	}

	expected := argumentList{
		Verbose:  true,
		Pretty:   true,
		Hostname: "otherhost",
		Port:     1234,
		Username: "dbuser",
		Password: "dbpwd",
		Config:   *sdk_args.NewJSON(map[string]interface{}{"arg1": 2.0}),
		List:     *sdk_args.NewJSON([]interface{}{"arg1", 2.0}),
	}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assertEqualArgs(t, expected, args)
}

func TestSetupArgsEnvironment(t *testing.T) {
	type argumentList struct {
		Verbose  bool          `default:"false" help:"Print more information to logs."`
		Pretty   bool          `default:"false" help:"Print pretty formatted JSON."`
		Hostname string        `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int           `default:"3306" help:"Port on which MySQL server is listening."`
		Username string        `help:"Username for accessing the database."`
		Password string        `help:"Passowrd for the given user."`
		Config   sdk_args.JSON `default:"randomstring" help:""`
	}
	var args argumentList

	_ = os.Setenv("USERNAME", "")
	_ = os.Setenv("VERBOSE", "true")
	_ = os.Setenv("HOSTNAME", "otherhost")
	_ = os.Setenv("PORT", "1234")
	_ = os.Setenv("CONFIG", "{\"arg1\": 2}")
	defer os.Clearenv()
	os.Args = []string{"cmd"}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	cfg := sdk_args.NewJSON(map[string]interface{}{"arg1": 2.0})
	expected := argumentList{
		Verbose: true, Pretty: false, Hostname: "otherhost", Port: 1234,
		Username: "", Password: "", Config: *cfg,
	}

	assertEqualArgs(t, expected, args)
}

func TestCliArgsOverrideEnvironmentArgs(t *testing.T) {
	var args struct {
		Verbose bool `default:"false" help:"Print more information to logs."`
	}

	_ = os.Setenv("VERBOSE", "false")
	defer os.Clearenv()
	os.Args = []string{
		"cmd",
		"-verbose",
	}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.True(t, args.Verbose)
}

func TestMetadataFlag(t *testing.T) {
	os.Args = []string{
		"cmd",
		"-metadata",
	}

	clearFlagSet()
	var args sdk_args.DefaultArgumentList
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.True(t, args.Metadata)
}

func TestClusterFlagViaCli(t *testing.T) {
	os.Args = []string{
		"cmd",
		"-nri_cluster=foo",
	}

	clearFlagSet()
	var args sdk_args.DefaultArgumentList
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.Equal(t, "foo", args.NriCluster)
}

func TestClusterFlagViaEnvVar(t *testing.T) {
	var args sdk_args.DefaultArgumentList

	_ = os.Setenv("NRI_CLUSTER", "bar")
	defer os.Clearenv()
	os.Args = []string{"cmd"}
	clearFlagSet()

	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.Equal(t, "bar", args.NriCluster)
}

func TestServiceFlagViaCli(t *testing.T) {
	os.Args = []string{
		"cmd",
		"-nri_service=foo",
	}

	clearFlagSet()
	var args sdk_args.DefaultArgumentList
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.Equal(t, "foo", args.NriService)
}

func TestServiceFlagViaEnvVar(t *testing.T) {
	var args sdk_args.DefaultArgumentList

	_ = os.Setenv("NRI_SERVICE", "bar")
	defer os.Clearenv()
	os.Args = []string{"cmd"}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.Equal(t, "bar", args.NriService)
}

func TestSetupArgsErrors(t *testing.T) {
	type argumentList struct {
		Verbose bool `default:"badbool" help:"Print more information to logs."`
	}

	os.Args = []string{"cmd"}

	clearFlagSet()
	err := sdk_args.SetupArgs(&argumentList{})
	if err == nil {
		t.Error()
	}

	type argumentList2 struct {
		Verbose int `default:"badint" help:"Print more information to logs."`
	}

	clearFlagSet()
	assert.Error(t, sdk_args.SetupArgs(&argumentList2{}))

	type argumentList3 struct {
		Verbose float64 `default:"0.12" help:"Print more information to logs."`
	}

	clearFlagSet()
	assert.Error(t, sdk_args.SetupArgs(&argumentList3{}))
}

func TestSetupArgsParseJsonError(t *testing.T) {
	type argumentList4 struct {
		Config sdk_args.JSON `default:"randomstring" help:""`
	}

	os.Args = []string{
		"cmd",
		"-config={\"arg1\": 2",
	}

	clearFlagSet()
	err := sdk_args.SetupArgs(&argumentList4{})
	if err == nil {
		t.Error()
	}
}

func TestDefaultArgumentsWithPretty(t *testing.T) {
	clearFlagSet()
	os.Args = []string{
		"cmd",
		"-pretty",
	}

	var args sdk_args.DefaultArgumentList
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assertEqualArgs(t, sdk_args.DefaultArgumentList{Pretty: true}, args)
}

func TestAddCustomArgumentsToDefault(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
	}

	var args argumentList
	clearFlagSet()
	os.Args = []string{
		"cmd",
		"-pretty",
		"-hostname=otherhost",
	}
	assert.NoError(t, sdk_args.SetupArgs(&args))

	expected := argumentList{
		DefaultArgumentList: sdk_args.DefaultArgumentList{Pretty: true},
		Hostname:            "otherhost",
	}

	assertEqualArgs(t, expected, args)
}

func TestGetDefaultArgs(t *testing.T) {
	type argumentListWithDefaults struct {
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
		sdk_args.DefaultArgumentList
	}
	da := sdk_args.GetDefaultArgs(&argumentListWithDefaults{})

	assertEqualArgs(t, sdk_args.DefaultArgumentList{}, *da)

	al := &argumentListWithDefaults{}
	al.Metrics = true
	dam := sdk_args.GetDefaultArgs(al)

	assertEqualArgs(t, sdk_args.DefaultArgumentList{Metrics: true}, *dam)
}

func TestGetDefaultArgsWithoutDefaults(t *testing.T) {
	type argumentListWithoutDefaults struct {
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
	}

	assertEqualArgs(t, sdk_args.DefaultArgumentList{}, *sdk_args.GetDefaultArgs(&argumentListWithoutDefaults{}))
}

// Vars to test getters
var (
	defaultArgs              = sdk_args.DefaultArgumentList{}
	defaultArgsWithMetrics   = sdk_args.DefaultArgumentList{Metrics: true}
	defaultArgsWithInventory = sdk_args.DefaultArgumentList{Inventory: true}
	defaultArgsWithEvents    = sdk_args.DefaultArgumentList{Events: true}
)

func TestDefaultArgumentList_All(t *testing.T) {
	assert.True(t, defaultArgs.All())
	assert.False(t, defaultArgsWithMetrics.All())
	assert.False(t, defaultArgsWithInventory.All())
	assert.False(t, defaultArgsWithEvents.All())
}

func TestDefaultArgumentList_HasMetrics(t *testing.T) {
	assert.True(t, defaultArgs.HasMetrics())
	assert.True(t, defaultArgsWithMetrics.HasMetrics())
	assert.False(t, defaultArgsWithInventory.HasMetrics())
	assert.False(t, defaultArgsWithEvents.HasMetrics())
}

func TestDefaultArgumentList_HasEvents(t *testing.T) {
	assert.True(t, defaultArgs.HasEvents())
	assert.False(t, defaultArgsWithMetrics.HasEvents())
	assert.False(t, defaultArgsWithInventory.HasEvents())
	assert.True(t, defaultArgsWithEvents.HasEvents())
}

func TestDefaultArgumentList_HasInventory(t *testing.T) {
	assert.True(t, defaultArgs.HasInventory())
	assert.False(t, defaultArgsWithMetrics.HasInventory())
	assert.True(t, defaultArgsWithInventory.HasInventory())
	assert.False(t, defaultArgsWithEvents.HasInventory())
}

func assertEqualArgs(t *testing.T, expected interface{}, args interface{}) {
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected args list:\n\t%+v\n not equal to result:\n\t%+v", expected, args)
	}
}

func clearFlagSet() {
	flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)
}
