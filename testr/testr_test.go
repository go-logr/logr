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

package testr

import (
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/go-logr/logr"
)

func TestLogger(t *testing.T) {
	log := New(t)
	log.Info("info")
	log.V(0).Info("V(0).info")
	log.V(1).Info("v(1).info")
	log.Error(fmt.Errorf("error"), "error")
	log.WithName("testing").WithValues("value", "test").Info("with prefix")
	log.WithName("testing").Error(fmt.Errorf("error"), "with prefix")
	Helper(log, "hello world")

	log = NewWithOptions(t, Options{
		LogTimestamp: true,
		Verbosity:    1,
	})
	log.V(1).Info("v(1).info with options")

	underlier, ok := log.GetSink().(Underlier)
	if !ok {
		t.Fatal("couldn't get underlier")
	}
	if t != underlier.GetUnderlying() {
		t.Error("invalid underlier")
	}
}

func TestLoggerInterface(t *testing.T) {
	log := NewWithInterface(t, Options{})
	log.Info("info")
	log.V(0).Info("V(0).info")
	log.V(1).Info("v(1).info")
	log.Error(fmt.Errorf("error"), "error")
	log.WithName("testing").WithValues("value", "test").Info("with prefix")
	log.WithName("testing").Error(fmt.Errorf("error"), "with prefix")
	Helper(log, "hello world")

	underlier, ok := log.GetSink().(UnderlierInterface)
	if !ok {
		t.Fatal("couldn't get underlier")
	}
	underlierT, ok := underlier.GetUnderlying().(*testing.T)
	if !ok {
		t.Fatal("couldn't get underlying *testing.T")
	}
	if t != underlierT {
		t.Error("invalid underlier")
	}
}

func TestLoggerTestingB(t *testing.T) {
	b := &testing.B{}
	_ = NewWithInterface(b, Options{})
}

func Helper(log logr.Logger, msg string) {
	helper, log := log.WithCallStackHelper()
	helper()
	helper2(log, msg)
}

func helper2(log logr.Logger, msg string) {
	helper, log := log.WithCallStackHelper()
	helper()
	log.Info(msg)
}

var testStopWG sync.WaitGroup

func TestMain(m *testing.M) {
	exitCode := m.Run()
	testStopWG.Wait()

	os.Exit(exitCode)
}

func TestStop(t *testing.T) {
	// This test is set up so that a subtest spawns a goroutine and that
	// goroutine logs through ktesting *after* the subtest has
	// completed. This is not supported by testing.T.Log and normally
	// leads to:
	//   panic: Log in goroutine after TestGoroutines/Sub has completed: INFO hello world
	//
	// It works with testr if (and only if) logging gets discarded
	// before returning from the test.

	var wg sync.WaitGroup
	wg.Add(1)
	testStopWG.Add(1)
	logger := New(t)
	go func() {
		defer testStopWG.Done()

		// Wait for test to have returned.
		wg.Wait()

		// This output must be discarded because the test has
		// completed.
		logger.Info("simple info message")
		logger.Error(nil, "error message")
		logger.WithName("me").WithValues("completed", true).Info("complex info message", "anotherValue", 1)
	}()
	// Allow goroutine above to proceed.
	wg.Done()
}
