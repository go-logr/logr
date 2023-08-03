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
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/slogr"
)

type e struct {
	str string
}

func (e e) Error() string {
	return e.str
}

func helper(log logr.Logger, msg string) {
	helper2(log, msg)
}

func helper2(log logr.Logger, msg string) {
	log.WithCallDepth(2).Info(msg)
}

func main() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.Level(-1),
	}
	handler := slog.NewJSONHandler(os.Stderr, &opts)
	log := slogr.New(handler)
	example(log)
}

func example(log logr.Logger) {
	log = log.WithName("my")
	log = log.WithName("logger")
	log = log.WithName("name")
	log = log.WithValues("saved", "value")
	log.Info("1) hello", "val1", 1, "val2", map[string]int{"k": 1})
	log.V(1).Info("2) you should see this")
	log.V(1).V(1).Info("you should NOT see this")
	log.Error(nil, "3) uh oh", "trouble", true, "reasons", []float64{0.1, 0.11, 3.14})
	log.Error(e{"an error occurred"}, "4) goodbye", "code", -1)
	helper(log, "5) thru a helper")
}
