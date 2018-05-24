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

	os.Setenv("HOSTNAME", "")
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

	os.Setenv("HOSTNAME", "")
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

	os.Setenv("USERNAME", "")
	os.Setenv("VERBOSE", "true")
	os.Setenv("HOSTNAME", "otherhost")
	os.Setenv("PORT", "1234")
	os.Setenv("CONFIG", "{\"arg1\": 2}")
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

func TestEnvironmentVarsOverrideCliArgs(t *testing.T) {
	var args struct {
		Verbose bool `default:"false" help:"Print more information to logs."`
	}

	os.Setenv("VERBOSE", "false")
	os.Args = []string{
		"cmd",
		"-verbose",
	}

	clearFlagSet()
	assert.NoError(t, sdk_args.SetupArgs(&args))

	assert.False(t, args.Verbose)
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

func TestSetupArgsWithDefaultArguments(t *testing.T) {
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

	assertEqualArgs(t, sdk_args.DefaultArgumentList{All: true}, *da)

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

func assertEqualArgs(t *testing.T, expected interface{}, args interface{}) {
	if !reflect.DeepEqual(args, expected) {
		t.Errorf("expected args list:\n\t%+v\n not equal to result:\n\t%+v", expected, args)
	}
}

func clearFlagSet() {
	flag.CommandLine = flag.NewFlagSet("cmd", flag.ContinueOnError)
}
