package log

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestSetupLogging(t *testing.T) {
	SetupLogging(false)
	if logrus.GetLevel() != logrus.InfoLevel {
		t.Error("level should be info when verbose is false")
	}

	SetupLogging(true)
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Error("level should be debug when verbose is true")
	}
}

func TestConfigureLogger(t *testing.T) {
	l := logrus.New()
	l.Out = ioutil.Discard

	ConfigureLogger(l, false)
	if l.Level != logrus.InfoLevel {
		t.Error("level should be info when verbose is false")
	}

	ConfigureLogger(l, true)
	if l.Level != logrus.DebugLevel {
		t.Error("level should be debug when verbose is true")
	}

	if l.Out != os.Stderr {
		t.Error("logger out should be stderr")
	}
}
