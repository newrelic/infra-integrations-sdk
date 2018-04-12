package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

// SetupLogging redirects logs to stderr and configures the log level.
func SetupLogging(verbose bool) {
	logrus.SetOutput(os.Stderr)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

// New creates an already configured logger.
func New(verbose bool) *logrus.Logger {
	l := logrus.New()
	ConfigureLogger(l, verbose)

	return l
}

// ConfigureLogger configures an already created logger. Redirects logs to stderr and configures the log level.
func ConfigureLogger(logger *logrus.Logger, verbose bool) {
	logger.Out = os.Stderr
	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
}

// Debugf logs a formatted message at level Debug.
func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

// Infof logs a formatted message at level Info.
func Infof(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// Warnf logs a formatted message at level Warn.
func Warnf(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

// Errorf logs a formatted message at level Error.
func Errorf(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(err error) {
	logrus.WithError(err).Fatal("can't continue")
}
