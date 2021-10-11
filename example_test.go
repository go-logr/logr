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

package logr_test

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// NewStdoutLogger returns a logr.Logger that prints to stdout.
func NewStdoutLogger() logr.Logger {
	return funcr.New(func(prefix, args string) {
		if prefix != "" {
			_ = fmt.Sprintf("%s: %s\n", prefix, args)
		} else {
			fmt.Println(args)
		}
	}, funcr.Options{})
}

func Example() {
	l := NewStdoutLogger()
	l.Info("default info log", "stringVal", "value", "intVal", 12345)
	l.V(0).Info("V(0) info log", "stringVal", "value", "intVal", 12345)
	l.Error(fmt.Errorf("an error"), "error log", "stringVal", "value", "intVal", 12345)
	// Output:
	// "level"=0 "msg"="default info log" "stringVal"="value" "intVal"=12345
	// "level"=0 "msg"="V(0) info log" "stringVal"="value" "intVal"=12345
	// "msg"="error log" "error"="an error" "stringVal"="value" "intVal"=12345
}
