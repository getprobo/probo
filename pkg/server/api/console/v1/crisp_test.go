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

package console_v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// resolveAPIKeyConnectorCredential is the sole gate keeping normal API-key
// providers requiring a customer key while ManagedAPIKey providers (Model B,
// e.g. Crisp) persist none. A regression either drops the required-key check for
// every provider or persists the managed key on the row.
func TestResolveAPIKeyConnectorCredential(t *testing.T) {
	t.Parallel()

	// NewBuiltinRegistry registers Crisp as a ManagedAPIKey provider; the
	// managed key must then be set for it to count as configured.
	configuredReg := provider.NewBuiltinRegistry()
	configuredReg.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")
	configured := &Resolver{providerRegistry: configuredReg}

	unconfigured := &Resolver{providerRegistry: provider.NewBuiltinRegistry()}

	clientKey := "customer-key"
	empty := ""

	t.Run("managed and configured persists no key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, nil)
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})

	t.Run("managed and configured ignores a client-supplied key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, &clientKey)
		require.NoError(t, err)
		assert.Equal(t, "", key)
	})

	t.Run("managed but unconfigured is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := unconfigured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderCrisp, nil)
		require.Error(t, err)
	})

	// Sentry declares no key shape, so it is what exercises the pass-through
	// branch: anything non-empty is the customer's to get wrong. Giving Sentry
	// a shape one day fails these two, which is where to look.
	t.Run("non-managed requires a key", func(t *testing.T) {
		t.Parallel()

		_, errNil := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderSentry, nil)
		require.EqualError(t, errNil, "apiKey is required")

		_, errEmpty := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderSentry, &empty)
		require.EqualError(t, errEmpty, "apiKey is required")
	})

	t.Run("non-managed returns the client key", func(t *testing.T) {
		t.Parallel()

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderSentry, &clientKey)
		require.NoError(t, err)
		assert.Equal(t, "customer-key", key)
	})

	// A key is copied out of a provider's UI, so it arrives with whatever the
	// clipboard picked up. The trim is what catches that, and only the trim:
	// a shape pattern cannot, because [^:] matches a newline and Go's $ is
	// end-of-text, so "pk-lf-a:sk-lf-b\n" satisfies Langfuse's own pattern.
	t.Run("surrounding whitespace is trimmed off the key", func(t *testing.T) {
		t.Parallel()

		pasted := " pk-lf-1111:sk-lf-2222\n"

		key, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderLangfuse, &pasted)
		require.NoError(t, err)
		assert.Equal(t, "pk-lf-1111:sk-lf-2222", key)
	})

	t.Run("a key of only whitespace is no key", func(t *testing.T) {
		t.Parallel()

		blank := "   "

		_, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderLangfuse, &blank)
		require.EqualError(t, err, "apiKey is required")
	})

	// The trim runs before the shape check, which is what lets a pasted key
	// with a trailing newline reach the provider at all.
	t.Run("a key of the wrong shape is refused", func(t *testing.T) {
		t.Parallel()

		half := " pk-lf-1111\n"

		_, err := configured.resolveAPIKeyConnectorCredential(coredata.ConnectorProviderLangfuse, &half)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pk-lf-…:sk-lf-…")
	})
}
