/*
Copyright 2022 The logr Authors.

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
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

const iterationsPerOp = 100

var variants = []struct {
	name             string
	withName         string
	withValuesBefore []interface{}
	withValuesAfter  []interface{}

	expectedOutput string
}{
	{
		// The simple benchmarks produce this output at least once. More elaborate
		// versions add additional key/value pairs.
		name:           "simple",
		expectedOutput: `{"logger":"","level":0,"msg":"ping","i":1,"j":2,"string":"hello world","int":1,"float":1}`,
	},
	{
		// One key/value pair injected through WithValues before retrieval from context.
		name:             "WithValuesBefore",
		withValuesBefore: []interface{}{"before", "abc"},
		expectedOutput:   `{"logger":"","level":0,"msg":"ping","before":"abc","i":1,"j":2,"string":"hello world","int":1,"float":1}`,
	},
	{
		// One key/value pair injected through WithValues after retrieval from context.
		name:            "WithValuesAfter",
		withValuesAfter: []interface{}{"after", "abc"},
		expectedOutput:  `{"logger":"","level":0,"msg":"ping","i":1,"j":2,"after":"abc","string":"hello world","int":1,"float":1}`,
	},
	{
		// A name gets added after retrieval from the context. The
		// corresponding case with a name added before adding the
		// Logger to the context is less interesting and not covered.
		name:           "WithName",
		withName:       "some-logger",
		expectedOutput: `{"logger":"some-logger","level":0,"msg":"ping","i":1,"j":2,"string":"hello world","int":1,"float":1}`,
	},
	{
		// A name and values get added after retrieval from the context.
		name:            "WithNameAndValuesAfter",
		withName:        "some-logger",
		withValuesAfter: []interface{}{"after", "abc"},
		expectedOutput:  `{"logger":"some-logger","level":0,"msg":"ping","i":1,"j":2,"after":"abc","string":"hello world","int":1,"float":1}`,
	},
	{
		// A name gets added after retrieval from the context and values before.
		name:             "WithNameAndValuesBefore",
		withName:         "some-logger",
		withValuesBefore: []interface{}{"before", "abc"},
		expectedOutput:   `{"logger":"some-logger","level":0,"msg":"ping","before":"abc","i":1,"j":2,"string":"hello world","int":1,"float":1}`,
	},
}

// 1% of the Info calls are invoked, all of those call the LogSink.
func BenchmarkNewContext1Percent(b *testing.B) {

	for _, variant := range variants {
		b.Run(variant.name, func(b *testing.B) {
			expectedCalls := int64(iterationsPerOp) / 100 * int64(b.N)
			ctx := setup(b, expectedCalls, variant.expectedOutput, variant.withValuesBefore)

			// Each iteration is expected to do exactly the same thing,
			// in particular do the same number of allocs.
			for i := 0; i < b.N; i++ {
				// Therefore we repeat newContext a certain number of
				// times. Individual repetitions are allowed to sometimes log
				// and sometimes not, but the overall execution is the same for
				// every outer loop iteration.
				for j := 0; j < iterationsPerOp; j++ {
					run(ctx, logSomeEntries(j, 100, 0, variant.withName, variant.withValuesAfter))
				}
			}
		})
	}
}

// 100% of the Info calls are invoked, none of those call the LogSink.
func BenchmarkNewContext100PercentDisabled(b *testing.B) {
	expectedCalls := int64(0)

	for _, variant := range variants {
		b.Run(variant.name, func(b *testing.B) {
			ctx := setup(b, expectedCalls, variant.expectedOutput, variant.withValuesBefore)

			for i := 0; i < b.N; i++ {
				for j := 0; j < iterationsPerOp; j++ {
					run(ctx, logSomeEntries(j, 1, 2 /* not logged by default by funcr */, variant.withName, variant.withValuesAfter))
				}
			}
		})
	}
}

// 100% of the Info calls are invoked, all of those call the LogSink.
func BenchmarkNewContext100Percent(b *testing.B) {
	for _, variant := range variants {
		b.Run(variant.name, func(b *testing.B) {
			expectedCalls := int64(b.N) * iterationsPerOp
			ctx := setup(b, expectedCalls, variant.expectedOutput, variant.withValuesBefore)

			for i := 0; i < b.N; i++ {
				for j := 0; j < iterationsPerOp; j++ {
					run(ctx, logSomeEntries(j, 1, 0, variant.withName, variant.withValuesAfter))
				}
			}
		})
	}
}

// 100% of the Info calls are invoked, with 10 Info calls per NewContext calls.
func BenchmarkNewContextMany(b *testing.B) {
	for _, variant := range variants {
		b.Run(variant.name, func(b *testing.B) {
			callsPerOp := 10
			expectedCalls := int64(b.N) * iterationsPerOp * int64(callsPerOp)
			ctx := setup(b, expectedCalls, variant.expectedOutput, variant.withValuesBefore)

			for i := 0; i < b.N; i++ {
				for j := 0; j < iterationsPerOp; j++ {
					run(ctx, logMultipleTimes(callsPerOp, variant.withName, variant.withValuesAfter))
				}
			}
		})
	}
}

type contextKey1 struct{}
type contextKey2 struct{}

func run(ctx context.Context, op func(ctx context.Context)) {
	// This is the currently recommended way of adding a value to a context
	// and ensuring that all future log calls include it.  Trace IDs might
	// get handled like this.
	logger := loggerFromContextOrDie(ctx)
	logger = logger.WithValues("i", 1, "j", 2)
	ctx = context.WithValue(ctx, contextKey1{}, 1)
	ctx = context.WithValue(ctx, contextKey2{}, 2)
	ctx = logr.NewContext(ctx, logger)
	op(ctx)
}

func logSomeEntries(j, mod, v int, withName string, withValues []interface{}) func(ctx context.Context) {
	return func(ctx context.Context) {
		if j%mod == 0 {
			logger := loggerFromContextOrDie(ctx)
			if withName != "" {
				logger = logger.WithName(withName)
			}
			if len(withValues) > 0 {
				logger = logger.WithValues(withValues...)
			}
			logger.V(v).Info("ping", "string", "hello world", "int", 1, "float", 1.0)
		}
	}
}

func logMultipleTimes(count int, withName string, withValues []interface{}) func(ctx context.Context) {
	return func(ctx context.Context) {
		logger := loggerFromContextOrDie(ctx)
		if withName != "" {
			logger = logger.WithName(withName)
		}
		if len(withValues) > 0 {
			logger = logger.WithValues(withValues...)
		}
		for i := 0; i < count; i++ {
			logger.Info("ping", "string", "hello world", "int", 1, "float", 1.0)
		}
	}
}

func setup(tb testing.TB, expectedCalls int64, expectedOutput string, withValues []interface{}) context.Context {
	var actualCalls int64
	tb.Cleanup(func() {
		if actualCalls != expectedCalls {
			tb.Errorf("expected %d calls to Info, got %d", expectedCalls, actualCalls)
		}
	})
	logger := funcr.NewJSON(func(actualOutput string) {
		if actualOutput != expectedOutput {
			tb.Fatalf("expected %s, got %s", expectedOutput, actualOutput)
		}
		actualCalls++
	}, funcr.Options{})
	if len(withValues) > 0 {
		logger = logger.WithValues(withValues...)
	}
	return logr.NewContext(context.Background(), logger)
}

func loggerFromContextOrDie(ctx context.Context) logr.Logger {
	logger, err := logr.FromContext(ctx)
	if err != nil {
		panic("failed to get Logger from Context")
	}
	return logger
}

func TestFromContext(t *testing.T) {
	expectedCalls := int64(iterationsPerOp) / 100

	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			ctx := setup(t, expectedCalls, variant.expectedOutput, variant.withValuesBefore)
			run(ctx, logMultipleTimes(1, variant.withName, variant.withValuesAfter))
		})
	}
}
