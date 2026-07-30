package mcp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
)

// RecoverFunc is called when a tool handler panics. It receives the recovered
// value (whatever was passed to panic) and returns an error to be reported to
// the client.
//
// This matches the signature and semantics of gqlgen's RecoverFunc.
//
// Example:
//
//	server.New(resolver, server.WithRecoverFunc(func(ctx context.Context, err any) error {
//	    log.Error("tool panic", "err", err)
//	    return errors.New("internal server error")
//	}))
type RecoverFunc func(ctx context.Context, err any) error

// DefaultRecoverFunc prints the panic and stack trace to stderr and returns a
// generic internal error. This matches gqlgen's DefaultRecover behavior.
func DefaultRecoverFunc(_ context.Context, err any) error {
	fmt.Fprintln(os.Stderr, err)
	fmt.Fprintln(os.Stderr)
	debug.PrintStack()
	return errors.New("internal system error")
}

// Option configures the generated MCP server.
type Option func(*Options)

// Options holds configuration for the generated MCP server.
type Options struct {
	RecoverFunc RecoverFunc
}

// WithRecoverFunc sets the panic recover function for tool handlers.
// The recover function is called when a tool handler panics, and its return
// value is sent to the client in place of the panic.
func WithRecoverFunc(fn RecoverFunc) Option {
	return func(o *Options) {
		o.RecoverFunc = fn
	}
}

// ApplyOptions applies the given options to an Options struct.
// If RecoverFunc is nil after applying options, it is set to DefaultRecoverFunc:
// recovery is always enabled, matching gqlgen's behavior.
func ApplyOptions(opts []Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	if o.RecoverFunc == nil {
		o.RecoverFunc = DefaultRecoverFunc
	}
	return o
}
