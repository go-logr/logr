//go:build go1.21
// +build go1.21

/*
Copyright 2023 The logr Authors.

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

// Package slogr provides bridges between github.com/go-logr/logr and Go's
// slog package.
package slogr

import (
	"context"
	"log/slog"
	"runtime"
	"time"

	"github.com/go-logr/logr"
)

// New returns a logr.Logger which logs through an arbitrary slog.Handler.
func New(handler slog.Handler) logr.Logger {
	return logr.New(newSink(handler))
}

// Underlier exposes access to the underlying logging function. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
type Underlier interface {
	GetUnderlying() slog.Handler
}

func newSink(handler slog.Handler) logr.LogSink {
	sink := &slogSink{
		handler: handler,
	}
	// For skipping logr.Logger.Info and .Error.
	return sink.WithCallDepth(1)
}

// slogSink inherits some of its LogSink implementation from Formatter
// and just needs to add some glue code.
type slogSink struct {
	handler slog.Handler
	name    string
	depth   int
}

// Init configures this Formatter from runtime info, such as the call depth
// imposed by logr itself.
// Note that this receiver is a pointer, so depth can be saved.
func (sink *slogSink) Init(info logr.RuntimeInfo) {
	sink.depth += info.CallDepth
}

// Enabled checks whether an info message at the given level should be logged.
func (sink slogSink) Enabled(level int) bool {
	return sink.handler.Enabled(context.Background(), slog.Level(-level))
}

func (sink slogSink) WithName(name string) logr.LogSink {
	if len(sink.name) > 0 {
		sink.name += "/"
	}
	sink.name += name
	return &sink
}

func (sink slogSink) WithValues(kvList ...interface{}) logr.LogSink {
	r := slog.NewRecord(time.Time{}, 0, "", 0)
	r.Add(kvList...)
	attrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	sink.handler = sink.handler.WithAttrs(attrs)
	return &sink
}

func (sink slogSink) WithCallDepth(depth int) logr.LogSink {
	sink.depth += depth
	return &sink
}

func (sink slogSink) Info(level int, msg string, kvList ...interface{}) {
	args := make([]interface{}, 0, len(kvList)+4)
	if len(sink.name) != 0 {
		args = append(args, "logger", sink.name)
	}
	args = append(args, "vlevel", level)
	args = append(args, kvList...)

	sink.log(slog.Level(-level), msg, args...)
}

func (sink slogSink) Error(err error, msg string, kvList ...interface{}) {
	args := make([]interface{}, 0, len(kvList)+4)
	if len(sink.name) != 0 {
		args = append(args, "logger", sink.name)
	}
	args = append(args, "err", err)
	args = append(args, kvList...)

	sink.log(slog.LevelError, msg, args...)
}

func (sink slogSink) log(level slog.Level, msg string, kvList ...interface{}) {
	// TODO: Should this be optional via HandlerOptions?  Literally
	// HandlerOptions or something like it?
	var pcs [1]uintptr
	runtime.Callers(sink.depth+2, pcs[:]) // 2 = this frame and Info/Error
	pc := pcs[0]
	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(kvList...)
	sink.handler.Handle(context.Background(), r)
}

func (sink slogSink) GetUnderlying() slog.Handler {
	return sink.handler
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &slogSink{}
var _ logr.CallDepthLogSink = &slogSink{}
var _ Underlier = &slogSink{}
