package args_test

import (
	"flag"
	"os"
	"reflect"
	"testing"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
)

func TestSetupArgsDefault(t *testing.T) {
	type argumentList struct {
		Verbose  bool   `default:"false" help:"Print more information to logs."`
		Pretty   bool   `default:"false" help:"Print pretty formatted JSON."`
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int    `default:"3306" help:"Port on which MySQL server is listening."`
		Username string `help:"Username for accessing the database."`
		Password string `help:"Passowrd for the given user."`
	}
	var args argumentList

	os.Setenv("USERNAME", "")
	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	sdk_args.SetupArgs(&args)

	expected := argumentList{Verbose: false, Pretty: false, Hostname: "localhost", Port: 3306, Username: "", Password: ""}
	if !reflect.DeepEqual(args, expected) {
		t.Error()
	}
}

func TestSetupArgsCommandLine(t *testing.T) {
	type argumentList struct {
		Verbose  bool   `default:"false" help:"Print more information to logs."`
		Pretty   bool   `default:"false" help:"Print pretty formatted JSON."`
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int    `default:"3306" help:"Port on which MySQL server is listening."`
		Username string `help:"Username for accessing the database."`
		Password string `help:"Passowrd for the given user."`
	}
	var args argumentList

	os.Setenv("USERNAME", "")
	os.Args = []string{
		"cmd",
		"-verbose",
		"-pretty",
		"-hostname=otherhost",
		"-port=1234",
		"-password=dbpwd",
		"-username=dbuser",
	}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	sdk_args.SetupArgs(&args)

	expected := argumentList{Verbose: true, Pretty: true, Hostname: "otherhost", Port: 1234, Username: "dbuser", Password: "dbpwd"}
	if !reflect.DeepEqual(args, expected) {
		t.Error()
	}
}

func TestSetupArgsEnvironment(t *testing.T) {
	type argumentList struct {
		Verbose  bool   `default:"false" help:"Print more information to logs."`
		Pretty   bool   `default:"false" help:"Print pretty formatted JSON."`
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
		Port     int    `default:"3306" help:"Port on which MySQL server is listening."`
		Username string `help:"Username for accessing the database."`
		Password string `help:"Passowrd for the given user."`
	}
	var args argumentList

	os.Setenv("USERNAME", "")
	os.Setenv("VERBOSE", "true")
	os.Setenv("HOSTNAME", "otherhost")
	os.Setenv("PORT", "1234")
	os.Args = []string{"cmd"}

	flag.CommandLine = flag.NewFlagSet("name", 0)

	sdk_args.SetupArgs(&args)

	expected := argumentList{Verbose: true, Pretty: false, Hostname: "otherhost", Port: 1234, Username: "", Password: ""}
	if !reflect.DeepEqual(args, expected) {
		t.Error()
	}
}

func TestSetupArgsErrors(t *testing.T) {
	type argumentList struct {
		Verbose bool `default:"badbool" help:"Print more information to logs."`
	}

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	err := sdk_args.SetupArgs(&argumentList{})
	if err == nil {
		t.Error()
	}

	type argumentList2 struct {
		Verbose int `default:"badint" help:"Print more information to logs."`
	}

	flag.CommandLine = flag.NewFlagSet("name", 0)

	err = sdk_args.SetupArgs(&argumentList2{})
	if err == nil {
		t.Error()
	}

	type argumentList3 struct {
		Verbose float64 `default:"0.12" help:"Print more information to logs."`
	}

	flag.CommandLine = flag.NewFlagSet("name", 0)

	err = sdk_args.SetupArgs(&argumentList3{})
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
	os.Args = []string{
		"cmd",
		"-verbose",
		"-pretty",
		"-hostname=otherhost",
	}
	sdk_args.SetupArgs(&args)

	expected := argumentList{
		DefaultArgumentList: sdk_args.DefaultArgumentList{Verbose: true, Pretty: true},
		Hostname:            "otherhost",
	}
	if !reflect.DeepEqual(args, expected) {
		t.Error()
	}
}

func TestGetDefaultArgs(t *testing.T) {
	type argumentListWithDefaults struct {
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
		sdk_args.DefaultArgumentList
	}
	da := sdk_args.GetDefaultArgs(&argumentListWithDefaults{})

	expected := sdk_args.DefaultArgumentList{All: true}
	if !reflect.DeepEqual(*da, expected) {
		t.Error()
	}

	al := &argumentListWithDefaults{}
	al.Metrics = true
	dam := sdk_args.GetDefaultArgs(al)

	expected = sdk_args.DefaultArgumentList{Metrics: true}
	if !reflect.DeepEqual(*dam, expected) {
		t.Error()
	}
}

func TestGetDefaultArgsWithoutDefaults(t *testing.T) {
	type argumentListWithoutDefaults struct {
		Hostname string `default:"localhost" help:"Hostname or IP where MySQL is running."`
	}
	da := sdk_args.GetDefaultArgs(&argumentListWithoutDefaults{})
	expected := sdk_args.DefaultArgumentList{}
	if !reflect.DeepEqual(*da, expected) {
		t.Error()
	}
}
