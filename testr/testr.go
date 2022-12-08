/*
Copyright 2019 The logr Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package testr provides support for using logr in tests.
package testr

import (
	"sync/atomic"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// New returns a logr.Logger that prints through a testing.T object.
// Info logs are only enabled at V(0).
func New(t *testing.T) logr.Logger {
	return NewWithOptions(t, Options{})
}

// Options carries parameters which influence the way logs are generated.
type Options struct {
	// LogTimestamp tells the logger to add a "ts" key to log
	// lines. This has some overhead, so some users might not want
	// it.
	LogTimestamp bool

	// Verbosity tells the logger which V logs to be write.
	// Higher values enable more logs.
	Verbosity int
}

// NewWithOptions returns a logr.Logger that prints through a testing.T object.
// In contrast to the simpler New, output formatting can be configured.
func NewWithOptions(t *testing.T, opts Options) logr.Logger {
	l := &testlogger{
		testloggerInterface: newLoggerInterfaceWithOptions(t, opts),
	}
	return logr.New(l)
}

// TestingT is an interface wrapper around testing.T, testing.B and testing.F.
type TestingT interface {
	Helper()
	Log(args ...interface{})
}

// cleanupT is implemented by the t instances that we get, but we don't require
// it in our API.
type cleanupT interface {
	Cleanup(func())
}

// NewWithInterface returns a logr.Logger that prints through a
// TestingT object.
// In contrast to the simpler New, output formatting can be configured.
func NewWithInterface(t TestingT, opts Options) logr.Logger {
	l := newLoggerInterfaceWithOptions(t, opts)
	return logr.New(&l)
}

var noT TestingT

func newLoggerInterfaceWithOptions(t TestingT, opts Options) testloggerInterface {
	l := testloggerInterface{
		t: new(atomic.Value),
		Formatter: funcr.NewFormatter(funcr.Options{
			LogTimestamp: opts.LogTimestamp,
			Verbosity:    opts.Verbosity,
		}),
	}
	l.t.Store(&t)
	if cleanup, ok := t.(cleanupT); ok {
		cleanup.Cleanup(func() {
			// Cannot store nil.
			l.t.Store(&noT)
		})
	}
	return l
}

// Underlier exposes access to the underlying testing.T instance. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
//
// Returns nil when the test has already completed.
type Underlier interface {
	GetUnderlying() *testing.T
}

// UnderlierInterface exposes access to the underlying TestingT instance. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
//
// Returns nil when the test has already completed.
type UnderlierInterface interface {
	GetUnderlying() TestingT
}

// Info logging implementation shared between testLogger and testLoggerInterface.
func logInfo(t TestingT, formatInfo func(int, string, []interface{}) (string, string), level int, msg string, kvList ...interface{}) {
	prefix, args := formatInfo(level, msg, kvList)
	t.Helper()
	if prefix != "" {
		args = prefix + ": " + args
	}
	t.Log(args)
}

// Error logging implementation shared between testLogger and testLoggerInterface.
func logError(t TestingT, formatError func(error, string, []interface{}) (string, string), err error, msg string, kvList ...interface{}) {
	prefix, args := formatError(err, msg, kvList)
	t.Helper()
	if prefix != "" {
		args = prefix + ": " + args
	}
	t.Log(args)
}

// This type exists to wrap and modify the method-set of testloggerInterface.
// In particular, it changes the GetUnderlying() method.
type testlogger struct {
	testloggerInterface
}

func (l testlogger) GetUnderlying() *testing.T {
	t := l.testloggerInterface.GetUnderlying()
	if t == nil {
		return nil
	}

	// This method is defined on testlogger, so the only type this could
	// possibly be is testing.T, even though that's not guaranteed by the type
	// system itself.
	return t.(*testing.T) //nolint:forcetypeassert
}

type testloggerInterface struct {
	funcr.Formatter

	// t will be cleared when the test is done because then t becomes
	// unusable. In particular t.Log will panic with "Log in goroutine
	// after ... has completed".
	t *atomic.Value
}

func (l testloggerInterface) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l testloggerInterface) WithValues(kvList ...interface{}) logr.LogSink {
	l.Formatter.AddValues(kvList)
	return &l
}

func (l testloggerInterface) GetCallStackHelper() func() {
	t := l.GetUnderlying()
	if t == nil {
		return func() {}
	}

	return t.Helper
}

func (l testloggerInterface) Info(level int, msg string, kvList ...interface{}) {
	t := l.GetUnderlying()
	if t == nil {
		return
	}

	t.Helper()
	logInfo(t, l.FormatInfo, level, msg, kvList...)
}

func (l testloggerInterface) Error(err error, msg string, kvList ...interface{}) {
	t := l.GetUnderlying()
	if t == nil {
		return
	}

	t.Helper()
	logError(t, l.FormatError, err, msg, kvList...)
}

func (l testloggerInterface) GetUnderlying() TestingT {
	t := l.t.Load()
	if t == &noT {
		return nil
	}
	// This is the only type that this could possibly be.
	return *t.(*TestingT) //nolint:forcetypeassert
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &testlogger{}
var _ logr.CallStackHelperLogSink = &testlogger{}
var _ Underlier = &testlogger{}

var _ logr.LogSink = &testloggerInterface{}
var _ logr.CallStackHelperLogSink = &testloggerInterface{}
var _ UnderlierInterface = &testloggerInterface{}
