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
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
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

func ExampleFromSlog() {
	logger := logr.FromSlog(slog.New(slog.NewTextHandler(os.Stdout, debugWithoutTime)))

	logger.Info("hello world")
	logger.Error(errors.New("fake error"), "ignore me")
	logger.WithValues("x", 1, "y", 2).WithValues("str", "abc").WithName("foo").WithName("bar").V(4).Info("with values, verbosity and name")

	// Output:
	// level=INFO msg="hello world"
	// level=ERROR msg="ignore me" err="fake error"
	// level=DEBUG msg="foo/bar: with values, verbosity and name" x=1 y=2 str=abc
}

func ExampleToSlog() {
	logger := logr.ToSlog(funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Fprintln(os.Stdout, prefix, args)
		} else {
			fmt.Fprintln(os.Stdout, args)
		}
	}, funcr.Options{}))

	logger.Info("hello world")
	logger.Error("ignore me", "err", errors.New("fake error"))
	logger.With("x", 1, "y", 2).WithGroup("group").With("str", "abc").Warn("with values and group")

	// Output:
	// "level"=0 "msg"="hello world"
	// "msg"="ignore me" "error"=null "err"="fake error"
	// "level"=0 "msg"="with values and group" "x"=1 "y"=2 "group.str"="abc"
}
