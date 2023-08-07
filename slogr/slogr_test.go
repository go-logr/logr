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

package slogr_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/go-logr/logr/slogr"
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

func ExampleNew() {
	logger := slogr.NewLogr(slog.NewTextHandler(os.Stdout, debugWithoutTime))

	logger.Info("hello world")
	logger.Error(errors.New("fake error"), "ignore me")
	logger.WithValues("x", 1, "y", 2).WithValues("str", "abc").WithName("foo").WithName("bar").V(4).Info("with values, verbosity and name")

	// Output:
	// level=INFO msg="hello world"
	// level=ERROR msg="ignore me" err="fake error"
	// level=DEBUG msg="with values, verbosity and name" x=1 y=2 str=abc logger=foo/bar
}

func ExampleNewSlogLogger() {
	funcrLogger := funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Fprintln(os.Stdout, prefix, args)
		} else {
			fmt.Fprintln(os.Stdout, args)
		}
	}, funcr.Options{
		Verbosity: 10,
	})

	logger := slog.New(slogr.NewSlogHandler(funcrLogger))
	logger.Info("hello world")
	logger.Error("ignore me", "err", errors.New("fake error"))
	logger.With("x", 1, "y", 2).WithGroup("group").With("str", "abc").Warn("with values and group")

	logger = slog.New(slogr.NewSlogHandler(funcrLogger.V(int(-slog.LevelDebug))))
	logger.Info("info message reduced to debug level")

	// Output:
	// "level"=0 "msg"="hello world"
	// "msg"="ignore me" "error"=null "err"="fake error"
	// "level"=0 "msg"="with values and group" "x"=1 "y"=2 "group.str"="abc"
	// "level"=4 "msg"="info message reduced to debug level"
}

func TestWithCallDepth(t *testing.T) {
	debugWithCaller := *debugWithoutTime
	debugWithCaller.AddSource = true
	var buffer bytes.Buffer
	logger := slogr.NewLogr(slog.NewTextHandler(&buffer, &debugWithCaller))

	logHelper(logger)
	_, file, line, _ := runtime.Caller(0)
	expectedSource := fmt.Sprintf("%s:%d", path.Base(file), line-1)
	actual := buffer.String()
	if !strings.Contains(actual, expectedSource) {
		t.Errorf("expected log entry with %s as caller source code location, got instead:\n%s", expectedSource, actual)
	}
}

func logHelper(logger logr.Logger) {
	logger.WithCallDepth(1).Info("hello")
}

func TestSlogHandler(t *testing.T) {
	var buffer bytes.Buffer
	funcrLogger := funcr.NewJSON(func(obj string) {
		fmt.Fprintln(&buffer, obj)
	}, funcr.Options{
		LogTimestamp: true,
		Verbosity:    10,
		RenderBuiltinsHook: func(kvList []any) []any {
			mappedKVList := make([]any, len(kvList))
			for i := 0; i < len(kvList); i += 2 {
				key := kvList[i]
				switch key {
				case "ts":
					mappedKVList[i] = "time"
				default:
					mappedKVList[i] = key
				}
				mappedKVList[i+1] = kvList[i+1]
			}
			return mappedKVList
		},
	})
	handler := slogr.NewSlogHandler(funcrLogger)

	err := slogtest.TestHandler(handler, func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buffer.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	})

	// Correlating failures with individual test cases is hard with the current API.
	// See https://github.com/golang/go/issues/61758
	t.Logf("Output:\n%s", buffer.String())
	if err != nil {
		if err, ok := err.(interface {
			Unwrap() []error
		}); ok {
			for _, err := range err.Unwrap() {
				if !containsOne(err.Error(),
					"a Handler should ignore a zero Record.Time",                     // Time is generated by sink.
					"a Handler should handle Group attributes",                       // funcr doesn't.
					"a Handler should inline the Attrs of a group with an empty key", // funcr doesn't know about groups.
					"a Handler should not output groups for an empty Record",         // Relies on WithGroup. Text may change, see https://go.dev/cl/516155
					"a Handler should handle the WithGroup method",                   // logHandler does by prefixing keys, which is not what the test expects.
					"a Handler should handle multiple WithGroup and WithAttr calls",  // Same.
					"a Handler should call Resolve on attribute values in groups",    // funcr doesn't do that and slogHandler can't do it for it.
				) {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		} else {
			// Shouldn't be reached, errors from errors.Join can be split up.
			t.Errorf("Unexpected errors:\n%v", err)
		}
	}
}

func containsOne(hay string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(hay, needle) {
			return true
		}
	}
	return false
}

func TestDiscard(t *testing.T) {
	logger := slog.New(slogr.NewSlogHandler(logr.Discard()))
	logger.WithGroup("foo").With("x", 1).Info("hello")
}

func TestConversion(t *testing.T) {
	d := logr.Discard()
	d2 := slogr.NewLogr(slogr.NewSlogHandler(d))
	expectEqual(t, d, d2)

	e := logr.Logger{}
	e2 := slogr.NewLogr(slogr.NewSlogHandler(e))
	expectEqual(t, e, e2)

	f := funcr.New(func(prefix, args string) {}, funcr.Options{})
	f2 := slogr.NewLogr(slogr.NewSlogHandler(f))
	expectEqual(t, f, f2)

	text := slog.NewTextHandler(io.Discard, nil)
	text2 := slogr.NewSlogHandler(slogr.NewLogr(text))
	expectEqual(t, text, text2)

	text3 := slogr.NewSlogHandler(slogr.NewLogr(text).V(1))
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
