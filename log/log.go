package log

import (
	"os"

	"io"
	"io/ioutil"
	"log"
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
	Discard = NewWriter(false, ioutil.Discard)
)

type defaultLogger struct {
	logger *log.Logger
	debug  bool
}

// NewStdErr creates a logger with stderr output, the argument enables Debug (verbose) output
func NewStdErr(debug bool) Logger {
	return NewWriter(debug, os.Stderr)
}

// NewWriter creates a logger using the provided writer. The 'debug' argument enables Debug (verbose) output
func NewWriter(debug bool, w io.Writer) Logger {
	return &defaultLogger{
		logger: log.New(w, "", 0),
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
	prev := log.Prefix()
	log.SetPrefix(prefix)
	l.logger.Printf(format, args...)
	log.SetPrefix(prev)
}

// Deprecated methods, kept to do v2 to v3 migration less painful

var globalLogger defaultLogger

func init() {
	SetupLogging(false)
}

// SetupLogging redirects global logs to stderr and configures the log level.
// Deprecated. Use log.NewWriter, log.NewStdErr or any custom implementation of the log.Logger interface.
func SetupLogging(verbose bool) {
	globalLogger = defaultLogger{
		logger: log.New(os.Stderr, "", 0),
		debug:  verbose,
	}
}

// New creates an already configured logger.
// Deprecated. Use log.NewWriter, log.NewStdErr or any custom implementation of the log.Logger interface.
func New(verbose bool) Logger {
	return NewStdErr(verbose)
}

// ConfigureLogger configures an already created logger. Redirects logs to stderr and configures the log level.
// Deprecated. Use log.NewWriter, log.NewStdErr or any custom implementation of the log.Logger interface.
func ConfigureLogger(logger Logger, verbose bool) {
	if logImpl, ok := logger.(*defaultLogger); ok {
		*logImpl = *New(verbose).(*defaultLogger)

		//logImpl.logger = log.New(os.Stderr, "", 0)
		//logImpl.debug = verbose
	}
}

// Debug logs a formatted message at level Debug.
// Deprecated. Use Debugf function of the log.Logger interface.
func Debug(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs a formatted message at level Info.
// Deprecated. Use Infof function of the log.Logger interface.
func Info(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a formatted message at level Warn.
// Deprecated. Use the log.Logger interface.
func Warn(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs a formatted message at level Error.
// Deprecated. Use Errorf function of the log.Logger interface.
func Error(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatal logs an error at level Fatal, and makes the program exit with an error code.
// Deprecated. Use the log.Logger interface.
func Fatal(err error) {
	globalLogger.prefixPrint("FATAL", "can't continue: %v", err)
	panic(err)
}
