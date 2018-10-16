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

// Debug log to debug log level
func Debug(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

// Info log to info log level
func Info(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

// Warn log to warn log level
func Warn(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

// Error log to error log level
func Error(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

// Fatal log to fatal log level and exits
func Fatal(err error) {
	logrus.WithError(err).Fatal("can't continue")
}
