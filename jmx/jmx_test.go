package jmx

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	timeoutMillis = 1500
	openAttempts  = 10
	// jmx mock cmds
	cmdEmpty         = "empty"
	cmdCrash         = "crash"
	cmdInvalid       = "invalid"
	cmdTimeout       = "timeout"
	cmdBigPayload    = "bigPayload"
	cmdBigPayloadErr = "bigPayloadError"
)

var query2IsErr = map[string]bool{
	cmdEmpty:   false,
	cmdCrash:   true,
	cmdInvalid: true,
	//cmdTimeout: true, // flaky test
}

func TestMain(m *testing.M) {
	var testType string
	flag.StringVar(&testType, "test.type", "", "")
	flag.Parse()

	if testType == "" {
		// Set the NR_JMX_TOOL to ourselves (the test binary) with the extra
		// parameter test.type=helper and run the tests as usual.
		_ = os.Setenv("NR_JMX_TOOL", fmt.Sprintf("%s -test.type helper --", os.Args[0]))
		os.Exit(m.Run())
	} else if testType == "helper" {
		// The test suite becomes a JMX Tool
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			command := scanner.Text()
			if command == cmdEmpty {
				fmt.Println("{}")
			} else if command == cmdCrash {
				os.Exit(1)
			} else if command == cmdInvalid {
				fmt.Println("not a json")
			} else if command == cmdTimeout {
				time.Sleep(timeoutMillis + 200*time.Millisecond)
				fmt.Println("{}")
			} else if command == cmdBigPayload {
				// Create a payload of more than 64K
				str := fmt.Sprintf("{\"first\": 1%s}", strings.Repeat(", \"s\": 2", 70*1024))
				fmt.Println(str)
			} else if command == cmdBigPayloadErr {
				// Create a payload of more than 4M
				str := fmt.Sprintf("{\"first\": 1%s}", strings.Repeat(", \"s\": 2", 4*1024*1024))
				fmt.Println(str)
			}

		}
		os.Exit(0)
	}
}

func TestOpenWithParameters_OnlyWorksWhenClosed(t *testing.T) {
	defer Close()

	assert.NoError(t, OpenNoAuth("", ""))
	assert.Error(t, OpenNoAuth("", ""))
	Close()
	assert.NoError(t, OpenNoAuth("", ""))
}

func TestOpenURL(t *testing.T) {
	defer Close()

	assert.NoError(t, OpenURL("sample.url", "", ""))
	lastArg := cmd.Args[len(cmd.Args)-1]
	assert.Equal(t, "sample.url", lastArg)
}

// nolint: goconst
func TestQuery(t *testing.T) {
	for q, isErr := range query2IsErr {
		require.NoError(t, openWait("", "", "", "", openAttempts), "error on opening for query %s", q)

		_, err := Query(q, timeoutMillis)
		if isErr {
			assert.Error(t, err, "case "+q)
		} else {
			assert.NoError(t, err, "case "+q)
		}
		Close()
	}
}

func TestQuery_WithSSL(t *testing.T) {
	for q, isErr := range query2IsErr {
		require.NoError(t, openWaitWithSSL("", "", "", "", "", "", "", "", openAttempts))

		_, err := Query(q, timeoutMillis)

		if isErr {
			assert.Error(t, err, "case "+q)
		} else {
			assert.NoError(t, err, "case "+q)
		}
		Close()
	}
}

func TestOpen_WithNrjmx(t *testing.T) {
	aux := os.Getenv("NR_JMX_TOOL")
	require.NoError(t, os.Unsetenv("NR_JMX_TOOL"))

	assert.Error(t, OpenNoAuth("", "", WithNrJmxTool("/foo")), "/foo is not an executable")
	assert.Equal(t, "/foo", cmd.Args[0])

	require.NoError(t, os.Setenv("NR_JMX_TOOL", aux))
}

func TestJmxNoTimeoutQuery(t *testing.T) {
	t.Skip("unreliable CI test")

	defer Close()

	if err := openWait("", "", "", "", openAttempts); err != nil {
		t.Error(err)
	}

	if _, err := Query(cmdTimeout, timeoutMillis+1000); err != nil {
		t.Error(err)
	}
}

func TestJmxTimeoutBigQuery(t *testing.T) {
	t.Skip("unreliable CI test")

	defer Close()

	if err := openWait("", "", "", "", openAttempts); err != nil {
		t.Error(err)
	}

	if _, err := Query(cmdBigPayload, timeoutMillis); err != nil {
		t.Error(err)
	}

	if _, err := Query(cmdBigPayloadErr, timeoutMillis); err == nil {
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
		time.Sleep(100 * time.Millisecond)

		return openWaitWithSSL(hostname, port, username, password, keyStore, keyStorePassword, trustStore, trustStorePassword, attempts)
	}

	return err
}

func Test_receiveResult_warningsDoNotBreakResultReception(t *testing.T) {
	// Make sure there is no unread errors or warnings.
	// TODO: This should perhaps be done in Close()?
	cmdErrC = make(chan error, cmdStdChanLen)
	cmdWarnC = make(chan string, cmdStdChanLen)

	var buf bytes.Buffer
	log.SetOutput(&buf)

	_, cancelFn := context.WithCancel(context.Background())

	resultCh := make(chan []byte, 1)
	queryErrCh := make(chan error)
	outTimeout := time.Duration(timeoutMillis) * time.Millisecond
	warningMessage := fmt.Sprint("WARNING foo bar")
	cmdWarnC <- warningMessage

	resultCh <- []byte("{\"foo\":1}")

	result, err := receiveResult(resultCh, queryErrCh, cancelFn, "foo", outTimeout)

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"foo": 1.,
	}, result)
	assert.Equal(t, fmt.Sprintf("[WARN] %s\n", warningMessage), buf.String())
}

func Test_receiveResult_invalidJsonIsPrintedInError(t *testing.T) {

	var buf bytes.Buffer
	log.SetOutput(&buf)

	_, cancelFn := context.WithCancel(context.Background())

	resultCh := make(chan []byte, 2)
	queryErrCh := make(chan error)
	outTimeout := time.Duration(timeoutMillis) * time.Millisecond

	resultCh <- []byte("#this is an invalid json")

	result, err := receiveResult(resultCh, queryErrCh, cancelFn, "foo", outTimeout)

	assert.Equal(t, "invalid return value for query: foo, error: invalid character '#' looking for beginning of value, line: \"#this is an invalid json\"", err.Error())
	assert.Nil(t, result)
}

func Test_DefaultPath_IsCorrectForOs(t *testing.T) {
	os := runtime.GOOS
	switch os {
	case "windows":
	case "darwin":
	case "linux":
		assert.True(t, len(defaultNrjmxExec) > 0)
	default:
		t.Fatal("unexpected value")
	}
}
