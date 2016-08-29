package testing

import "github.com/thockin/logr"

// NullLogger is a logr.Logger that does nothing.
type NullLogger struct{}

var _ logr.Logger = NullLogger{}

func (_ NullLogger) Info(_ ...interface{}) {
	// Do nothing.
}

func (_ NullLogger) Infof(_ string, _ ...interface{}) {
	// Do nothing.
}

func (_ NullLogger) Enabled() bool {
	return false
}

func (_ NullLogger) Error(_ ...interface{}) {
	// Do nothing.
}

func (_ NullLogger) Errorf(_ string, _ ...interface{}) {
	// Do nothing.
}

func (log NullLogger) V(_ int) logr.InfoLogger {
	return log
}

func (log NullLogger) NewWithPrefix(_ string) logr.Logger {
	return log
}
