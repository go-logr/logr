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
)

type slogHandler struct {
	sink        LogSink
	groupPrefix string
}

func (l *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return l.sink.Enabled(levelFromSlog(level))
}

func (l *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	kvList := make([]any, 0, 2*record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		kvList = append(kvList, appendPrefix(l.groupPrefix, attr.Key), attr.Value.Any())
		return true
	})
	if record.Level >= slog.LevelError {
		l.sink.Error(nil, record.Message, kvList...)
	} else {
		l.sink.Info(levelFromSlog(record.Level), record.Message, kvList...)
	}
	return nil
}

func (l slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	kvList := make([]any, 0, 2*len(attrs))
	for _, attr := range attrs {
		kvList = append(kvList, appendPrefix(l.groupPrefix, attr.Key), attr.Value.Any())
	}
	l.sink = l.sink.WithValues(kvList...)
	return &l
}

func (l slogHandler) WithGroup(name string) slog.Handler {
	l.groupPrefix = appendPrefix(l.groupPrefix, name)
	return &l
}

func appendPrefix(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

var _ slog.Handler = &slogHandler{}
