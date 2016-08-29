package testing

import (
	"testing"

	"github.com/thockin/logr"
)

// TestLogger is a logr.Logger that prints through a testing.T object.
// Only error logs will have any effect.
type TestLogger struct {
	T *testing.T
}

var _ logr.Logger = TestLogger{}

func (_ TestLogger) Info(_ ...interface{}) {
	// Do nothing.
}

func (_ TestLogger) Infof(_ string, _ ...interface{}) {
	// Do nothing.
}

func (_ TestLogger) Enabled() bool {
	return false
}

func (log TestLogger) Error(args ...interface{}) {
	log.T.Log(args...)
}

func (log TestLogger) Errorf(format string, args ...interface{}) {
	log.T.Logf(format, args...)
}

func (log TestLogger) V(v int) logr.InfoLogger {
	return log
}

func (log TestLogger) NewWithPrefix(_ string) logr.Logger {
	return log
}
