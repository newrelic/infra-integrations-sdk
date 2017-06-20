package log

import (
	"os"

	"github.com/Sirupsen/logrus"
)

// SetupLogging redirect logs to stderr and configures the log level.
func SetupLogging(verbose bool) {
	logrus.SetOutput(os.Stderr)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func Debug(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}

func Info(format string, args ...interface{}) {
	logrus.Infof(format, args...)
}

func Warn(format string, args ...interface{}) {
	logrus.Warnf(format, args...)
}

func Error(format string, args ...interface{}) {
	logrus.Errorf(format, args...)
}

func Fatal(err error) {
	logrus.WithError(err).Fatal("can't continue")
}
