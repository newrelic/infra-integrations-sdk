package jmx_test

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/jmx"
)

func TestMain(m *testing.M) {
	var testType string
	flag.StringVar(&testType, "test.type", "", "")
	flag.Parse()

	if testType == "" {
		// Set the NR_JMX_TOOL to ourselves (the test binary) with the extra
		// parameter test.type=helper and run the tests as usual.
		os.Setenv("NR_JMX_TOOL", fmt.Sprintf("%s -test.type helper --", os.Args[0]))
		os.Exit(m.Run())
	} else if testType == "helper" {
		// The test suite becomes a JMX Tool
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			command := scanner.Text()
			if command == "empty" {
				fmt.Println("{}")
			} else if command == "crash" {
				os.Exit(1)
			} else if command == "invalid" {
				fmt.Println("not a json")
			} else if command == "timeout" {
				time.Sleep(1000 * time.Millisecond)
				fmt.Println("{}")
			}
		}
		os.Exit(0)
	}
}

func TestJmxOpen(t *testing.T) {
	defer jmx.Close()

	if jmx.Open("", "", "", "") != nil {
		t.Error()
	}

	if jmx.Open("", "", "", "") == nil {
		t.Error()
	}
}

func TestJmxQuery(t *testing.T) {
	defer jmx.Close()

	if jmx.Open("", "", "", "") != nil {
		t.Error()
	}

	if _, err := jmx.Query("empty"); err != nil {
		t.Error()
	}
}

func TestJmxCrashQuery(t *testing.T) {
	defer jmx.Close()

	if jmx.Open("", "", "", "") != nil {
		t.Error()
	}

	if _, err := jmx.Query("crash"); err == nil {
		t.Error()
	}
}

func TestJmxInvalidQuery(t *testing.T) {
	defer jmx.Close()

	if jmx.Open("", "", "", "") != nil {
		t.Error()
	}

	if _, err := jmx.Query("invalid"); err == nil {
		t.Error()
	}
}

func TestJmxTimeoutQuery(t *testing.T) {
	defer jmx.Close()

	if jmx.Open("", "", "", "") != nil {
		t.Error()
	}

	if _, err := jmx.Query("timeout"); err == nil {
		t.Error()
	}

	if _, err := jmx.Query("empty"); err == nil {
		t.Error()
	}
}
