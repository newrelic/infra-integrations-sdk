package jmx

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdIn io.WriteCloser
var cmdErr = make(chan error, 1)
var done sync.WaitGroup

var jmxCommand = "./bin/nrjmx"

const (
	outTimeout    time.Duration = 1000 * time.Millisecond
	jmxLineBuffer               = 4 * 1024 * 1024 // Max 4MB per line. If single lines are outputting more JSON than that, we likely need smaller-scoped JMX queries
)

func getCommand(hostname, port, username, password string) []string {
	var cliCommand []string

	if os.Getenv("NR_JMX_TOOL") != "" {
		cliCommand = strings.Split(os.Getenv("NR_JMX_TOOL"), " ")
	} else {
		cliCommand = []string{jmxCommand}
	}

	cliCommand = append(
		cliCommand, "--hostname", hostname, "--port", port,
		"--username", username, "--password", password,
	)

	return cliCommand
}

// Open will start the nrjmx command with the provided connection parameters.
func Open(hostname, port, username, password string) error {
	if cmd != nil {
		return fmt.Errorf("JMX tool is already running with PID: %d", cmd.Process.Pid)
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
	// Avoid stupid errors/warnings b/c cancel is not used in this method
	_ = cancel

	cmd = exec.CommandContext(ctx, cliCommand[0], cliCommand[1:]...)

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
		if err = cmd.Wait(); err != nil {
			cmdErr <- fmt.Errorf("JMX tool exited with error: %s", err)
		}
		done.Done()
		cmd = nil
	}()

	return nil
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	if cmd == nil {
		return
	}

	cmdIn.Close()
	cancel()
	done.Wait()
}

func doQuery(out chan []byte, errorChan chan error, queryString []byte) {
	cmdIn.Write(queryString)

	scanner := bufio.NewScanner(cmdOut)
	scanner.Buffer([]byte{}, jmxLineBuffer) // Override default buffer to increase buffer size

	if scanner.Scan() {
		out <- scanner.Bytes()
	} else {
		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("Error reading output from JMX tool: %v", err)
		} else {
			// If scanner.Scan() returns false but err is also nil, it hit EOF. We consider that a problem, so we should return an error.
			errorChan <- fmt.Errorf("Got an EOF while reading JMX tool output")
		}
	}
}

// Query returns a map with the attribute names and its values from the nrjmx tool
func Query(objectPattern string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	pipe := make(chan []byte)
	queryErrors := make(chan error)

	// Send the query async to the underlying process so we can timeout it
	go doQuery(pipe, queryErrors, []byte(fmt.Sprintf("%s\n", objectPattern)))

	select {
	case line := <-pipe:
		if line == nil {
			Close()
			return nil, fmt.Errorf("Got empty result for query: %s", objectPattern)
		}
		if err := json.Unmarshal(line, &result); err != nil {
			return nil, fmt.Errorf("Invalid return value for query: %s, %s", objectPattern, err)
		}
	case err := <-cmdErr: // Will receive an error if the nrjmx tool exited prematurely
		return nil, err
	case err := <-queryErrors: // Will receive an error if we failed while reading query output
		return nil, err
	case <-time.After(outTimeout):
		// In case of timeout, we want to close the command to avoid mixing up results coming up latter
		Close()
		return nil, fmt.Errorf("Timeout while waiting for query: %s", objectPattern)
	}
	return result, nil
}
