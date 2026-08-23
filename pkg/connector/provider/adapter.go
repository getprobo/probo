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

// A capability factory takes a *Handle so one registration entry serves every
// credential family. The two adapters below are what keep that uniform
// signature from forcing each factory to unwrap the credential itself: a
// factory names the family it is written against — connector.HTTPCredential or
// connector.CloudCredential — and the adapter narrows the Handle to it.
//
// They differ only in what they hand back, a capability or a verdict, and they
// are the only place that narrowing happens.
package provider

import (
	"context"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
)

// Over adapts a capability constructor written against one credential family
// to the family-agnostic factory a registry stores. G is the capability the
// consuming domain declared, which this package never names.
func Over[C connector.Credential, G any](
	fn func(context.Context, C, *Handle, *log.Logger) (G, error),
) func(context.Context, *Handle, *log.Logger) (G, error) {
	return func(ctx context.Context, h *Handle, logger *log.Logger) (G, error) {
		credential, err := credentialOf[C](h)
		if err != nil {
			var zero G

			return zero, err
		}

		return fn(ctx, credential, h, logger)
	}
}

// ProbeOver is Over for a connection check, which returns only a verdict.
func ProbeOver[C connector.Credential](
	fn func(context.Context, C, *Handle) error,
) func(context.Context, *Handle) error {
	return func(ctx context.Context, h *Handle) error {
		credential, err := credentialOf[C](h)
		if err != nil {
			return err
		}

		return fn(ctx, credential, h)
	}
}

// credentialOf narrows a Handle to the credential family a factory was written
// against. Reaching the error means a provider registered a factory for a
// family its protocol never mints, which Register cannot see because it cannot
// inspect a closure; naming the mismatch beats handing a zero value to a
// consumer.
func credentialOf[C connector.Credential](h *Handle) (C, error) {
	credential, ok := h.Credential.(C)
	if !ok {
		var zero C

		return zero, fmt.Errorf(
			"cannot use %s connector as %T: its %s credential is a %T",
			h.Connector.Provider,
			zero,
			h.Connector.Protocol,
			h.Credential,
		)
	}

	return credential, nil
}
