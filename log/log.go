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

func (l *defaultLogger) prefixPrint(prefix string, format string, args ...interface{}) {
	prev := log.Prefix()
	log.SetPrefix(prefix)
	l.logger.Printf(format, args...)
	log.SetPrefix(prev)
}
