package v1

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	sdk_args "github.com/newrelic/infra-integrations-sdk/args"
	"github.com/stretchr/testify/assert"
)

func TestNewIntegrationData(t *testing.T) {
	pd, err := NewIntegration("TestPlugin", "1.0", new(struct{}))
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
	pd, err := NewIntegration("TestPlugin", "1.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}

	metric1 := pd.NewMetricSet("TestPlugin")
	assert.Equal(t, metric1, pd.Metrics[0])

	metric2 := pd.NewMetricSet("TestMetric2")
	assert.Equal(t, metric2, pd.Metrics[1])
}

func TestNewIntegrationWithDefaultArguments(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	var al argumentList

	os.Args = []string{"cmd"}
	flag.CommandLine = flag.NewFlagSet("name", 0)

	pd, err := NewIntegration("TestPlugin", "1.0", &al)
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

func TestPublish(t *testing.T) {
	type argumentList struct {
		sdk_args.DefaultArgumentList
	}

	var al argumentList
	flag.CommandLine = flag.NewFlagSet("name", 0)

	pd, err := NewIntegration("TestPlugin", "1.0", &al)
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

	pd, err = NewIntegration("TestPlugin", "1.0", &al)
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

func TestSetInventoryItem(t *testing.T) {
	pd, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal()
	}
	pd.Inventory.SetItem("foo/bar", "valueOne", "bar")
	pd.Inventory.SetItem("foo/bar", "valueTwo", "bar")

	if len(pd.Inventory) != 1 {
		t.Error()
	}
	if len(pd.Inventory["foo/bar"]) != 2 {
		t.Error()
	}
	expectedValue := "bar"
	actualValue := pd.Inventory["foo/bar"]["valueOne"]
	if expectedValue != actualValue {
		t.Errorf("For '%s' inventory item, the expected field '%s' is '%s'. Actual value: '%s'", "foo/bar", "valueOne", expectedValue, actualValue)
	}
	actualValue = pd.Inventory["foo/bar"]["valueTwo"]
	if expectedValue != actualValue {
		t.Errorf("For '%s' inventory item, the expected field '%s' is '%s'. Actual value: '%s'", "foo/bar", "valueTwo", expectedValue, actualValue)
	}
}

func TestAddNotificationEvent_Integration(t *testing.T) {
	i, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal(err)
	}

	err = i.AddNotificationEvent("TestSummary")
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}

	if i.Events[0].Summary != "TestSummary" || i.Events[0].Category != "notifications" {
		t.Error("event malformed")
	}

	if len(i.Events) != 1 {
		t.Error("not expected length of events")
	}
}

func TestAddNotificationEvent_Integration_NoSummary_Error(t *testing.T) {
	i, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal(err)
	}

	err = i.AddNotificationEvent("")
	if err == nil {
		t.Error("error was expected for empty summary")
	}

	if len(i.Events) != 0 {
		t.Error("not expected length of events")
	}
}

func TestAddEvent_Integration(t *testing.T) {
	i, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal(err)
	}

	err = i.AddEvent(Event{Summary: "TestSummary", Category: "TestCategory"})
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}

	if i.Events[0].Summary != "TestSummary" || i.Events[0].Category != "TestCategory" {
		t.Error("event malformed")
	}

	if len(i.Events) != 1 {
		t.Error("not expected length of events")
	}
}

func TestAddEvent_Integration_TheSameEvents_And_NoCategory(t *testing.T) {
	i, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal(err)
	}

	err = i.AddEvent(Event{Summary: "TestSummary"})
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}
	err = i.AddEvent(Event{Summary: "TestSummary"})
	if err != nil {
		t.Errorf("error not expected, got: %s", err)
	}

	if i.Events[0].Summary != "TestSummary" || i.Events[0].Category != "" {
		t.Error("event malformed")
	}
	if i.Events[1].Summary != "TestSummary" || i.Events[1].Category != "" {
		t.Error("event malformed")
	}
	if len(i.Events) != 2 {
		t.Error("not expected length of events")
	}
}

func TestAddEvent_Integration_EmptySummary_Error(t *testing.T) {
	i, err := NewIntegration("TestIntegration", "1.0", new(struct{}))
	if err != nil {
		t.Fatal(err)
	}

	err = i.AddEvent(Event{Category: "TestCategory"})
	if err == nil {
		t.Error("error was expected for empty summary")
	}

	if len(i.Events) != 0 {
		t.Error("not expected length of events")
	}
}

// Helpers
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
