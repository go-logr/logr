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
	"path/filepath"
	"runtime"
	"testing"

	"github.com/go-logr/logr"
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

func makeKV(args ...interface{}) []interface{} {
	return args
}

func TestFlatten(t *testing.T) {
	testCases := []struct {
		name   string
		kv     []interface{}
		expect string
	}{{
		name:   "nil",
		kv:     nil,
		expect: "",
	}, {
		name:   "empty",
		kv:     []interface{}{},
		expect: "",
	}, {
		name:   "primitives",
		kv:     makeKV("int", 1, "str", "ABC", "bool", true),
		expect: `"int"=1 "str"="ABC" "bool"=true`,
	}, {
		name:   "missing value",
		kv:     makeKV("key"),
		expect: `"key"="<no-value>"`,
	}, {
		name:   "non-string key",
		kv:     makeKV(123, "val"),
		expect: `"<non-string-key-0>"="val"`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := flatten(tc.kv...)
			if r != tc.expect {
				t.Errorf("expected %q, got %q", tc.expect, r)
			}
		})
	}
}

func TestEnabled(t *testing.T) {
	t.Run("default V", func(t *testing.T) {
		log := newSink(func(prefix, args string) {}, Options{})
		if !log.Enabled(0) {
			t.Errorf("expected true")
		}
		if log.Enabled(1) {
			t.Errorf("expected false")
		}
	})
	t.Run("V=9", func(t *testing.T) {
		log := newSink(func(prefix, args string) {}, Options{Verbosity: 9})
		if !log.Enabled(8) {
			t.Errorf("expected true")
		}
		if !log.Enabled(9) {
			t.Errorf("expected true")
		}
		if log.Enabled(10) {
			t.Errorf("expected false")
		}
	})
}

type capture struct {
	log string
}

func (c *capture) Func(prefix, args string) {
	c.log = prefix + " " + args
}

func TestInfo(t *testing.T) {
	testCases := []struct {
		name   string
		args   []interface{}
		expect string
	}{{
		name:   "just msg",
		args:   makeKV(),
		expect: ` "level"=0 "msg"="msg"`,
	}, {
		name:   "primitives",
		args:   makeKV("int", 1, "str", "ABC", "bool", true),
		expect: ` "level"=0 "msg"="msg" "int"=1 "str"="ABC" "bool"=true`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			sink.Info(0, "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestInfoWithCaller(t *testing.T) {
	t.Run("LogCaller=All", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: All})
		sink.Info(0, "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "level"=0 "msg"="msg"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Info", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: Info})
		sink.Info(0, "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "level"=0 "msg"="msg"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Error", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: Error})
		sink.Info(0, "msg")
		expect := ` "level"=0 "msg"="msg"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=None", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: None})
		sink.Info(0, "msg")
		expect := ` "level"=0 "msg"="msg"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
}

func TestError(t *testing.T) {
	testCases := []struct {
		name   string
		args   []interface{}
		expect string
	}{{
		name:   "just msg",
		args:   makeKV(),
		expect: ` "msg"="msg" "error"="err"`,
	}, {
		name:   "primitives",
		args:   makeKV("int", 1, "str", "ABC", "bool", true),
		expect: ` "msg"="msg" "error"="err" "int"=1 "str"="ABC" "bool"=true`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			sink.Error(fmt.Errorf("err"), "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestErrorWithCaller(t *testing.T) {
	t.Run("LogCaller=All", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: All})
		sink.Error(fmt.Errorf("err"), "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "msg"="msg" "error"="err"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Error", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: Error})
		sink.Error(fmt.Errorf("err"), "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "msg"="msg" "error"="err"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Info", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: Info})
		sink.Error(fmt.Errorf("err"), "msg")
		expect := ` "msg"="msg" "error"="err"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=None", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: None})
		sink.Error(fmt.Errorf("err"), "msg")
		expect := ` "msg"="msg" "error"="err"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
}

func TestInfoWithName(t *testing.T) {
	testCases := []struct {
		name   string
		names  []string
		args   []interface{}
		expect string
	}{{
		name:   "one",
		names:  []string{"pfx1"},
		args:   makeKV("k", "v"),
		expect: `pfx1 "level"=0 "msg"="msg" "k"="v"`,
	}, {
		name:   "two",
		names:  []string{"pfx1", "pfx2"},
		args:   makeKV("k", "v"),
		expect: `pfx1/pfx2 "level"=0 "msg"="msg" "k"="v"`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			for _, n := range tc.names {
				sink = sink.WithName(n)
			}
			sink.Info(0, "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestErrorWithName(t *testing.T) {
	testCases := []struct {
		name   string
		names  []string
		args   []interface{}
		expect string
	}{{
		name:   "one",
		names:  []string{"pfx1"},
		args:   makeKV("k", "v"),
		expect: `pfx1 "msg"="msg" "error"="err" "k"="v"`,
	}, {
		name:   "two",
		names:  []string{"pfx1", "pfx2"},
		args:   makeKV("k", "v"),
		expect: `pfx1/pfx2 "msg"="msg" "error"="err" "k"="v"`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			for _, n := range tc.names {
				sink = sink.WithName(n)
			}
			sink.Error(fmt.Errorf("err"), "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestInfoWithValues(t *testing.T) {
	testCases := []struct {
		name   string
		values []interface{}
		args   []interface{}
		expect string
	}{{
		name:   "zero",
		values: makeKV(),
		args:   makeKV("k", "v"),
		expect: ` "level"=0 "msg"="msg" "k"="v"`,
	}, {
		name:   "one",
		values: makeKV("one", 1),
		args:   makeKV("k", "v"),
		expect: ` "level"=0 "msg"="msg" "one"=1 "k"="v"`,
	}, {
		name:   "two",
		values: makeKV("one", 1, "two", 2),
		args:   makeKV("k", "v"),
		expect: ` "level"=0 "msg"="msg" "one"=1 "two"=2 "k"="v"`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			sink = sink.WithValues(tc.values...)
			sink.Info(0, "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestErrorWithValues(t *testing.T) {
	testCases := []struct {
		name   string
		values []interface{}
		args   []interface{}
		expect string
	}{{
		name:   "zero",
		values: makeKV(),
		args:   makeKV("k", "v"),
		expect: ` "msg"="msg" "error"="err" "k"="v"`,
	}, {
		name:   "one",
		values: makeKV("one", 1),
		args:   makeKV("k", "v"),
		expect: ` "msg"="msg" "error"="err" "one"=1 "k"="v"`,
	}, {
		name:   "two",
		values: makeKV("one", 1, "two", 2),
		args:   makeKV("k", "v"),
		expect: ` "msg"="msg" "error"="err" "one"=1 "two"=2 "k"="v"`,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cap := &capture{}
			sink := newSink(cap.Func, Options{})
			sink = sink.WithValues(tc.values...)
			sink.Error(fmt.Errorf("err"), "msg", tc.args...)
			if cap.log != tc.expect {
				t.Errorf("\nexpected %q\n     got %q", tc.expect, cap.log)
			}
		})
	}
}

func TestInfoWithCallDepth(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: All})
		dSink, _ := sink.(logr.CallDepthLogSink)
		sink = dSink.WithCallDepth(1)
		sink.Info(0, "msg")
		_, file, line, _ := runtime.Caller(1)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "level"=0 "msg"="msg"`, filepath.Base(file), line)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
}

func TestErrorWithCallDepth(t *testing.T) {
	t.Run("one", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, Options{LogCaller: All})
		dSink, _ := sink.(logr.CallDepthLogSink)
		sink = dSink.WithCallDepth(1)
		sink.Error(fmt.Errorf("err"), "msg")
		_, file, line, _ := runtime.Caller(1)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "msg"="msg" "error"="err"`, filepath.Base(file), line)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
}
