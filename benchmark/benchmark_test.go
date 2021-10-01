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
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

//go:noinline
func doInfoOneArg(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		log.Info("this is", "a", "string")
	}
}

//go:noinline
func doInfoSeveralArgs(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		log.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doInfoWithValues(b *testing.B, log logr.Logger) {
	log = log.WithValues("k1", "str", "k2", 222, "k3", true, "k4", 1.0)
	for i := 0; i < b.N; i++ {
		log.Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doV0Info(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		log.V(0).Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doV9Info(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		log.V(9).Info("multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doError(b *testing.B, log logr.Logger) {
	err := fmt.Errorf("error message")
	for i := 0; i < b.N; i++ {
		log.Error(err, "multi",
			"bool", true, "string", "str", "int", 42,
			"float", 3.14, "struct", struct{ X, Y int }{93, 76})
	}
}

//go:noinline
func doWithValues(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		l := log.WithValues("k1", "v1", "k2", "v2")
		_ = l
	}
}

//go:noinline
func doWithName(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		l := log.WithName("name")
		_ = l
	}
}

//go:noinline
func doWithCallDepth(b *testing.B, log logr.Logger) {
	for i := 0; i < b.N; i++ {
		l := log.WithCallDepth(1)
		_ = l
	}
}

func BenchmarkDiscardLogInfoOneArg(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doInfoOneArg(b, log)
}

func BenchmarkDiscardLogInfoSeveralArgs(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doInfoSeveralArgs(b, log)
}

func BenchmarkDiscardLogInfoWithValues(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doInfoWithValues(b, log)
}

func BenchmarkDiscardLogV0Info(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doV0Info(b, log)
}

func BenchmarkDiscardLogV9Info(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doV9Info(b, log)
}

func BenchmarkDiscardLogError(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doError(b, log)
}

func BenchmarkDiscardWithValues(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doWithValues(b, log)
}

func BenchmarkDiscardWithName(b *testing.B) {
	var log logr.Logger = logr.Discard()
	doWithName(b, log)
}

func noopKV(prefix, args string) {}
func noopJSON(obj string)        {}

func BenchmarkFuncrLogInfoOneArg(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doInfoOneArg(b, log)
}

func BenchmarkFuncrJSONLogInfoOneArg(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doInfoOneArg(b, log)
}

func BenchmarkFuncrLogInfoSeveralArgs(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doInfoSeveralArgs(b, log)
}

func BenchmarkFuncrJSONLogInfoSeveralArgs(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doInfoSeveralArgs(b, log)
}

func BenchmarkFuncrLogInfoWithValues(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doInfoWithValues(b, log)
}

func BenchmarkFuncrJSONLogInfoWithValues(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doInfoWithValues(b, log)
}

func BenchmarkFuncrLogV0Info(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doV0Info(b, log)
}

func BenchmarkFuncrJSONLogV0Info(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doV0Info(b, log)
}

func BenchmarkFuncrLogV9Info(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doV9Info(b, log)
}

func BenchmarkFuncrJSONLogV9Info(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doV9Info(b, log)
}

func BenchmarkFuncrLogError(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doError(b, log)
}

func BenchmarkFuncrJSONLogError(b *testing.B) {
	var log logr.Logger = funcr.NewJSON(noopJSON, funcr.Options{})
	doError(b, log)
}

func BenchmarkFuncrWithValues(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doWithValues(b, log)
}

func BenchmarkFuncrWithName(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doWithName(b, log)
}

func BenchmarkFuncrWithCallDepth(b *testing.B) {
	var log logr.Logger = funcr.New(noopKV, funcr.Options{})
	doWithCallDepth(b, log)
}
