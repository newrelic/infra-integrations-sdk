package jmx

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	timeout      = 1000
	openAttempts = 5
)

var query2IsErr = map[string]bool{
	"empty":   false,
	"crash":   true,
	"invalid": true,
}

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
			} else if command == "bigPayload" {
				// Create a payload of more than 64K
				fmt.Println(fmt.Sprintf("{\"first\": 1%s}", strings.Repeat(", \"s\": 2", 70*1024)))
			} else if command == "bigPayloadError" {
				// Create a payload of more than 4M
				fmt.Println(fmt.Sprintf("{\"first\": 1%s}", strings.Repeat(", \"s\": 2", 4*1024*1024)))
			}
		}
		os.Exit(0)
	}
}

func TestOpenWithParameters_OnlyWorksWhenClosed(t *testing.T) {
	defer Close()

	assert.NoError(t, Open("", "", "", ""))
	assert.Error(t, Open("", "", "", ""))
	Close()
	assert.NoError(t, Open("", "", "", ""))
}

func TestQuery(t *testing.T) {
	for q, isErr := range query2IsErr {
		assert.NoError(t, openWait("", "", "", "", openAttempts), "error on opening for query %s", q)

		_, err := Query(q, timeout)
		if isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		Close()
	}
}

func TestQuery_WithSSL(t *testing.T) {
	for q, isErr := range query2IsErr {
		assert.NoError(t, openWaitWithSSL("", "", "", "", "", "", "", "", openAttempts))

		_, err := Query(q, timeout)
		if isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		Close()
	}
}

func TestQuery_TimeoutReturnsError(t *testing.T) {
	defer Close()

	if err := openWait("", "", "", "", openAttempts); err != nil {
		t.Error(err)
	}

	if _, err := Query("timeout", timeout); err == nil {
		t.Error()
	}

	if _, err := Query("empty", timeout); err == nil {
		t.Error()
	}
}

func TestJmxNoTimeoutQuery(t *testing.T) {
	t.Skip("unreliable CI test")

	defer Close()

	if err := openWait("", "", "", "", openAttempts); err != nil {
		t.Error(err)
	}

	if _, err := Query("timeout", 1500); err != nil {
		t.Error(err)
	}
}

func TestJmxTimeoutBigQuery(t *testing.T) {
	t.Skip("unreliable CI test")

	defer Close()

	if err := openWait("", "", "", "", openAttempts); err != nil {
		t.Error(err)
	}

	if _, err := Query("bigPayload", timeout); err != nil {
		t.Error(err)
	}

	if _, err := Query("bigPayloadError", timeout); err == nil {
		t.Error()
	}
}

// tests can overlap, and as jmx-cmd is a singleton, waiting for it to be closed is mandatory
func openWait(hostname, port, username, password string, attempts int) error {
	return openWaitWithSSL(hostname, port, username, password, "", "", "", "", attempts)
}

func openWaitWithSSL(hostname, port, username, password, keyStore, keyStorePassword, trustStore, trustStorePassword string, attempts int) error {
	ssl := WithSSL(keyStore, keyStorePassword, trustStore, trustStorePassword)
	err := Open(hostname, port, username, password, ssl)
	if err == ErrJmxCmdRunning && attempts > 0 {
		attempts--
		time.Sleep(10 * time.Millisecond)

		return openWaitWithSSL(hostname, port, username, password, keyStore, keyStorePassword, trustStore, trustStorePassword, attempts)
	}

	return err
}

// test that if we receive a WARNING message we still will receive the actual data.
func TestLoop(t *testing.T) {
	defer flushWarnings()
	_, cancelFn := context.WithCancel(context.Background())

	lineCh := make(chan []byte, jmxLineBuffer*2)
	queryErrors := make(chan error)
	outTimeout := time.Duration(timeout) * time.Millisecond
	receiveResult(lineCh, queryErrors, cancelFn, "empty", outTimeout)
	warningMessage := "WARNING foo bar"
	cmdErr <- fmt.Errorf(warningMessage)
	errorChannel := <-cmdErr
	assert.Equal(t, errorChannel, fmt.Errorf(warningMessage))
	b := []byte("{foo}")
	lineCh <- b
	msg := string(<-lineCh)
	assert.Equal(t, msg, "{foo}")

}
