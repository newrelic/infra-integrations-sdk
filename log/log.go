package log

import (
	"os"

	"log"
	"io"
)

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

type defaultLogger struct {
	logger *log.Logger
	debug  bool
}

// New creates a logger with stderr output, verbose enables Debug output
func NewStdErr(debug bool) Logger {
	return New(debug, os.Stderr)
}

// New creates a logger using provided writer, verbose enables Debug output
func New(debug bool, w io.Writer) Logger {
	return &defaultLogger{
		logger: log.New(w, "", 0),
		debug:  debug,
	}
}

// Debug logs a formatted message at level Debug.
func (l *defaultLogger) Debug(format string, args ...interface{}) {
	if l.debug {
		l.prefixPrint("DEBUG", format, args)
	}
}

// Info logs a formatted message at level Info.
func (l *defaultLogger) Info(format string, args ...interface{}) {
	l.prefixPrint("INFO", format, args)
}

// Error logs a formatted message at level Error.
func (l *defaultLogger) Error(format string, args ...interface{}) {
	l.prefixPrint("ERR", format, args)
}

func (l *defaultLogger) prefixPrint(prefix string, format string, args ...interface{}) {
	prev := log.Prefix()
	log.SetPrefix(prefix)
	l.logger.Printf(format, args...)
	log.SetPrefix(prev)
}
