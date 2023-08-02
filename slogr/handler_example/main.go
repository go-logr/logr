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

// Package main is an example of using funcr.
package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-logr/logr/funcr"
	"github.com/go-logr/logr/slogr"
)

type e struct {
	str string
}

func (e e) Error() string {
	return e.str
}

func helper(log *slog.Logger, msg string) {
	helper2(log, msg)
}

func helper2(log *slog.Logger, msg string) {
	log.Info(msg)
}

func main() {
	logrLogger := funcr.New(
		func(pfx, args string) { fmt.Println(pfx, args) },
		funcr.Options{
			LogCaller:    funcr.All,
			LogTimestamp: true,
			Verbosity:    1,
		})
	log := slog.New(slogr.NewSlogHandler(logrLogger))
	example(log)
}

func example(log *slog.Logger) {
	log = log.With("saved", "value")
	log.Info("1) hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.Log(context.TODO(), slog.Level(-1), "2) you should see this")
	log.Log(context.TODO(), slog.Level(-2), "you should NOT see this")
	log.Error("3) uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error("4) goodbye", "code", -1, "err", e{"an error occurred"})
	helper(log, "5) thru a helper")
}
