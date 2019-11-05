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

const (
	jmxLineInitialBuffer = 4 * 1024 // initial 4KB per line, it'll be increased when required
	cmdStdChanLen        = 1000
)

var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdError io.ReadCloser
var cmdIn io.WriteCloser
var cmdExitErr = make(chan error, cmdStdChanLen)
var done sync.WaitGroup

var (
	defaultNrjmxCommand = "/usr/bin/nrjmx"
	// ErrJmxCmdRunning error returned when trying to Open and nrjmx command is still running
	ErrJmxCmdRunning = errors.New("JMX tool is already running")
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
	executablePath        string
}

func (cfg *connectionConfig) isSSL() bool {
	return cfg.keyStore != "" && cfg.keyStorePassword != "" && cfg.trustStore != "" && cfg.trustStorePassword != ""
}

func (cfg *connectionConfig) command() []string {
	c := make([]string, 0)
	if os.Getenv("NR_JMX_TOOL") != "" {
		c = strings.Split(os.Getenv("NR_JMX_TOOL"), " ")
	} else {
		c = []string{cfg.executablePath}
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

// OpenNoAuth executes a nrjmx command without user/pass using the given options.
func OpenNoAuth(hostname, port string, opts ...Option) error {
	return Open(hostname, port, "", "", opts...)
}

// Open executes a nrjmx command using the given options.
func Open(hostname, port, username, password string, opts ...Option) error {
	config := &connectionConfig{
		hostname:       hostname,
		port:           port,
		username:       username,
		password:       password,
		executablePath: defaultNrjmxCommand,
	}

	for _, opt := range opts {
		opt(config)
	}

	return openConnection(config)
}

// Option sets an option on integration level.
type Option func(config *connectionConfig)

// WithNrJmxTool for specifying non standard `nrjmx` tool executable location.
// Has less precedence than `NR_JMX_TOOL` environment variable.
func WithNrJmxTool(executablePath string) Option {
	return func(config *connectionConfig) {
		config.executablePath = executablePath
	}
}

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
	if cmd != nil {
		return ErrJmxCmdRunning
	}

	// Drain error channel to prevent showing past errors
	cmdExitErr = make(chan error, cmdStdChanLen)

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

	go handleStdErr(ctx)

	if err = cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err = cmd.Wait(); err != nil {
			if err != nil {
				cmdExitErr <- fmt.Errorf("nrjmx error: %s [proc-state: %s]", err, cmd.ProcessState)
			}
		}

		cmd = nil

		done.Done()
	}()

	return nil
}

func handleStdErr(ctx context.Context) {
	scanner := bufio.NewReaderSize(cmdError, jmxLineInitialBuffer)

	var line string
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		line, err = scanner.ReadString('\n')
		// API needs re to allow stderr full read before closing
		if err != nil && err != io.EOF && !strings.Contains(err.Error(), "file already closed") {
			log.Error(fmt.Sprintf("error reading stderr from JMX tool: %s", err.Error()))
		}
		if strings.HasPrefix(line, "WARNING") {
			log.Warn(line[7:])
		}
		if err != nil {
			return
		}
	}
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	cancel()

	done.Wait()
}

func doQuery(ctx context.Context, out chan []byte, queryErrC chan error, queryString []byte) {
	if _, err := cmdIn.Write(queryString); err != nil {
		queryErrC <- fmt.Errorf("writing query string: %s", err.Error())
		return
	}

	scanner := bufio.NewReaderSize(cmdOut, jmxLineInitialBuffer)

	var b []byte
	var err error
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		b, err = scanner.ReadBytes('\n')
		if err != nil && err != io.EOF {
			queryErrC <- fmt.Errorf("error reading output from JMX tool: %v", err)
		}
		out <- b
		if err == io.EOF {
			return
		}
	}
}

// Query executes JMX query against nrjmx tool waiting up to timeout (in milliseconds)
func Query(objectPattern string, timeoutMillis int) (result map[string]interface{}, err error) {
	defer Close()

	ctx, cancelFn := context.WithCancel(context.Background())

	lineCh := make(chan []byte, cmdStdChanLen)
	queryErrors := make(chan error, cmdStdChanLen)
	outTimeout := time.Duration(timeoutMillis) * time.Millisecond
	// Send the query async to the underlying process so we can timeout it
	go doQuery(ctx, lineCh, queryErrors, []byte(fmt.Sprintf("%s\n", objectPattern)))

	return receiveResult(lineCh, queryErrors, cancelFn, objectPattern, outTimeout)
}

// receiveResult checks for channels to receive result from nrjmx command.
func receiveResult(lineCh chan []byte, queryErrors chan error, cancelFn context.CancelFunc, objectPattern string, timeout time.Duration) (result map[string]interface{}, err error) {
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
			var r map[string]interface{}
			if err = json.Unmarshal(line, &r); err != nil {
				err = fmt.Errorf("invalid return value for query: %s, error: %s", objectPattern, err)
				return
			}
			if result == nil {
				result = make(map[string]interface{})
			}
			for k, v := range r {
				result[k] = v
			}

		case err = <-cmdExitErr:
			gotResult = true
			return

		case err = <-queryErrors:
			gotResult = true

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
