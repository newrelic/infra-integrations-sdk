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

// Error vars to ease Query response handling.
var (
	ErrBeanPattern = errors.New("cannot parse MBean glob pattern, valid: 'DOMAIN:BEAN'")
	ErrConnection  = errors.New("jmx endpoint connection error")
)

var cmd *exec.Cmd
var cancel context.CancelFunc
var cmdOut io.ReadCloser
var cmdError io.ReadCloser
var cmdIn io.WriteCloser
var cmdErrC = make(chan error, cmdStdChanLen)
var cmdWarnC = make(chan string, cmdStdChanLen)
var done sync.WaitGroup

var (
	// DefaultNrjmxExec default nrjmx tool executable path
	DefaultNrjmxExec = "/usr/bin/nrjmx"
	// ErrJmxCmdRunning error returned when trying to Open and nrjmx command is still running
	ErrJmxCmdRunning = errors.New("JMX tool is already running")
)

// connectionConfig is the configuration for the nrjmx command.
type connectionConfig struct {
	connectionURL         string
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
	verbose               bool
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

	if cfg.connectionURL != "" {
		c = append(c, "--connURL", cfg.connectionURL)
	} else {
		c = append(c, "--hostname", cfg.hostname, "--port", cfg.port)
		if cfg.uriPath != "" {
			c = append(c, "--uriPath", cfg.uriPath)
		}
		if cfg.remote {
			c = append(c, "--remote")
		}
		if cfg.remoteJBossStandalone {
			c = append(c, "--remoteJBossStandalone")
		}
	}
	if cfg.username != "" && cfg.password != "" {
		c = append(c, "--username", cfg.username, "--password", cfg.password)
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

// OpenURL executes a nrjmx command using the provided full connection URL and options.
func OpenURL(connectionURL, username, password string, opts ...Option) error {
	opts = append(opts, WithConnectionURL(connectionURL))
	return Open("", "", username, password, opts...)
}

// Open executes a nrjmx command using the given options.
func Open(hostname, port, username, password string, opts ...Option) error {
	config := &connectionConfig{
		hostname:       hostname,
		port:           port,
		username:       username,
		password:       password,
		executablePath: DefaultNrjmxExec,
	}

	for _, opt := range opts {
		opt(config)
	}

	log.SetupLogging(config.verbose)

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

// WithURIPath for specifying non standard(jmxrmi) path on jmx service uri
func WithURIPath(uriPath string) Option {
	return func(config *connectionConfig) {
		config.uriPath = uriPath
	}
}

// WithConnectionURL for specifying non standard(jmxrmi) path on jmx service uri
func WithConnectionURL(connectionURL string) Option {
	return func(config *connectionConfig) {
		config.connectionURL = connectionURL
	}
}

// WithVerbose enables verbose mode for nrjmx.
func WithVerbose() Option {
	return func(config *connectionConfig) {
		config.verbose = true
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
	cmdErrC = make(chan error, cmdStdChanLen)
	cmdWarnC = make(chan string, cmdStdChanLen)

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
				cmdErrC <- fmt.Errorf("nrjmx error: %s [proc-state: %s]", err, cmd.ProcessState)
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
			msg := line[7:]
			if strings.Contains(msg, "Can't parse bean name") {
				cmdErrC <- ErrBeanPattern
				return
			}
			cmdWarnC <- msg
		}
		if strings.HasPrefix(line, "SEVERE:") {
			msg := line[7:]
			if strings.Contains(msg, "jmx connection error") {
				cmdErrC <- ErrConnection
			} else {
				cmdErrC <- errors.New(msg)
			}
			return
		}
		if err != nil {
			cmdErrC <- err
			return
		}
	}
}

// Close will finish the underlying nrjmx application by closing its standard
// input and canceling the execution afterwards to clean-up.
func Close() {
	if cancel != nil {
		cancel()
	}

	done.Wait()
}

func doQuery(ctx context.Context, out chan []byte, queryErrC chan error, queryString []byte) {
	if _, err := cmdIn.Write(queryString); err != nil {
		queryErrC <- fmt.Errorf("writing nrjmx stdin: %s", err.Error())
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
			queryErrC <- fmt.Errorf("reading nrjmx stdout: %s", err.Error())
		}
		out <- b
		return
	}
}

// Query executes JMX query against nrjmx tool waiting up to timeout (in milliseconds)
func Query(objectPattern string, timeoutMillis int) (result map[string]interface{}, err error) {
	ctx, cancelFn := context.WithCancel(context.Background())

	lineCh := make(chan []byte, cmdStdChanLen)
	queryErrors := make(chan error, cmdStdChanLen)
	outTimeout := time.Duration(timeoutMillis) * time.Millisecond
	// Send the query async to the underlying process so we can timeout it
	go doQuery(ctx, lineCh, queryErrors, []byte(fmt.Sprintf("%s\n", objectPattern)))

	return receiveResult(lineCh, queryErrors, cancelFn, objectPattern, outTimeout)
}

// receiveResult checks for channels to receive result from nrjmx command.
func receiveResult(lineC chan []byte, queryErrC chan error, cancelFn context.CancelFunc, objectPattern string, timeout time.Duration) (result map[string]interface{}, err error) {
	var warn string
	for {
		select {
		case line := <-lineC:
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
			return

		case warn = <-cmdWarnC:
			// change on the API is required to return warnings
			log.Warn(warn)
			return

		case err = <-cmdErrC:
			return

		case err = <-queryErrC:
			return

		case <-time.After(timeout):
			// In case of timeout, we want to close the command to avoid mixing up results coming up latter
			cancelFn()
			Close()
			err = fmt.Errorf("timeout waiting for query: %s", objectPattern)
			return
		}
	}
}
