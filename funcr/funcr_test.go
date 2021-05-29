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

package funcr

import (
	"encoding/json"
	"fmt"
	"testing"
)

type substr string

func ptrint(i int) *int {
	return &i
}
func ptrstr(s string) *string {
	return &s
}

func TestPretty(t *testing.T) {
	cases := []interface{}{
		"strval",
		substr("substrval"),
		true,
		false,
		int(93),
		int8(93),
		int16(93),
		int32(93),
		int64(93),
		int(-93),
		int8(-93),
		int16(-93),
		int32(-93),
		int64(-93),
		uint(93),
		uint8(93),
		uint16(93),
		uint32(93),
		uint64(93),
		uintptr(93),
		float32(93.76),
		float64(93.76),
		ptrint(93),
		ptrstr("pstrval"),
		[]int{9, 3, 7, 6},
		[4]int{9, 3, 7, 6},
		struct {
			Int    int
			String string
		}{
			93, "seventy-six",
		},
		map[string]int{
			"nine": 3,
		},
		map[substr]int{
			"nine": 3,
		},
		fmt.Errorf("error"),
		struct {
			X int `json:"x"`
			Y int `json:"y"`
		}{
			93, 76,
		},
		struct {
			X []int
			Y map[int]int
			Z struct{ P, Q int }
		}{
			[]int{9, 3, 7, 6},
			map[int]int{9: 3},
			struct{ P, Q int }{9, 3},
		},
		[]struct{ X, Y string }{
			{"nine", "three"},
			{"seven", "six"},
		},
	}

	for i, tc := range cases {
		ours := pretty(tc)
		std, err := json.Marshal(tc)
		if err != nil {
			t.Errorf("[%d]: unexpected error: %v", i, err)
		}
		if ours != string(std) {
			t.Errorf("[%d]: expected %q, got %q", i, std, ours)
		}
	}
}

func TestWithName(t *testing.T) {
	var called string
	x := New(func(prefix, _ string) { called = prefix }, Options{})
	y := x.WithName("y")
	z := y.WithName("z")

	x.Info("any")
	if called != "" {
		t.Errorf("expected no prefix, got %q", called)
	}

	y.Info("any")
	if expected := "y"; called != expected {
		t.Errorf("expected %q, got %q", expected, called)
	}

	z.Info("any")
	if expected := "y/z"; called != expected {
		t.Errorf("expected %q, got %q", expected, called)
	}
}
