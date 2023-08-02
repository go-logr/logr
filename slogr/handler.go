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

package slogr

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
)

// NewSlogHandler returns a slog.Handler which logs through an arbitrary
// logr.Logger.
func NewSlogHandler(logger logr.Logger) slog.Handler {
	return &logrHandler{
		// Account for extra wrapper functions.
		logger: logger.WithCallDepth(3),
	}
}

// logrHandler implements slog.Handler in terms of a logr.Logger.
type logrHandler struct {
	logger logr.Logger
}

func (h logrHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.logger.V(int(-level)).Enabled()
}

func (h logrHandler) Handle(_ context.Context, record slog.Record) error {
	//FIXME: I don't know what to do with record.Time or record.PC.  Neither of
	// them map to logr.
	args := recordToKV(record)
	if record.Level >= slog.LevelError {
		h.logger.Error(nil, record.Message, args...)
	} else {
		h.logger.V(int(-record.Level)).Info(record.Message, args...)
	}
	return nil
}

func recordToKV(record slog.Record) []any {
	kv := make([]any, 0, record.NumAttrs()*2)
	fn := func(attr slog.Attr) bool {
		kv = append(kv, attr.Key, attr.Value.Any())
		return true
	}
	record.Attrs(fn)
	return kv
}

func attrsToKV(attrs []slog.Attr) []any {
	kv := make([]any, 0, len(attrs)*2)
	for _, attr := range attrs {
		kv = append(kv, attr.Key, attr.Value.Any())
	}
	return kv
}

func (h logrHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	args := attrsToKV(attrs)
	h.logger = h.logger.WithValues(args...)
	return h
}

func (h logrHandler) WithGroup(name string) slog.Handler {
	//FIXME: I don't know how to implement this in logr, but it's an
	//interesting idea
	return h
}

var _ slog.Handler = logrHandler{}
