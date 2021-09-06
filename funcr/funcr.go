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

// Package funcr implements formatting of structured log messages and
// optionally captures the call site. This will not call String or
// Error methods on values.
//
// The simplest way to use it is via its implementation of a
// github.com/go-logr/logr.LogSink with output through an arbitrary
// "write" function. Alternatively, funcr can also be embedded inside
// a custom LogSink implementation. This is useful when the LogSink
// needs to implement additional methods.
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
		Formatter: NewFormatter(opts),
		write:     fn,
	}
	// For skipping fnlogger.Info and fnlogger.Error.
	l.Formatter.AddCallDepth(1)
	return l
}

// Options carries parameters which influence the way logs are generated.
type Options struct {
	// LogCaller tells funcr to add a "caller" key to some or all log lines.
	// This has some overhead, so some users might not want it.
	LogCaller MessageClass

	// LogTimestamp tells funcr to add a "ts" key to log lines.  This has some
	// overhead, so some users might not want it.
	LogTimestamp bool

	// Verbosity tells funcr which V logs to be write.  Higher values enable
	// more logs.
	Verbosity int
}

// MessageClass indicates which category or categories of messages to consider.
type MessageClass int

const (
	// None ignores all message classes.
	None MessageClass = iota
	// All considers all message classes.
	All
	// Info only considers info messages.
	Info
	// Error only considers error messages.
	Error
)

const timestampFmt = "2006-01-02 15:04:05.000000"

// fnlogger inherits some of its LogSink implementation from Formatter
// and just needs to add some glue code.
type fnlogger struct {
	Formatter
	write func(prefix, args string)
}

func (l fnlogger) WithName(name string) logr.LogSink {
	l.Formatter.AddName(name)
	return &l
}

func (l fnlogger) WithValues(kvList ...interface{}) logr.LogSink {
	l.Formatter.AddValues(kvList)
	return &l
}

func (l fnlogger) WithCallDepth(depth int) logr.LogSink {
	l.Formatter.AddCallDepth(depth)
	return &l
}

func (l fnlogger) Info(level int, msg string, kvList ...interface{}) {
	prefix, args := l.FormatInfo(level, msg, kvList)
	l.write(prefix, args)
}

func (l fnlogger) Error(err error, msg string, kvList ...interface{}) {
	prefix, args := l.FormatError(err, msg, kvList)
	l.write(prefix, args)
}

func (l fnlogger) GetUnderlying() func(prefix, args string) {
	return l.write
}

// Assert conformance to the interfaces.
var _ logr.LogSink = &fnlogger{}
var _ logr.CallDepthLogSink = &fnlogger{}
var _ Underlier = &fnlogger{}

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
		if v.IsNil() {
			return "null"
		}
		return pretty(v.Elem().Interface())
	}
	return fmt.Sprintf(`"<unhandled-%s>"`, t.Kind().String())
}

type callerID struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// NewFormatter constructs a Formatter.
func NewFormatter(opts Options) Formatter {
	f := Formatter{
		prefix:       "",
		values:       nil,
		depth:        0,
		logCaller:    opts.LogCaller,
		logTimestamp: opts.LogTimestamp,
		verbosity:    opts.Verbosity,
	}
	return f
}

// Formatter is an opaque struct which can be embedded in a LogSink
// implementation. It should be constructed with NewFormatter. Some of
// its methods directly implement logr.LogSink.
type Formatter struct {
	prefix       string
	values       []interface{}
	depth        int
	logCaller    MessageClass
	logTimestamp bool
	verbosity    int
}

func (f Formatter) caller() callerID {
	// +1 for this frame, +1 for Info/Error.
	_, file, line, ok := runtime.Caller(f.depth + 2)
	if !ok {
		return callerID{"<unknown>", 0}
	}
	return callerID{filepath.Base(file), line}
}

// Init configures this Formatter from runtime info, such as the call depth
// imposed by logr itself.
// Note that this receiver is a pointer, so depth can be saved.
func (f *Formatter) Init(info logr.RuntimeInfo) {
	f.depth += info.CallDepth
}

// Enabled checks whether an info message at the given level should be logged.
func (f Formatter) Enabled(level int) bool {
	return level <= f.verbosity
}

// GetDepth returns the current depth of this Formatter.  This is useful for
// implementations which do their own caller attribution.
func (f Formatter) GetDepth() int {
	return f.depth
}

// FormatInfo flattens an Info log message into strings.
// The prefix will be empty when no names were set.
func (f Formatter) FormatInfo(level int, msg string, kvList []interface{}) (prefix, argsStr string) {
	args := make([]interface{}, 0, 64) // using a constant here impacts perf
	if f.logTimestamp {
		args = append(args, "ts", time.Now().Format(timestampFmt))
	}
	if f.logCaller == All || f.logCaller == Info {
		args = append(args, "caller", f.caller())
	}
	args = append(args, "level", level, "msg", msg)
	args = append(args, f.values...)
	args = append(args, kvList...)
	return f.prefix, flatten(args...)
}

// FormatError flattens an Error log message into strings.
// The prefix will be empty when no names were set.
func (f Formatter) FormatError(err error, msg string, kvList []interface{}) (prefix, argsStr string) {
	args := make([]interface{}, 0, 64) // using a constant here impacts perf
	if f.logTimestamp {
		args = append(args, "ts", time.Now().Format(timestampFmt))
	}
	if f.logCaller == All || f.logCaller == Error {
		args = append(args, "caller", f.caller())
	}
	args = append(args, "msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	args = append(args, "error", loggableErr)
	args = append(args, f.values...)
	args = append(args, kvList...)
	return f.prefix, flatten(args...)
}

// AddName appends the specified name.  funcr uses '/' characters to separate
// name elements.  Callers should not pass '/' in the provided name string, but
// this library does not actually enforce that.
func (f *Formatter) AddName(name string) {
	if len(f.prefix) > 0 {
		f.prefix += "/"
	}
	f.prefix += name
}

// AddValues adds key-value pairs to the set of saved values to be logged with
// each log line.
func (f *Formatter) AddValues(kvList []interface{}) {
	// Three slice args forces a copy.
	n := len(f.values)
	f.values = append(f.values[:n:n], kvList...)
}

// AddCallDepth increases the number of stack-frames to skip when attributing
// the log line to a file and line.
func (f *Formatter) AddCallDepth(depth int) {
	f.depth += depth
}
