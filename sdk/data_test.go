package sdk_test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/newrelic/infra-integrations-sdk/sdk"
)

func TestNewIntegrationData(t *testing.T) {
	pd, err := sdk.NewIntegration("TestPlugin", "1.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	if pd.Name != "TestPlugin" {
		t.Error()
	}
	if pd.IntegrationVersion != "1.0" {
		t.Error()
	}
	if pd.ProtocolVersion != "1" {
		t.Error()
	}
	if len(pd.Inventory) != 0 {
		t.Error()
	}
	if len(pd.Metrics) != 0 {
		t.Error()
	}
	if len(pd.Events) != 0 {
		t.Error()
	}
}

func TestNewMetricSet(t *testing.T) {
	pd, err := sdk.NewIntegration("TestPlugin", "1.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	metric1 := pd.NewMetricSet("TestPlugin", "TestProvider")
	if metric1 != pd.Metrics[0] {
		t.Error()
	}

	metric2 := pd.NewMetricSet("TestMetric2", "TestProvider2")
	if metric2 != pd.Metrics[1] {
		t.Error()
	}
}

func TestNewIntegrationWithDefaultArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	var al argumentList

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	pd, err := sdk.NewIntegration("TestPlugin", "1.0", &al)
	if err != nil {
		t.Fail()
	}
	if pd.Name != "TestPlugin" {
		t.Error()
	}
	if pd.IntegrationVersion != "1.0" {
		t.Error()
	}
	if pd.ProtocolVersion != "1" {
		t.Error()
	}
	if len(pd.Inventory) != 0 {
		t.Error()
	}
	if len(pd.Metrics) != 0 {
		t.Error()
	}
	if len(pd.Events) != 0 {
		t.Error()
	}
	if al.All != true {
		t.Error()
	}
	if al.Pretty != false {
		t.Error()
	}
	if al.Verbose != false {
		t.Error()
	}

}

func mockStdout(f func() error) (string, error) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := f(); err != nil {
		return "", err
	}

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	return out, nil
}

func TestPublish(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	var al argumentList
	flag.CommandLine = flag.NewFlagSet("name", 0)

	pd, err := sdk.NewIntegration("TestPlugin", "1.0", &al)
	if err != nil {
		t.Error()
	}

	jsonString, err := mockStdout(pd.Publish)
	if err != nil {
		t.Error()
	}

	lines := strings.Split(strings.TrimSpace(jsonString), "\n")
	if len(lines) != 1 {
		t.Error()
	}

	os.Args = []string{
		"cmd",
		"-pretty",
	}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	pd, err = sdk.NewIntegration("TestPlugin", "1.0", &al)
	if err != nil {
		t.Error()
	}

	jsonString, err = mockStdout(pd.Publish)
	if err != nil {
		t.Error()
	}

	lines = strings.Split(strings.TrimSpace(jsonString), "\n")
	fmt.Println(len(lines), al.Pretty)
	if len(lines) <= 1 {
		t.Error()
	}
}
