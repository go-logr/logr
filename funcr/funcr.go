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

// Package funcr implements github.com/go-logr/logr.Logger in terms of
// an arbitrary "write" function.  This will not call String or
// Error methods on values.
package funcr

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
)

// New returns a logr.Logger which is implemented by an arbitrary function.
func New(fn func(prefix, args string), opts Options) logr.Logger {
	return logr.New(newSink(fn, opts))
}

// Underlier exposes access to the underlying logging function. Since
// callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an
// abstraction and more of a way to test type conversion.
type Underlier interface {
	GetUnderlying() func(prefix, args string)
}

func newSink(fn func(prefix, args string), opts Options) logr.LogSink {
	l := &fnlogger{
		prefix:       "",
		values:       nil,
		depth:        0,
		write:        fn,
		logCaller:    opts.LogCaller,
		logTimestamp: opts.LogTimestamp,
		verbosity:    opts.Verbosity,
		helper:       opts.Helper,
	}
	if l.helper == nil {
		// We have to have a valid function for GetCallStackHelper, so
		// we might as well just cover the nil case once here.
		l.helper = func() {}
	}
	return l
}

// Options carries parameters which influence the way logs are generated.
type Options struct {
	// LogCaller tells funcr to add a "caller" key to some or all log lines.
	// This has some overhead, so some users might not want it.
	LogCaller MessageClass

	// LogTimestamps tells funcr to add a "ts" key to log lines.  This has some
	// overhead, so some users might not want it.
	LogTimestamp bool

	// Verbosity tells funcr which V logs to be write.  Higher values enable
	// more logs.
	Verbosity int

	// Helper is an optional function that funcr will call to mark its own
	// stack frames as helper functions.
	Helper func()
}

// MessageClass indicates which category or categories of messages to consider.
type MessageClass int

const (
	None MessageClass = iota
	All
	Info
	Error
)

const timestampFmt = "2006-01-02 15:04:05.000000"

type fnlogger struct {
	prefix       string
	values       []interface{}
	depth        int
	write        func(prefix, args string)
	helper       func()
	logCaller    MessageClass
	logTimestamp bool
	verbosity    int
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &fnlogger{}
var _ logr.CallDepthLogSink = &fnlogger{}
var _ Underlier = &fnlogger{}
var _ logr.CallStackHelperLogSink = &fnlogger{}

func flatten(kvList ...interface{}) string {
	if len(kvList)%2 != 0 {
		kvList = append(kvList, "<no-value>")
	}
	// Empirically bytes.Buffer is faster than strings.Builder for this.
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			k = fmt.Sprintf("<non-string-key-%d>", i/2)
		}
		v := kvList[i+1]

		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteRune('"')
		buf.WriteString(k)
		buf.WriteRune('"')
		buf.WriteRune('=')
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

func pretty(value interface{}) string {
	return prettyWithFlags(value, 0)
}

const (
	flagRawString = 0x1
)

// TODO: This is not fast. Most of the overhead goes here.
func prettyWithFlags(value interface{}, flags uint32) string {
	// Handling the most common types without reflect is a small perf win.
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case string:
		if flags&flagRawString > 0 {
			return v
		}
		// This is empirically faster than strings.Builder.
		return `"` + v + `"`
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uintptr:
		return strconv.FormatUint(uint64(v), 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	t := reflect.TypeOf(value)
	if t == nil {
		return "null"
	}
	v := reflect.ValueOf(value)
	switch t.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.String:
		if flags&flagRawString > 0 {
			return v.String()
		}
		// This is empirically faster than strings.Builder.
		return `"` + v.String() + `"`
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(int64(v.Int()), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(uint64(v.Uint()), 10)
	case reflect.Float32:
		return strconv.FormatFloat(float64(v.Float()), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Struct:
		buf.WriteRune('{')
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				// reflect says this field is only defined for non-exported fields.
				continue
			}
			if i > 0 {
				buf.WriteRune(',')
			}
			buf.WriteRune('"')
			name := f.Name
			if tag, found := f.Tag.Lookup("json"); found {
				if comma := strings.Index(tag, ","); comma != -1 {
					name = tag[:comma]
				} else {
					name = tag
				}
			}
			buf.WriteString(name)
			buf.WriteRune('"')
			buf.WriteRune(':')
			buf.WriteString(pretty(v.Field(i).Interface()))
		}
		buf.WriteRune('}')
		return buf.String()
	case reflect.Slice, reflect.Array:
		buf.WriteRune('[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				buf.WriteRune(',')
			}
			e := v.Index(i)
			buf.WriteString(pretty(e.Interface()))
		}
		buf.WriteRune(']')
		return buf.String()
	case reflect.Map:
		buf.WriteRune('{')
		// This does not sort the map keys, for best perf.
		it := v.MapRange()
		i := 0
		for it.Next() {
			if i > 0 {
				buf.WriteRune(',')
			}
			// JSON only does string keys.
			buf.WriteRune('"')
			buf.WriteString(prettyWithFlags(it.Key().Interface(), flagRawString))
			buf.WriteRune('"')
			buf.WriteRune(':')
			buf.WriteString(pretty(it.Value().Interface()))
			i++
		}
		buf.WriteRune('}')
		return buf.String()
	case reflect.Ptr, reflect.Interface:
		return pretty(v.Elem().Interface())
	}
	return fmt.Sprintf(`"<unhandled-%s>"`, t.Kind().String())
}

type callerID struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

func (l fnlogger) caller() callerID {
	// +1 for this frame, +1 for Info/Error.
	_, file, line, ok := runtime.Caller(l.depth + 2)
	if !ok {
		return callerID{"<unknown>", 0}
	}
	return callerID{filepath.Base(file), line}
}

// Note that this receiver is a pointer, so depth can be saved.
func (l *fnlogger) Init(info logr.RuntimeInfo) {
	l.depth += info.CallDepth
}

func (l fnlogger) Enabled(level int) bool {
	return level <= l.verbosity
}

func (l fnlogger) Info(level int, msg string, kvList ...interface{}) {
	args := make([]interface{}, 0, 64) // using a constant here impacts perf
	if l.logTimestamp {
		args = append(args, "ts", time.Now().Format(timestampFmt))
	}
	if l.logCaller == All || l.logCaller == Info {
		args = append(args, "caller", l.caller())
	}
	args = append(args, "level", level, "msg", msg)
	args = append(args, l.values...)
	args = append(args, kvList...)
	argsStr := flatten(args...)
	l.helper()
	l.write(l.prefix, argsStr)
}

func (l fnlogger) Error(err error, msg string, kvList ...interface{}) {
	args := make([]interface{}, 0, 64) // using a constant here impacts perf
	if l.logTimestamp {
		args = append(args, "ts", time.Now().Format(timestampFmt))
	}
	if l.logCaller == All || l.logCaller == Error {
		args = append(args, "caller", l.caller())
	}
	args = append(args, "msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	args = append(args, "error", loggableErr)
	args = append(args, l.values...)
	args = append(args, kvList...)
	argsStr := flatten(args...)
	l.helper()
	l.write(l.prefix, argsStr)
}

// WithName returns a new Logger with the specified name appended.  funcr
// uses '/' characters to separate name elements.  Callers should not pass '/'
// in the provided name string, but this library does not actually enforce that.
func (l fnlogger) WithName(name string) logr.LogSink {
	if len(l.prefix) > 0 {
		l.prefix += "/"
	}
	l.prefix += name
	return &l
}

func (l fnlogger) WithValues(kvList ...interface{}) logr.LogSink {
	// Three slice args forces a copy.
	n := len(l.values)
	l.values = append(l.values[:n:n], kvList...)
	return &l
}

func (l fnlogger) WithCallDepth(depth int) logr.LogSink {
	l.depth += depth
	return &l
}

func (l fnlogger) GetUnderlying() func(prefix, args string) {
	return l.write
}

func (l fnlogger) GetCallStackHelper() func() {
	return l.helper
}
