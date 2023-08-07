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

// Package slogr enables usage of a slog.Handler with logr.Logger as front-end
// API and of a logr.LogSink through the slog.Handler and thus slog.Logger
// APIs.
//
// Both approaches are currently experimental and need further work.
package slogr

import (
	"log/slog"

	"github.com/go-logr/logr"
)

// NewLogr returns a logr.Logger which writes to the slog.Handler.
//
// The logr verbosity level is mapped to slog levels such that V(0) becomes
// slog.LevelInfo and V(4) becomes slog.LevelDebug.
func NewLogr(handler slog.Handler) logr.Logger {
	return logr.New(&slogSink{handler: handler})
}

// NewSlogHandler returns a slog.Handler which writes to the same sink as the logr.Logger.
//
// The returned logger writes all records with level >= slog.LevelError as
// error log entries with LogSink.Error, regardless of the verbosity level of
// the logr.Logger:
//
//	logger := <some logr.Logger with 0 as verbosity level>
//	slog.New(NewSlogHandler(logger.V(10))).Error(...) -> logSink.Error(...)
//
// The level of all other records gets reduced by the verbosity
// level of the logr.Logger and the result is negated. If it happens
// to be negative, then it gets replaced by zero because a LogSink
// is not expected to handled negative levels:
//
//	slog.New(NewSlogHandler(logger)).Debug(...) -> logger.GetSink().Info(level=4, ...)
//	slog.New(NewSlogHandler(logger)).Warning(...) -> logger.GetSink().Info(level=0, ...)
//	slog.New(NewSlogHandler(logger)).Info(...) -> logger.GetSink().Info(level=0, ...)
//	slog.New(NewSlogHandler(logger.V(4))).Info(...) -> logger.GetSink().Info(level=4, ...)
func NewSlogHandler(logger logr.Logger) slog.Handler {
	// This offset currently (Go 1.21.0) works for slog.New(NewSlogHandler(...)).Info.
	// There's no guarantee that the call chain won't change and wrapping
	// the handler will also break unwinding, but it's still better than not
	// adjusting at all.
	logger = logger.WithCallDepth(2)
	return &slogHandler{sink: logger.GetSink(), levelBias: slog.Level(logger.GetV())}
}
