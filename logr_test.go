/*
Copyright 2021 The logr Authors.

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
	"testing"
)

// testLogSink is a Logger just for testing that does nothing.
type testLogSink struct{}

func (l *testLogSink) Init(RuntimeInfo) {
}

func (l *testLogSink) Enabled(int) bool {
	return false
}

func (l *testLogSink) Info(level int, msg string, keysAndValues ...interface{}) {
}

func (l *testLogSink) Error(err error, msg string, keysAndValues ...interface{}) {
}

func (l *testLogSink) WithValues(keysAndValues ...interface{}) LogSink {
	return l
}

func (l *testLogSink) WithName(name string) LogSink {
	return l
}

// Verify that it actually implements the interface
var _ LogSink = &testLogSink{}

func TestContext(t *testing.T) {
	ctx := context.TODO()

	if out, err := FromContext(ctx); err == nil {
		t.Errorf("expected error, got %#v", out)
	}
	out := FromContextOrDiscard(ctx)
	if _, ok := out.sink.(discardLogger); !ok {
		t.Errorf("expected a discardLogger, got %#v", out)
	}

	sink := &testLogSink{}
	logger := New(sink)
	lctx := NewContext(ctx, logger)
	if out, err := FromContext(lctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if p := out.sink.(*testLogSink); p != sink {
		t.Errorf("expected output to be the same as input: got in=%p, out=%p", sink, p)
	}
	out = FromContextOrDiscard(lctx)
	if p := out.sink.(*testLogSink); p != sink {
		t.Errorf("expected output to be the same as input: got in=%p, out=%p", sink, p)
	}
}

// testCallDepthLogSink is a Logger just for testing that does nothing.
type testCallDepthLogSink struct {
	*testLogSink
	depth int
}

func (l *testCallDepthLogSink) WithCallDepth(depth int) LogSink {
	return &testCallDepthLogSink{l.testLogSink, l.depth + depth}
}

// Verify that it actually implements the interface
var _ CallDepthLogSink = &testCallDepthLogSink{}

func TestWithCallDepth(t *testing.T) {
	// Test an impl that does not support it.
	t.Run("not supported", func(t *testing.T) {
		in := &testLogSink{}
		l := New(in)
		out := l.WithCallDepth(42)
		if p := out.sink.(*testLogSink); p != in {
			t.Errorf("expected output to be the same as input: got in=%p, out=%p", in, p)
		}
	})

	// Test an impl that does support it.
	t.Run("supported", func(t *testing.T) {
		in := &testCallDepthLogSink{&testLogSink{}, 0}
		l := New(in)
		out := l.WithCallDepth(42)
		if out.sink.(*testCallDepthLogSink) == in {
			t.Errorf("expected output to be different than input: got in=out=%p", in)
		}
		if cdl := out.sink.(*testCallDepthLogSink); cdl.depth != 42 {
			t.Errorf("expected depth=42, got %d", cdl.depth)
		}
	})
}
