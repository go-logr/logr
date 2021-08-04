/*
Copyright 2019 The logr Authors.

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

package testing

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
)

// NewTestLogger returns a logr.Logger that prints through a testing.T object.
// Info logs are only enabled at V(0).
func NewTestLogger(t *testing.T) logr.Logger {
	fn := func(prefix, args string) {
		t.Helper()
		t.Logf("%s: %s", prefix, args)
	}
	return funcr.New(fn, funcr.Options{Helper: t.Helper})
}
