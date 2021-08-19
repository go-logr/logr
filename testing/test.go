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

package testing

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// NewTestLogger returns a logr.Logger that prints through a testing.T object.
// Info logs are only enabled at V(0).
func NewTestLogger(t *testing.T) logr.Logger {
	l := &testlogger{
		Formatter: funcr.NewFormatter(funcr.Options{}),
		t:         t,
	}
	return logr.New(l)
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

// NewTestLoggerWithOptions returns a logr.Logger that prints through a testing.T object.
// In contrast to the simpler NewTestLogger, output formatting can be configured.
func NewTestLoggerWithOptions(t *testing.T, opts Options) logr.Logger {
	l := &testlogger{
		Formatter: funcr.NewFormatter(funcr.Options{
			LogTimestamp: opts.LogTimestamp,
			Verbosity:    opts.Verbosity,
		}),
		t: t,
	}
	return logr.New(l)
}

// Underlier exposes access to the underlying testing.T instance. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
type Underlier interface {
	GetUnderlying() *testing.T
}

type testlogger struct {
	funcr.Formatter
	t *testing.T
}

func (l testlogger) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l testlogger) WithValues(kvList ...interface{}) logr.LogSink {
	l.Formatter.AddValues(kvList)
	return &l
}

func (l testlogger) GetCallStackHelper() func() {
	return l.t.Helper
}

func (l testlogger) Info(level int, msg string, kvList ...interface{}) {
	prefix, args := l.FormatInfo(level, msg, kvList)
	l.t.Helper()
	if prefix != "" {
		l.t.Logf("%s: %s", prefix, args)
	} else {
		l.t.Log(args)
	}
}

func (l testlogger) Error(err error, msg string, kvList ...interface{}) {
	prefix, args := l.FormatError(err, msg, kvList)
	l.t.Helper()
	if prefix != "" {
		l.t.Logf("%s: %s", prefix, args)
	} else {
		l.t.Log(args)
	}
}

func (l testlogger) GetUnderlying() *testing.T {
	return l.t
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &testlogger{}
var _ logr.CallStackHelperLogSink = &testlogger{}
var _ Underlier = &testlogger{}
