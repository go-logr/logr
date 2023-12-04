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
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

var _ SlogSink = &testLogSink{}

// testSlogSink gets embedded in testLogSink to add slog-specific fields
// which are only available when slog is supported by Go.
type testSlogSink struct {
	attrs  []slog.Attr
	groups []string

	fnHandle    func(l *testLogSink, ctx context.Context, record slog.Record)
	fnWithAttrs func(l *testLogSink, attrs []slog.Attr)
	fnWithGroup func(l *testLogSink, name string)
}

func (l *testLogSink) Handle(ctx context.Context, record slog.Record) error {
	if l.fnHandle != nil {
		l.fnHandle(l, ctx, record)
	}
	return nil
}

func (l *testLogSink) WithAttrs(attrs []slog.Attr) SlogSink {
	if l.fnWithAttrs != nil {
		l.fnWithAttrs(l, attrs)
	}
	out := *l
	n := len(out.attrs)
	out.attrs = append(out.attrs[:n:n], attrs...)
	return &out
}

func (l *testLogSink) WithGroup(name string) SlogSink {
	if l.fnWithGroup != nil {
		l.fnWithGroup(l, name)
	}
	out := *l
	n := len(out.groups)
	out.groups = append(out.groups[:n:n], name)
	return &out
}

func withAttrs(record slog.Record, attrs ...slog.Attr) slog.Record {
	record = record.Clone()
	record.AddAttrs(attrs...)
	return record
}

func toJSON(record slog.Record) string {
	var buffer bytes.Buffer
	record.Time = time.Time{}
	handler := slog.NewJSONHandler(&buffer, nil)
	if err := handler.Handle(context.Background(), record); err != nil {
		return err.Error()
	}
	return buffer.String()
}

func TestToSlogHandler(t *testing.T) {
	lvlThreshold := 0
	actualCalledHandle := 0
	var actualRecord slog.Record

	sink := &testLogSink{}
	logger := New(sink)

	sink.fnEnabled = func(lvl int) bool {
		return lvl <= lvlThreshold
	}

	sink.fnHandle = func(l *testLogSink, ctx context.Context, record slog.Record) {
		actualCalledHandle++

		// Combine attributes from sink and call. Ordering of WithValues and WithAttrs
		// is wrong, but good enough for test cases.
		var values slog.Record
		values.Add(l.withValues...)
		var attrs []any
		add := func(attr slog.Attr) bool {
			attrs = append(attrs, attr)
			return true
		}
		values.Attrs(add)
		record.Attrs(add)
		for _, attr := range l.attrs {
			attrs = append(attrs, attr)
		}

		// Wrap them in groups - not quite correct for WithValues that
		// follows WithGroup, but good enough for test cases.
		for i := len(l.groups) - 1; i >= 0; i-- {
			attrs = []any{slog.Group(l.groups[i], attrs...)}
		}

		actualRecord = slog.Record{
			Level:   record.Level,
			Message: record.Message,
		}
		actualRecord.Add(attrs...)
	}

	verify := func(t *testing.T, expectedRecord slog.Record) {
		actual := toJSON(actualRecord)
		expected := toJSON(expectedRecord)
		if expected != actual {
			t.Errorf("JSON dump did not match, expected:\n%s\nGot:\n%s\n", expected, actual)
		}
	}

	reset := func() {
		lvlThreshold = 0
		actualCalledHandle = 0
		actualRecord = slog.Record{}
	}

	testcases := map[string]struct {
		run            func()
		expectedRecord slog.Record
	}{
		"simple": {
			func() { slog.New(ToSlogHandler(logger)).Info("simple") },
			slog.Record{Message: "simple"},
		},

		"disabled": {
			func() { slog.New(ToSlogHandler(logger.V(1))).Info("") },
			slog.Record{},
		},

		"enabled": {
			func() {
				lvlThreshold = 1
				slog.New(ToSlogHandler(logger.V(1))).Info("enabled")
			},
			slog.Record{Level: -1, Message: "enabled"},
		},

		"error": {
			func() { slog.New(ToSlogHandler(logger.V(100))).Error("error") },
			slog.Record{Level: slog.LevelError, Message: "error"},
		},

		"with-parameters": {
			func() { slog.New(ToSlogHandler(logger)).Info("", "answer", 42, "foo", "bar") },
			withAttrs(slog.Record{}, slog.Int("answer", 42), slog.String("foo", "bar")),
		},

		"with-values": {
			func() { slog.New(ToSlogHandler(logger.WithValues("answer", 42, "foo", "bar"))).Info("") },
			withAttrs(slog.Record{}, slog.Int("answer", 42), slog.String("foo", "bar")),
		},

		"with-group": {
			func() { slog.New(ToSlogHandler(logger)).WithGroup("group").Info("", "answer", 42, "foo", "bar") },
			withAttrs(slog.Record{}, slog.Group("group", slog.Int("answer", 42), slog.String("foo", "bar"))),
		},

		"with-values-and-group": {
			func() {
				slog.New(ToSlogHandler(logger.WithValues("answer", 42, "foo", "bar"))).WithGroup("group").Info("")
			},
			// Behavior of testLogSink is not quite correct here.
			withAttrs(slog.Record{}, slog.Group("group", slog.Int("answer", 42), slog.String("foo", "bar"))),
		},

		"with-group-and-values": {
			func() {
				slog.New(ToSlogHandler(logger)).WithGroup("group").With("answer", 42, "foo", "bar").Info("")
			},
			withAttrs(slog.Record{}, slog.Group("group", slog.Int("answer", 42), slog.String("foo", "bar"))),
		},

		"with-group-and-logr-values": {
			func() {
				slogLogger := slog.New(ToSlogHandler(logger)).WithGroup("group")
				logrLogger := FromSlogHandler(slogLogger.Handler()).WithValues("answer", 42, "foo", "bar")
				slogLogger = slog.New(ToSlogHandler(logrLogger))
				slogLogger.Info("")
			},
			withAttrs(slog.Record{}, slog.Group("group", slog.Int("answer", 42), slog.String("foo", "bar"))),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			tc.run()
			verify(t, tc.expectedRecord)
			reset()
		})
	}
}
