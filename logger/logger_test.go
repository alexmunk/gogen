package logging

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestLogLevel(t *testing.T) {
	if DefaultLogLevel != logrus.ErrorLevel {
		t.Fatalf("Log not set to Error")
	}
}
