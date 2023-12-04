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

package logr_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
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

func ExampleFromSlogHandler() {
	logrLogger := logr.FromSlogHandler(slog.NewTextHandler(os.Stdout, debugWithoutTime))

	logrLogger.Info("hello world")
	logrLogger.Error(errors.New("fake error"), "ignore me")
	logrLogger.WithValues("x", 1, "y", 2).WithValues("str", "abc").WithName("foo").WithName("bar").V(4).Info("with values, verbosity and name")

	// Output:
	// level=INFO msg="hello world"
	// level=ERROR msg="ignore me" err="fake error"
	// level=DEBUG msg="with values, verbosity and name" x=1 y=2 str=abc logger=foo/bar
}

func ExampleToSlogHandler() {
	funcrLogger := funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Fprintln(os.Stdout, prefix, args)
		} else {
			fmt.Fprintln(os.Stdout, args)
		}
	}, funcr.Options{
		Verbosity: 10,
	})

	slogLogger := slog.New(logr.ToSlogHandler(funcrLogger))
	slogLogger.Info("hello world")
	slogLogger.Error("ignore me", "err", errors.New("fake error"))
	slogLogger.With("x", 1, "y", 2).WithGroup("group").With("str", "abc").Warn("with values and group")

	slogLogger = slog.New(logr.ToSlogHandler(funcrLogger.V(int(-slog.LevelDebug))))
	slogLogger.Info("info message reduced to debug level")

	// Output:
	// "level"=0 "msg"="hello world"
	// "msg"="ignore me" "error"=null "err"="fake error"
	// "level"=0 "msg"="with values and group" "x"=1 "y"=2 "group"={"str"="abc"}
	// "level"=4 "msg"="info message reduced to debug level"
}

func TestWithCallDepth(t *testing.T) {
	debugWithCaller := *debugWithoutTime
	debugWithCaller.AddSource = true
	var buffer bytes.Buffer
	logger := logr.FromSlogHandler(slog.NewTextHandler(&buffer, &debugWithCaller))

	logHelper := func(logger logr.Logger) {
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
		logger := logr.FromSlogHandler(handler)
		return logr.ToSlogHandler(logger)
	})
}

func TestSlogSinkOnDiscard(_ *testing.T) {
	// Compile-test
	logger := slog.New(logr.ToSlogHandler(logr.Discard()))
	logger.WithGroup("foo").With("x", 1).Info("hello")
}

func TestConversion(t *testing.T) {
	d := logr.Discard()
	d2 := logr.FromSlogHandler(logr.ToSlogHandler(d))
	expectEqual(t, d, d2)

	e := logr.Logger{}
	e2 := logr.FromSlogHandler(logr.ToSlogHandler(e))
	expectEqual(t, e, e2)

	text := slog.NewTextHandler(io.Discard, nil)
	text2 := logr.ToSlogHandler(logr.FromSlogHandler(text))
	expectEqual(t, text, text2)

	text3 := logr.ToSlogHandler(logr.FromSlogHandler(text).V(1))
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
