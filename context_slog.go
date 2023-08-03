//go:build go1.21
// +build go1.21

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

package logr

import (
	"context"
	"log/slog"
)

// This file contains the version of NewContext and FromContext which supports
// storing different types of loggers and converts as needed when retrieving
// the most recent one.

// FromContext returns a Logger from ctx or an error if no Logger is found.
func FromContext(ctx context.Context) (Logger, error) {
	l := ctx.Value(contextKey{})

	switch l := l.(type) {
	case Logger:
		return l, nil
	case *slog.Logger:
		return FromSlog(l), nil
	case slog.Handler:
		return FromSlogHandler(l), nil
	}

	return Logger{}, notFoundError{}
}

// FromContextOrDiscard returns a Logger from ctx.  If no Logger is found, this
// returns a Logger that discards all log messages.
func FromContextOrDiscard(ctx context.Context) Logger {
	l, err := FromContext(ctx)
	if err != nil {
		return Discard()
	}
	return l
}

// SlogFromContext is a variant of FromContext that returns a slog.Logger.
func SlogFromContext(ctx context.Context) (*slog.Logger, error) {
	l := ctx.Value(contextKey{})

	switch l := l.(type) {
	case Logger:
		return ToSlog(l), nil
	case *slog.Logger:
		return l, nil
	case slog.Handler:
		return slog.New(l), nil
	}

	return nil, notFoundError{}
}

// SlogFromContextOrDiscard is a variant of FromContextOrDiscard that returns a slog.Logger.
func SlogFromContextOrDiscard(ctx context.Context) *slog.Logger {
	l, err := SlogFromContext(ctx)
	if err != nil {
		return ToSlog(Discard()) // TODO: use something simpler
	}
	return l
}

// SlogHandlerFromContext is a variant of FromContext that returns a slog.Handler.
func SlogHandlerFromContext(ctx context.Context) (slog.Handler, error) {
	l := ctx.Value(contextKey{})

	switch l := l.(type) {
	case Logger:
		return ToSlogHandler(l), nil
	case *slog.Logger:
		return l.Handler(), nil
	case slog.Handler:
		return l, nil
	}

	return nil, notFoundError{}
}

// SlogHandlerFromContextOrDiscard is a variant of FromContextOrDiscard that returns a slog.Handler.
func SlogHandlerFromContextOrDiscard(ctx context.Context) slog.Handler {
	l, err := SlogHandlerFromContext(ctx)
	if err != nil {
		return ToSlog(Discard()).Handler() // TODO: use something simpler
	}
	return l
}

// NewContext returns a new Context, derived from ctx, which carries the
// provided Logger.
func NewContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// SlogNewContext returns a new Context, derived from ctx, which carries the
// provided slog.Logger.
func SlogNewContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

// SlogHandlerNewContext returns a new Context, derived from ctx, which carries the
// provided slog.Handler.
func SlogHandlerNewContext(ctx context.Context, handler slog.Handler) context.Context {
	return context.WithValue(ctx, contextKey{}, handler)
}
