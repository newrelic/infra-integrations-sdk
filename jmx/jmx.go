package jmx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdIn io.WriteCloser
var cmdErr = make(chan error, 1)
var mutex = &sync.Mutex{}

var jmxCommand = "./bin/nrjmx"

const (
	outTimeout time.Duration = 1000 * time.Millisecond
)

// Open will start the nrjmx command with the provided connection parameters.
func Open(hostname, port, username, password string) error {
	if cmd != nil && cmd.ProcessState == nil {
		return fmt.Errorf("JMX tool is already running with PID: %d", cmd.Process.Pid)
	}

	var err error
	var ctx context.Context

	ctx, cancel = context.WithCancel(context.Background())
	// Avoid stupid errors/warnings b/c cancel is not used in this method
	_ = cancel

	if os.Getenv("NR_JMX_TOOL") != "" {
		jmxCommand = os.Getenv("NR_JMX_TOOL")
	}

	cmd = exec.CommandContext(
		ctx, jmxCommand,
		"--hostname", hostname, "--port", port,
		"--username", username, "--password", password,
	)

	if cmdOut, err = cmd.StdoutPipe(); err != nil {
		return err
	}
	if cmdIn, err = cmd.StdinPipe(); err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		cmdErr <- cmd.Wait()
	}()

	return status()
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	if cmd == nil {
		return
	}

	cmdIn.Close()
	cancel()
	cmd.Wait()
	cmd = nil
}

// status checks if the current NR JMX background process has exited and, returns
// the exit error if set.
func status() error {
	select {
	default:
		return nil
	case err := <-cmdErr:
		if err != nil {
			return fmt.Errorf("JMX tool exited with error: %s", err)
		}
	}

	return nil
}

// reload cancels the current NR JMX background process and spawns a new one with
// the same arguments passed to jmx.Open
func reload() error {
	if cmd == nil {
		return fmt.Errorf("JMX tool is not running, call Open before performing any query")
	}
	cancel()
	cmd.Wait()
	return Open(cmd.Args[2], cmd.Args[4], cmd.Args[6], cmd.Args[8])
}

func doQuery(out chan []byte, queryString []byte) {
	cmdIn.Write(queryString)

	scanner := bufio.NewScanner(cmdOut)
	if scanner.Scan() {
		out <- scanner.Bytes()
	}
	close(out)
}

// Query returns a map with the attribute names and its values from the nrjmx tool
func Query(objectPattern string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	pipe := make(chan []byte, 1)

	mutex.Lock()
	defer mutex.Unlock()

	// Send the query async to the underlying process so we can timeout it
	go doQuery(pipe, []byte(fmt.Sprintf("%s\n", objectPattern)))

	select {
	case line := <-pipe:
		if line == nil {
			return nil, status()
		}
		if err := json.Unmarshal(line, &result); err != nil {
			return nil, fmt.Errorf("Invalid return value for query: %s, %s", objectPattern, err)
		}
	case <-time.After(outTimeout):
		// In case of timeout, we want to reset the command to avoid mixing up results coming up latter
		reload()
		return nil, fmt.Errorf("Timeout while waiting for query: %s", objectPattern)
	}
	return result, nil
}
