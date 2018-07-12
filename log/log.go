package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Logger defines a facade for a simple logger
type Logger interface {
	Debugf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var (
	// Discard provides a discard all policy logger
	Discard = New(false, ioutil.Discard)
)

type defaultLogger struct {
	logger *log.Logger
	debug  bool
}

// NewStdErr creates a logger with stderr output, the argument enables Debug (verbose) output
func NewStdErr(debug bool) Logger {
	return New(debug, os.Stderr)
}

// New creates a logger using the provided writer. The 'debug' argument enables Debug (verbose) output
func New(debug bool, w io.Writer) Logger {
	return &defaultLogger{
		logger: log.New(w, "", log.Lmicroseconds),
		debug:  debug,
	}
}

// Debugf logs a formatted message at level Debug.
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	if l.debug {
		l.prefixPrint("DEBUG", format, args...)
	}
}

// Infof logs a formatted message at level Info.
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.prefixPrint("INFO", format, args...)
}

// Errorf logs a formatted message at level Error.
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.prefixPrint("ERR", format, args...)
}

// Warnf logs a formatted message at level Warning.
func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	l.prefixPrint("WARN", format, args...)
}

func (l *defaultLogger) prefixPrint(prefix string, format string, args ...interface{}) {
	l.logger.Printf(fmt.Sprintf("[%s] %s", prefix, format), args...)
}

// Global logger instance to provide easy access and retro-compatibility.

var globalLogger defaultLogger

func init() {
	SetupLogging(false)
}

// SetupLogging redirects global logs to stderr and configures the log level.
func SetupLogging(verbose bool) {
	globalLogger = defaultLogger{
		logger: log.New(os.Stderr, "", 0),
		debug:  verbose,
	}
}

// Debug logs a formatted message at level Debug.
func Debug(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs a formatted message at level Info.
func Info(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a formatted message at level Warn.
func Warn(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs a formatted message at level Error.
func Error(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatal logs an error at level Fatal, and makes the program exit with an error code.
func Fatal(err error) {
	globalLogger.prefixPrint("FATAL", "can't continue: %v", err)
	panic(err)
}
