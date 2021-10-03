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

// Logging this should result in the MarshalLog() value.
type Tmarshaler string

func (t Tmarshaler) MarshalLog() interface{} {
	return struct{ Inner string }{"I am a logr.Marshaler"}
}

func (t Tmarshaler) String() string {
	return "String(): you should not see this"
}

func (t Tmarshaler) Error() string {
	return "Error(): you should not see this"
}

// Logging this should result in the String() value.
type Tstringer string

func (t Tstringer) String() string {
	return "I am a fmt.Stringer"
}

func (t Tstringer) Error() string {
	return "Error(): you should not see this"
}

type TjsontagsString struct {
	String1 string `json:"string1"`           // renamed
	String2 string `json:"-"`                 // ignored
	String3 string `json:"-,"`                // named "-"
	String4 string `json:"string4,omitempty"` // renamed, ignore if empty
	String5 string `json:","`                 // no-op
	String6 string `json:",omitempty"`        // ignore if empty
}

type TjsontagsBool struct {
	Bool1 bool `json:"bool1"`           // renamed
	Bool2 bool `json:"-"`               // ignored
	Bool3 bool `json:"-,"`              // named "-"
	Bool4 bool `json:"bool4,omitempty"` // renamed, ignore if empty
	Bool5 bool `json:","`               // no-op
	Bool6 bool `json:",omitempty"`      // ignore if empty
}

type TjsontagsInt struct {
	Int1 int `json:"int1"`           // renamed
	Int2 int `json:"-"`              // ignored
	Int3 int `json:"-,"`             // named "-"
	Int4 int `json:"int4,omitempty"` // renamed, ignore if empty
	Int5 int `json:","`              // no-op
	Int6 int `json:",omitempty"`     // ignore if empty
}

type TjsontagsUint struct {
	Uint1 uint `json:"uint1"`           // renamed
	Uint2 uint `json:"-"`               // ignored
	Uint3 uint `json:"-,"`              // named "-"
	Uint4 uint `json:"uint4,omitempty"` // renamed, ignore if empty
	Uint5 uint `json:","`               // no-op
	Uint6 uint `json:",omitempty"`      // ignore if empty
}

type TjsontagsFloat struct {
	Float1 float64 `json:"float1"`           // renamed
	Float2 float64 `json:"-"`                // ignored
	Float3 float64 `json:"-,"`               // named "-"
	Float4 float64 `json:"float4,omitempty"` // renamed, ignore if empty
	Float5 float64 `json:","`                // no-op
	Float6 float64 `json:",omitempty"`       // ignore if empty
}

type TjsontagsComplex struct {
	Complex1 complex128 `json:"complex1"`           // renamed
	Complex2 complex128 `json:"-"`                  // ignored
	Complex3 complex128 `json:"-,"`                 // named "-"
	Complex4 complex128 `json:"complex4,omitempty"` // renamed, ignore if empty
	Complex5 complex128 `json:","`                  // no-op
	Complex6 complex128 `json:",omitempty"`         // ignore if empty
}

type TjsontagsPtr struct {
	Ptr1 *string `json:"ptr1"`           // renamed
	Ptr2 *string `json:"-"`              // ignored
	Ptr3 *string `json:"-,"`             // named "-"
	Ptr4 *string `json:"ptr4,omitempty"` // renamed, ignore if empty
	Ptr5 *string `json:","`              // no-op
	Ptr6 *string `json:",omitempty"`     // ignore if empty
}

type TjsontagsArray struct {
	Array1 [2]string `json:"array1"`           // renamed
	Array2 [2]string `json:"-"`                // ignored
	Array3 [2]string `json:"-,"`               // named "-"
	Array4 [2]string `json:"array4,omitempty"` // renamed, ignore if empty
	Array5 [2]string `json:","`                // no-op
	Array6 [2]string `json:",omitempty"`       // ignore if empty
}

type TjsontagsSlice struct {
	Slice1 []string `json:"slice1"`           // renamed
	Slice2 []string `json:"-"`                // ignored
	Slice3 []string `json:"-,"`               // named "-"
	Slice4 []string `json:"slice4,omitempty"` // renamed, ignore if empty
	Slice5 []string `json:","`                // no-op
	Slice6 []string `json:",omitempty"`       // ignore if empty
}

type TjsontagsMap struct {
	Map1 map[string]string `json:"map1"`           // renamed
	Map2 map[string]string `json:"-"`              // ignored
	Map3 map[string]string `json:"-,"`             // named "-"
	Map4 map[string]string `json:"map4,omitempty"` // renamed, ignore if empty
	Map5 map[string]string `json:","`              // no-op
	Map6 map[string]string `json:",omitempty"`     // ignore if empty
}

type Tinnerstruct struct {
	Inner string
}
type Tinnerint int
type Tinnermap map[string]string
type Tinnerslice []string

type Tembedstruct struct {
	Tinnerstruct
	Outer string
}

type Tembednonstruct struct {
	Tinnerint
	Tinnermap
	Tinnerslice
}

type Tinner1 Tinnerstruct
type Tinner2 Tinnerstruct
type Tinner3 Tinnerstruct
type Tinner4 Tinnerstruct
type Tinner5 Tinnerstruct
type Tinner6 Tinnerstruct

type Tembedjsontags struct {
	Outer   string
	Tinner1 `json:"inner1"`
	Tinner2 `json:"-"`
	Tinner3 `json:"-,"`
	Tinner4 `json:"inner4,omitempty"`
	Tinner5 `json:","`
	Tinner6 `json:"inner6,omitempty"`
}

func TestPretty(t *testing.T) {
	// used below
	newStr := func(s string) *string {
		return &s
	}

	cases := []struct {
		val interface{}
		exp string // used in cases where JSON can't handle it
	}{
		{val: "strval"},
		{val: substr("substrval")},
		{val: true},
		{val: false},
		{val: int(93)},
		{val: int8(93)},
		{val: int16(93)},
		{val: int32(93)},
		{val: int64(93)},
		{val: int(-93)},
		{val: int8(-93)},
		{val: int16(-93)},
		{val: int32(-93)},
		{val: int64(-93)},
		{val: uint(93)},
		{val: uint8(93)},
		{val: uint16(93)},
		{val: uint32(93)},
		{val: uint64(93)},
		{val: uintptr(93)},
		{val: float32(93.76)},
		{val: float64(93.76)},
		{
			val: complex64(93i),
			exp: `"(0+93i)"`,
		},
		{
			val: complex128(93i),
			exp: `"(0+93i)"`,
		},
		{val: ptrint(93)},
		{val: ptrstr("pstrval")},
		{val: []int{}},
		{
			val: []int(nil),
			exp: `[]`,
		},
		{val: []int{9, 3, 7, 6}},
		{val: [4]int{9, 3, 7, 6}},
		{
			val: struct {
				Int         int
				notExported string
				String      string
			}{
				93, "you should not see this", "seventy-six",
			},
		},
		{val: map[string]int{}},
		{
			val: map[string]int(nil),
			exp: `{}`,
		},
		{
			val: map[string]int{
				"nine": 3,
			},
		},
		{
			val: map[substr]int{
				"nine": 3,
			},
		},
		{
			val: struct {
				X int `json:"x"`
				Y int `json:"y"`
			}{
				93, 76,
			},
		},
		{
			val: struct {
				X []int
				Y map[int]int
				Z struct{ P, Q int }
			}{
				[]int{9, 3, 7, 6},
				map[int]int{9: 3},
				struct{ P, Q int }{9, 3},
			},
		},
		{
			val: []struct{ X, Y string }{
				{"nine", "three"},
				{"seven", "six"},
			},
		},
		{
			val: struct {
				A *int
				B *int
				C interface{}
				D interface{}
			}{
				B: ptrint(1),
				D: interface{}(2),
			},
		},
		{
			val: Tmarshaler("foobar"),
			exp: `{"Inner":"I am a logr.Marshaler"}`,
		},
		{
			val: Tstringer("foobar"),
			exp: `"I am a fmt.Stringer"`,
		},
		{
			val: fmt.Errorf("error"),
			exp: `"error"`,
		},
		{
			val: TjsontagsString{
				String1: "v1",
				String2: "v2",
				String3: "v3",
				String4: "v4",
				String5: "v5",
				String6: "v6",
			},
		},
		{val: TjsontagsString{}},
		{
			val: TjsontagsBool{
				Bool1: true,
				Bool2: true,
				Bool3: true,
				Bool4: true,
				Bool5: true,
				Bool6: true,
			},
		},
		{val: TjsontagsBool{}},
		{
			val: TjsontagsInt{
				Int1: 1,
				Int2: 2,
				Int3: 3,
				Int4: 4,
				Int5: 5,
				Int6: 6,
			},
		},
		{val: TjsontagsInt{}},
		{
			val: TjsontagsUint{
				Uint1: 1,
				Uint2: 2,
				Uint3: 3,
				Uint4: 4,
				Uint5: 5,
				Uint6: 6,
			},
		},
		{val: TjsontagsUint{}},
		{
			val: TjsontagsFloat{
				Float1: 1.1,
				Float2: 2.2,
				Float3: 3.3,
				Float4: 4.4,
				Float5: 5.5,
				Float6: 6.6,
			},
		},
		{val: TjsontagsFloat{}},
		{
			val: TjsontagsComplex{
				Complex1: 1i,
				Complex2: 2i,
				Complex3: 3i,
				Complex4: 4i,
				Complex5: 5i,
				Complex6: 6i,
			},
			exp: `{"complex1":"(0+1i)","-":"(0+3i)","complex4":"(0+4i)","Complex5":"(0+5i)","Complex6":"(0+6i)"}`,
		},
		{
			val: TjsontagsComplex{},
			exp: `{"complex1":"(0+0i)","-":"(0+0i)","Complex5":"(0+0i)"}`,
		},
		{
			val: TjsontagsPtr{
				Ptr1: newStr("1"),
				Ptr2: newStr("2"),
				Ptr3: newStr("3"),
				Ptr4: newStr("4"),
				Ptr5: newStr("5"),
				Ptr6: newStr("6"),
			},
		},
		{val: TjsontagsPtr{}},
		{
			val: TjsontagsArray{
				Array1: [2]string{"v1", "v1"},
				Array2: [2]string{"v2", "v2"},
				Array3: [2]string{"v3", "v3"},
				Array4: [2]string{"v4", "v4"},
				Array5: [2]string{"v5", "v5"},
				Array6: [2]string{"v6", "v6"},
			},
		},
		{val: TjsontagsArray{}},
		{
			val: TjsontagsSlice{
				Slice1: []string{"v1", "v1"},
				Slice2: []string{"v2", "v2"},
				Slice3: []string{"v3", "v3"},
				Slice4: []string{"v4", "v4"},
				Slice5: []string{"v5", "v5"},
				Slice6: []string{"v6", "v6"},
			},
		},
		{
			val: TjsontagsSlice{},
			exp: `{"slice1":[],"-":[],"Slice5":[]}`,
		},
		{
			val: TjsontagsMap{
				Map1: map[string]string{"k1": "v1"},
				Map2: map[string]string{"k2": "v2"},
				Map3: map[string]string{"k3": "v3"},
				Map4: map[string]string{"k4": "v4"},
				Map5: map[string]string{"k5": "v5"},
				Map6: map[string]string{"k6": "v6"},
			},
		},
		{
			val: TjsontagsMap{},
			exp: `{"map1":{},"-":{},"Map5":{}}`,
		},
		{val: Tembedstruct{}},
		{
			val: Tembednonstruct{},
			exp: `{"Tinnerint":0,"Tinnermap":{},"Tinnerslice":[]}`,
		},
		{val: Tembedjsontags{}},
	}

	f := NewFormatter(Options{})
	for i, tc := range cases {
		ours := f.pretty(tc.val)
		want := ""
		if tc.exp != "" {
			want = tc.exp
		} else {
			jb, err := json.Marshal(tc.val)
			if err != nil {
				t.Errorf("[%d]: unexpected error: %v", i, err)
			}
			want = string(jb)
		}
		if ours != want {
			t.Errorf("[%d]:\n\texpected %q\n\tgot      %q", i, want, ours)
		}
	}
}

func makeKV(args ...interface{}) []interface{} {
	return args
}

func TestRender(t *testing.T) {
	testCases := []struct {
		name       string
		kv         []interface{}
		expectKV   string
		expectJSON string
	}{{
		name:       "nil",
		kv:         nil,
		expectKV:   "",
		expectJSON: "{}",
	}, {
		name:       "empty",
		kv:         []interface{}{},
		expectKV:   "",
		expectJSON: "{}",
	}, {
		name:       "primitives",
		kv:         makeKV("int", 1, "str", "ABC", "bool", true),
		expectKV:   `"int"=1 "str"="ABC" "bool"=true`,
		expectJSON: `{"int":1,"str":"ABC","bool":true}`,
	}, {
		name:       "missing value",
		kv:         makeKV("key"),
		expectKV:   `"key"="<no-value>"`,
		expectJSON: `{"key":"<no-value>"}`,
	}, {
		name:       "non-string key",
		kv:         makeKV(123, "val"),
		expectKV:   `"<non-string-key>"="val"`,
		expectJSON: `{"<non-string-key>":"val"}`,
	}}

	fKV := NewFormatter(Options{})
	fJSON := NewFormatterJSON(Options{})
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("KV", func(t *testing.T) {
				r := fKV.render(tc.kv, nil)
				if r != tc.expectKV {
					t.Errorf("wrong KV output:\nexpected %q\n     got %q", tc.expectKV, r)
				}
			})
			t.Run("JSON", func(t *testing.T) {
				r := fJSON.render(tc.kv, nil)
				if r != tc.expectJSON {
					t.Errorf("wrong JSON output:\nexpected %q\n     got %q", tc.expectJSON, r)
				}
			})
		})
	}
}

func TestEnabled(t *testing.T) {
	t.Run("default V", func(t *testing.T) {
		log := newSink(func(prefix, args string) {}, NewFormatter(Options{}))
		if !log.Enabled(0) {
			t.Errorf("expected true")
		}
		if log.Enabled(1) {
			t.Errorf("expected false")
		}
	})
	t.Run("V=9", func(t *testing.T) {
		log := newSink(func(prefix, args string) {}, NewFormatter(Options{Verbosity: 9}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: All}))
		sink.Info(0, "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "level"=0 "msg"="msg"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Info", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: Info}))
		sink.Info(0, "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "level"=0 "msg"="msg"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Error", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: Error}))
		sink.Info(0, "msg")
		expect := ` "level"=0 "msg"="msg"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=None", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: None}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: All}))
		sink.Error(fmt.Errorf("err"), "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "msg"="msg" "error"="err"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Error", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: Error}))
		sink.Error(fmt.Errorf("err"), "msg")
		_, file, line, _ := runtime.Caller(0)
		expect := fmt.Sprintf(` "caller"={"file":%q,"line":%d} "msg"="msg" "error"="err"`, filepath.Base(file), line-1)
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=Info", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: Info}))
		sink.Error(fmt.Errorf("err"), "msg")
		expect := ` "msg"="msg" "error"="err"`
		if cap.log != expect {
			t.Errorf("\nexpected %q\n     got %q", expect, cap.log)
		}
	})
	t.Run("LogCaller=None", func(t *testing.T) {
		cap := &capture{}
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: None}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
			sink := newSink(cap.Func, NewFormatter(Options{}))
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
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: All}))
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
		sink := newSink(cap.Func, NewFormatter(Options{LogCaller: All}))
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
