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
	"log/slog"
)

// SlogImplementor is an interface that a logr.LogSink can implement
// to support efficient logging through the slog.Logger API.
type SlogImplementor interface {
	// GetSlogHandler returns a handler which uses the same settings as the logr.LogSink.
	GetSlogHandler() slog.Handler
}

// LogrImplementor is an interface that a slog.Handler can implement
// to support efficient logging through the logr.Logger API.
type LogrImplementor interface {
	// GetLogrLogSink returns a sink which uses the same settings as the slog.Handler.
	GetLogrLogSink() LogSink
}

// ToSlog returns a slog.Logger which writes to the same backend as the logr.Logger.
func ToSlog(logger Logger) *slog.Logger {
	return slog.New(ToSlogHandler(logger))
}

// ToSlog returns a slog.Handler which writes to the same backend as the logr.Logger.
func ToSlogHandler(logger Logger) slog.Handler {
	if slogImplementor, ok := logger.GetSink().(SlogImplementor); ok {
		handler := slogImplementor.GetSlogHandler()
		return handler
	}

	return &slogHandler{sink: logger.GetSink()}
}

// FromSlog returns a logr.Logger which writes to the same backend as the slog.Logger.
func FromSlog(logger *slog.Logger) Logger {
	return FromSlogHandler(logger.Handler())
}

// FromSlog returns a logr.Logger which writes to the same backend as the slog.Handler.
func FromSlogHandler(handler slog.Handler) Logger {
	if logrImplementor, ok := handler.(LogrImplementor); ok {
		logSink := logrImplementor.GetLogrLogSink()
		return New(logSink)
	}

	return New(&slogSink{handler: handler})
}

func levelFromSlog(level slog.Level) int {
	if level >= 0 {
		// logr has no level lower than 0, so we have to truncate.
		return 0
	}
	return int(-level)
}

func levelToSlog(level int) slog.Level {
	// logr starts at info = 0 and higher values go towards debugging (negative in slog).
	return slog.Level(-level)
}
