package log

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestSetupLogging(t *testing.T) {
	SetupLogging(false)
	if logrus.GetLevel() != logrus.InfoLevel {
		t.Error()
	}

	SetupLogging(true)
	if logrus.GetLevel() != logrus.DebugLevel {
		t.Error()
	}
}
