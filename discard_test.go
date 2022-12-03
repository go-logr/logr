/*
Copyright 2020 The logr Authors.

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
	"errors"
	"reflect"
	"testing"
)

func TestDiscard(t *testing.T) {
	l := Discard()
	if l.GetSink() != nil {
		t.Error("did not return the expected underlying type")
	}
	// Verify that none of the methods panic, there is not more we can test.
	l.WithName("discard").WithValues("z", 5).Info("Hello world")
	l.Info("Hello world", "x", 1, "y", 2)
	l.V(1).Error(errors.New("foo"), "a", 123)
	if l.Enabled() {
		t.Error("discard loggers must always be disabled")
	}
}

func TestComparable(t *testing.T) {
	a := Discard()
	if !reflect.TypeOf(a).Comparable() {
		t.Fatal("discard loggers must be comparable")
	}

	b := Discard()
	if a != b {
		t.Fatal("any two discard Loggers must be equal")
	}

	c := Discard().V(2)
	if b != c {
		t.Fatal("any two discard Loggers must be equal")
	}
}
