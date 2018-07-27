/*
Package jmx is a library to get metrics through JMX. It requires additional
setup. Read https://github.com/newrelic/infra-integrations-sdk#jmx-support for
instructions. */
package jmx

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var lock sync.Mutex
var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdError io.ReadCloser
var cmdIn io.WriteCloser
var cmdErr = make(chan error, 1)
var done sync.WaitGroup
var warnings []string

var (
	jmxCommand = "/usr/bin/nrjmx"
	// ErrJmxCmdRunning error returned when trying to Open and nrjmx command is still running
	ErrJmxCmdRunning = errors.New("JMX tool is already running")
)

const (
	jmxLineBuffer = 4 * 1024 * 1024 // Max 4MB per line. If single lines are outputting more JSON than that, we likely need smaller-scoped JMX queries
)

func getCommand(hostname, port, username, password string) []string {
	var cliCommand []string

	if os.Getenv("NR_JMX_TOOL") != "" {
		cliCommand = strings.Split(os.Getenv("NR_JMX_TOOL"), " ")
	} else {
		cliCommand = []string{jmxCommand}
	}

	cliCommand = append(cliCommand, "--hostname", hostname, "--port", port)
	if username != "" && password != "" {
		cliCommand = append(cliCommand, "--username", username, "--password", password)
	}

	return cliCommand
}

// Open will start the nrjmx command with the provided connection parameters.
func Open(hostname, port, username, password string) error {
	lock.Lock()
	defer lock.Unlock()

	if cmd != nil {
		return ErrJmxCmdRunning
	}

	// Drain error channel to prevent showing past errors
	if len(cmdErr) > 0 {
		<-cmdErr
	}

	done.Add(1)

	var err error
	var ctx context.Context

	cliCommand := getCommand(hostname, port, username, password)

	ctx, cancel = context.WithCancel(context.Background())
	cmd = exec.CommandContext(ctx, cliCommand[0], cliCommand[1:]...)

	if cmdOut, err = cmd.StdoutPipe(); err != nil {
		return err
	}
	if cmdIn, err = cmd.StdinPipe(); err != nil {
		return err
	}

	if cmdError, err = cmd.StderrPipe(); err != nil {
		return err
	}

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err = cmd.Wait(); err != nil {
			stdErr, _ := ioutil.ReadAll(cmdError)
			cmdErr <- fmt.Errorf("JMX tool exited with error: %s [state: %s] (%s)", err, cmd.ProcessState, string(stdErr))
		}
		done.Done()

		lock.Lock()
		defer lock.Unlock()
		cmd = nil
	}()

	return nil
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	lock.Lock()
	defer lock.Unlock()

	if cmd == nil {
		return
	}

	cancel()

	done.Wait()
	closeCmdIO()
}

func closeCmdIO() {
	_ = cmdIn.Close()
	_ = cmdError.Close()
}

func doQuery(ctx context.Context, out chan []byte, errorChan chan error, queryString []byte) {
	lock.Lock()
	if _, err := cmdIn.Write(queryString); err != nil {
		lock.Unlock()
		errorChan <- fmt.Errorf("writing query string: %s", err.Error())
		return
	}

	scanner := bufio.NewScanner(cmdOut)
	scanner.Buffer([]byte{}, jmxLineBuffer) // Override default buffer to increase buffer size
	lock.Unlock()

	if scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		case out <- scanner.Bytes():
		default:
		}
	} else {
		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading output from JMX tool: %v", err)
		} else {
			// If scanner.Scan() returns false but err is also nil, it hit EOF. We consider that a problem, so we should return an error.
			errorChan <- fmt.Errorf("got an EOF while reading JMX tool output")
		}
	}
}

// Query executes JMX query against nrjmx tool waiting up to timeout (in milliseconds)
func Query(objectPattern string, timeout int) (map[string]interface{}, error) {
	defer flushWarnings()
	ctx, cancelFn := context.WithCancel(context.Background())

	lineCh := make(chan []byte)
	queryErrors := make(chan error)
	outTimeout := time.Duration(timeout) * time.Millisecond
	// Send the query async to the underlying process so we can timeout it
	go doQuery(ctx, lineCh, queryErrors, []byte(fmt.Sprintf("%s\n", objectPattern)))

	return receiveResult(lineCh, queryErrors, cancelFn, objectPattern, outTimeout)
}

// receiveResult checks for channels to receive result from nrjmx command.
func receiveResult(lineCh chan []byte, queryErrors chan error, queryCancel context.CancelFunc, objectPattern string, timeout time.Duration) (result map[string]interface{}, err error) {
	select {
	case line := <-lineCh:
		if line == nil {
			queryCancel()
			closeCmdIO()
			return nil, fmt.Errorf("got empty result for query: %s", objectPattern)
		}
		if err := json.Unmarshal(line, &result); err != nil {
			return nil, fmt.Errorf("invalid return value for query: %s, %s", objectPattern, err)
		}
	case err := <-cmdErr: // Will receive an error if the nrjmx tool exited prematurely
		if strings.HasPrefix(err.Error(), "WARNING") {
			warnings = append(warnings, err.Error())
		} else {
			return nil, err
		}
	case err := <-queryErrors: // Will receive an error if we failed while reading query output
		return nil, err
	case <-time.After(timeout):
		// In case of timeout, we want to close the command to avoid mixing up results coming up latter
		queryCancel()
		closeCmdIO()
		return nil, fmt.Errorf("timeout while waiting for query: %s", objectPattern)
	}
	return result, nil
}

func flushWarnings() {
	for w := range warnings {
		_, _ = os.Stderr.WriteString(string(w) + "\n")
	}
}
