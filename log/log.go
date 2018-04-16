package log

import (
	"os"

	"log"
	"io"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	SetDebug(enable bool) // deprecated TODO remove when deleting global scope from cache
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

// Debugf logs a formatted message at level Debug.
func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	if l.debug {
		l.prefixPrint("DEBUG", format, args)
	}
}

// Infof logs a formatted message at level Info.
func (l *defaultLogger) Infof(format string, args ...interface{}) {
	l.prefixPrint("INFO", format, args)
}

// Errorf logs a formatted message at level Error.
func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	l.prefixPrint("ERR", format, args)
}

func (l *defaultLogger) prefixPrint(prefix string, format string, args ...interface{}) {
	prev := log.Prefix()
	log.SetPrefix(prefix)
	l.logger.Printf(format, args...)
	log.SetPrefix(prev)
}

// deprecated TODO remove when deleting global scope from cache
func (l *defaultLogger) SetDebug(enable bool) {
	l.debug = enable
}