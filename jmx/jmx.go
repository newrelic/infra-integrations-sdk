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
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
)

var lock sync.Mutex
var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdError io.ReadCloser
var cmdIn io.WriteCloser
var cmdExitErr = make(chan error, 1000)
var done sync.WaitGroup
var cmdWarn = make(chan string, 1000)

var (
	jmxCommand = "/usr/bin/nrjmx"
	// ErrJmxCmdRunning error returned when trying to Open and nrjmx command is still running
	ErrJmxCmdRunning = errors.New("JMX tool is already running")
)

const (
	jmxLineBuffer = 4 * 1024 * 1024 // Max 4MB per line. If single lines are outputting more JSON than that, we likely need smaller-scoped JMX queries
)

// connectionConfig is the configuration for the nrjmx command.
type connectionConfig struct {
	hostname              string
	port                  string
	uriPath               string
	username              string
	password              string
	keyStore              string
	keyStorePassword      string
	trustStore            string
	trustStorePassword    string
	remote                bool
	remoteJBossStandalone bool
}

func (cfg *connectionConfig) isSSL() bool {
	return cfg.keyStore != "" && cfg.keyStorePassword != "" && cfg.trustStore != "" && cfg.trustStorePassword != ""
}

func (cfg *connectionConfig) command() []string {
	c := make([]string, 0)
	if os.Getenv("NR_JMX_TOOL") != "" {
		c = strings.Split(os.Getenv("NR_JMX_TOOL"), " ")
	} else {
		c = []string{jmxCommand}
	}

	c = append(c, "--hostname", cfg.hostname, "--port", cfg.port)
	if cfg.uriPath != "" {
		c = append(c, "--uriPath", cfg.uriPath)
	}
	if cfg.username != "" && cfg.password != "" {
		c = append(c, "--username", cfg.username, "--password", cfg.password)
	}
	if cfg.remote {
		c = append(c, "--remote")
	}
	if cfg.remoteJBossStandalone {
		c = append(c, "--remoteJBossStandalone")
	}
	if cfg.isSSL() {
		c = append(c, "--keyStore", cfg.keyStore, "--keyStorePassword", cfg.keyStorePassword, "--trustStore", cfg.trustStore, "--trustStorePassword", cfg.trustStorePassword)
	}

	return c
}

// Open executes a nrjmx command using the given options.
func Open(hostname, port, username, password string, opts ...Option) error {
	config := &connectionConfig{
		hostname: hostname,
		port:     port,
		username: username,
		password: password,
	}

	for _, opt := range opts {
		opt(config)
	}

	return openConnection(config)
}

// Option sets an option on integration level.
type Option func(config *connectionConfig)

//WithURIPath for specifying non standard(jmxrmi) path on jmx service uri
func WithURIPath(uriPath string) Option {
	return func(config *connectionConfig) {
		config.uriPath = uriPath
	}
}

// WithSSL for SSL connection configuration.
func WithSSL(keyStore, keyStorePassword, trustStore, trustStorePassword string) Option {
	return func(config *connectionConfig) {
		config.keyStore = keyStore
		config.keyStorePassword = keyStorePassword
		config.trustStore = trustStore
		config.trustStorePassword = trustStorePassword
	}
}

// WithRemoteProtocol uses the remote JMX protocol URL (by default on JBoss Domain-mode).
func WithRemoteProtocol() Option {
	return func(config *connectionConfig) {
		config.remote = true
	}
}

// WithRemoteStandAloneJBoss uses the remote JMX protocol URL on JBoss Standalone-mode.
func WithRemoteStandAloneJBoss() Option {
	return func(config *connectionConfig) {
		config.remote = true
		config.remoteJBossStandalone = true
	}
}

func openConnection(config *connectionConfig) (err error) {
	lock.Lock()
	defer lock.Unlock()

	if cmd != nil {
		return ErrJmxCmdRunning
	}

	// Drain error channel to prevent showing past errors
	cmdExitErr = make(chan error, 1000)

	done.Add(1)

	var ctx context.Context

	cliCommand := config.command()

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

	go func() {
		scanner := bufio.NewScanner(cmdError)
		scanner.Buffer([]byte{}, jmxLineBuffer)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				break
			}
			scanner.Scan()

			line := scanner.Text()
			if strings.HasPrefix(line, "WARNING") {
				cmdWarn <- fmt.Sprintf("cannot read nrjmx stderr: %s [proc-state: %s] (line: %s)", err, cmd.ProcessState, line)
			}
		}
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err = cmd.Wait(); err != nil {
			if err != nil {
				cmdExitErr <- fmt.Errorf("nrjmx error: %s [proc-state: %s]", err, cmd.ProcessState)
			}
		}

		lock.Lock()
		defer lock.Unlock()
		cmd = nil

		done.Done()
	}()

	return nil
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	lock.Lock()

	if cmd == nil {
		lock.Unlock()
		return
	}

	cancel()
	_ = cmdIn.Close()
	_ = cmdError.Close()

	lock.Unlock()

	done.Wait()
}

func doQuery(ctx context.Context, out chan []byte, queryErrC chan error, queryString []byte) {
	lock.Lock()
	if _, err := cmdIn.Write(queryString); err != nil {
		lock.Unlock()
		queryErrC <- fmt.Errorf("writing query string: %s", err.Error())
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
		// not EOF & error
		if err := scanner.Err(); err != nil {
			queryErrC <- fmt.Errorf("error reading output from JMX tool: %v", err)
		}
	}
}

// Query executes JMX query against nrjmx tool waiting up to timeout (in milliseconds)
func Query(objectPattern string, timeoutMillis int) (result map[string]interface{}, err error) {
	ctx, cancelFn := context.WithCancel(context.Background())

	lineCh := make(chan []byte)
	queryErrors := make(chan error)
	outTimeout := time.Duration(timeoutMillis) * time.Millisecond
	// Send the query async to the underlying process so we can timeout it
	go doQuery(ctx, lineCh, queryErrors, []byte(fmt.Sprintf("%s\n", objectPattern)))

	return receiveResult(lineCh, queryErrors, cancelFn, objectPattern, outTimeout)
}

// receiveResult checks for channels to receive result from nrjmx command.
func receiveResult(lineCh chan []byte, queryErrors chan error, cancelFn context.CancelFunc, objectPattern string, timeout time.Duration) (result map[string]interface{}, err error) {
	defer Close()
	gotResult := false
	for {
		select {
		case line := <-lineCh:
			gotResult = true
			if len(line) == 0 {
				cancelFn()
				log.Warn(fmt.Sprintf("empty result for query: %s", objectPattern))
				continue
			}
			if err = json.Unmarshal(line, &result); err != nil {
				err = fmt.Errorf("invalid return value for query: %s, error: %s", objectPattern, err)
				return
			}

		case err = <-cmdExitErr:
			gotResult = true
			return

		case err = <-queryErrors:
			gotResult = true

		case w := <-cmdWarn:
			log.Warn(w)

		case <-time.After(timeout):
			// In case of timeout, we want to close the command to avoid mixing up results coming up latter
			cancelFn()
			if !gotResult {
				err = fmt.Errorf("timeout waiting for query: %s", objectPattern)
			}
			return
		}
	}
}
