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

package logr

import (
	"context"
	"log/slog"
	"runtime"
	"time"
)

type slogSink struct {
	callDepth int
	prefix    string
	handler   slog.Handler
}

func (l *slogSink) Init(info RuntimeInfo) {
	l.callDepth = info.CallDepth
}

func (l *slogSink) WithCallDepth(depth int) LogSink {
	newLogger := *l
	newLogger.callDepth += depth
	return &newLogger
}

func (l *slogSink) Enabled(level int) bool {
	return l.handler.Enabled(context.Background(), levelToSlog(level))
}

func (l *slogSink) Info(level int, msg string, kvList ...interface{}) {
	if l.prefix != "" {
		msg = l.prefix + ": " + msg
	}
	record := slog.NewRecord(time.Now(), levelToSlog(level), msg, l.infoOrErrorCaller())
	record.Add(kvList...)
	l.handler.Handle(context.Background(), record)
}

func (l *slogSink) Error(err error, msg string, kvList ...interface{}) {
	if l.prefix != "" {
		msg = l.prefix + ": " + msg
	}
	record := slog.NewRecord(time.Now(), slog.LevelError, msg, l.infoOrErrorCaller())
	if err != nil {
		record.Add("err", err)
	}
	record.Add(kvList...)
	l.handler.Handle(context.Background(), record)
}

func (l *slogSink) WithName(name string) LogSink {
	new := *l
	if l.prefix != "" {
		new.prefix = l.prefix + "/"
	}
	new.prefix += name
	return &new
}

func (l *slogSink) WithValues(kvList ...interface{}) LogSink {
	new := *l
	new.handler = l.handler.WithAttrs(kvListToAttrs(kvList...))
	return &new
}

// infoOrErrorCaller is a helper for Info and Error.
func (l *slogSink) infoOrErrorCaller() uintptr {
	var pcs [1]uintptr
	// skip [runtime.Callers, this function, Info/Error, and all helper functions above that.
	runtime.Callers(4+l.callDepth, pcs[:])
	return pcs[0]
}

func kvListToAttrs(kvList ...interface{}) []slog.Attr {
	// We don't need the record itself, only its Add method.
	record := slog.NewRecord(time.Time{}, 0, "", 0)
	record.Add(kvList...)
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	return attrs
}

var _ LogSink = &slogSink{}
var _ CallDepthLogSink = &slogSink{}
