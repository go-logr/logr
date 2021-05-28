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

package testing

import (
	"fmt"
	"testing"
)

func TestTestLogger(t *testing.T) {
	logger := NewTestLogger(t)
	logger.Info("info")
	logger.V(0).Info("V(0).info")
	logger.V(1).Info("v(1).info")
	logger.Error(fmt.Errorf("error"), "error")
}
