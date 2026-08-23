// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package drivers

import (
	"context"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"sync"
)

type (
	// Factory builds the driver for one provider from an opened connector. It
	// is written against a single credential family and adapted with
	// provider.Over, which is what lets an HTTP provider and a federating one
	// share this signature.
	Factory func(context.Context, *provider.Handle, *log.Logger) (Driver, error)

	// Registry maps a connector provider to the driver that reviews its
	// accounts.
	//
	// It lives here, next to the drivers, rather than on a provider's
	// Registration: an access review is a *consumer* of connectors, so what it
	// needs from one is its own business. A second consumer — an asset
	// inventory, say — brings its own registry and pkg/connector does not
	// change.
	//
	// Safe for concurrent use.
	Registry struct {
		mu        sync.RWMutex
		factories map[coredata.ConnectorProvider]Factory
	}
)

// NewRegistry returns an empty *Registry. Production code uses
// NewBuiltinRegistry; a test registers only the providers it exercises.
func NewRegistry() *Registry {
	return &Registry{factories: make(map[coredata.ConnectorProvider]Factory)}
}

// Register adds a factory for p, rejecting a nil factory or a duplicate so the
// caller can decide whether the condition is a programmer error worth crashing
// on.
func (r *Registry) Register(p coredata.ConnectorProvider, factory Factory) error {
	if p == "" {
		return fmt.Errorf("cannot register access review driver: missing provider")
	}

	if factory == nil {
		return fmt.Errorf("cannot register access review driver for %q: nil factory", p)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, dup := r.factories[p]; dup {
		return fmt.Errorf("cannot register access review driver for %q: duplicate registration", p)
	}

	r.factories[p] = factory

	return nil
}

// New builds the driver for an opened connector.
func (r *Registry) New(
	ctx context.Context,
	handle *provider.Handle,
	logger *log.Logger,
) (Driver, error) {
	p := handle.Connector.Provider

	r.mu.RLock()
	factory, ok := r.factories[p]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("cannot create access review driver: provider %q reviews no accounts", p)
	}

	return factory(ctx, handle, logger)
}

// Supports reports whether p has an access review driver at all. The console
// catalog gates on it, since a connector Probo can open but cannot review is of
// no use as an access source.
func (r *Registry) Supports(p coredata.ConnectorProvider) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.factories[p]

	return ok
}
