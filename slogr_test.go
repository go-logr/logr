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
	"fmt"
	"io"
	"log/slog"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/go-logr/logr/internal/testhelp"
)

var debugWithoutTime = &slog.HandlerOptions{
	ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "time" {
			return slog.Attr{}
		}
		return a
	},
	Level: slog.LevelDebug,
}

func TestWithCallDepth(t *testing.T) {
	debugWithCaller := *debugWithoutTime
	debugWithCaller.AddSource = true
	var buffer bytes.Buffer
	logger := FromSlogHandler(slog.NewTextHandler(&buffer, &debugWithCaller))

	logHelper := func(logger Logger) {
		logger.WithCallDepth(1).Info("hello")
	}

	logHelper(logger)
	_, file, line, _ := runtime.Caller(0)
	expectedSource := fmt.Sprintf("%s:%d", path.Base(file), line-1)
	actual := buffer.String()
	if !strings.Contains(actual, expectedSource) {
		t.Errorf("expected log entry with %s as caller source code location, got instead:\n%s", expectedSource, actual)
	}
}

func TestRunSlogTestsOnSlogSink(t *testing.T) {
	// This proves that slogSink passes slog's own tests.
	testhelp.RunSlogTests(t, func(buffer *bytes.Buffer) slog.Handler {
		handler := slog.NewJSONHandler(buffer, nil)
		logger := FromSlogHandler(handler)
		return ToSlogHandler(logger)
	})
}

func TestSlogSinkOnDiscard(_ *testing.T) {
	// Compile-test
	logger := slog.New(ToSlogHandler(Discard()))
	logger.WithGroup("foo").With("x", 1).Info("hello")
}

func TestConversion(t *testing.T) {
	d := Discard()
	d2 := FromSlogHandler(ToSlogHandler(d))
	expectEqual(t, d, d2)

	e := Logger{}
	e2 := FromSlogHandler(ToSlogHandler(e))
	expectEqual(t, e, e2)

	text := slog.NewTextHandler(io.Discard, nil)
	text2 := ToSlogHandler(FromSlogHandler(text))
	expectEqual(t, text, text2)

	text3 := ToSlogHandler(FromSlogHandler(text).V(1))
	if handler, ok := text3.(interface {
		GetLevel() slog.Level
	}); ok {
		expectEqual(t, handler.GetLevel(), slog.Level(1))
	} else {
		t.Errorf("Expected a slogHandler which implements V(1), got instead: %T %+v", text3, text3)
	}
}

func expectEqual(t *testing.T, expected, actual any) {
	if expected != actual {
		t.Helper()
		t.Errorf("Expected %T %+v, got instead: %T %+v", expected, expected, actual, actual)
	}
}
