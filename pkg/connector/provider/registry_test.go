// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package provider_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestEveryProviderRegistered asserts that every
// coredata.ConnectorProvider constant has a matching Registration in
// the registry, that the registration carries the minimum metadata
// (Provider, DisplayName), and that the access-review NewDriver
// closure is wired — so the provider can actually drive a review.
func TestEveryProviderRegistered(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	for _, p := range coredata.ConnectorProviders() {
		t.Run(string(p), func(t *testing.T) {
			t.Parallel()

			reg, ok := r.Get(p)
			require.Truef(t, ok, "provider %q has no Registration", p)
			require.NotNil(t, reg, "provider %q Registration is nil", p)
			require.Equalf(t, p, reg.Provider, "provider %q has mismatching Registration.Provider", p)
			assert.NotEmptyf(t, reg.DisplayName, "provider %q has empty DisplayName", p)
			assert.NotNilf(t, reg.NewDriver, "provider %q has nil NewDriver", p)
		})
	}
}

// TestRegistry_Register exercises the validation and duplicate-detection
// paths on Register. Programmer errors at NewBuiltinRegistry time —
// nil, empty Provider, empty DisplayName, duplicate — must all surface
// as errors rather than silently registering a malformed entry.
func TestRegistry_Register(t *testing.T) {
	t.Parallel()

	t.Run("nil Registration", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil Registration")
	})

	t.Run("empty Provider", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{DisplayName: "X"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing Provider")
	})

	t.Run("empty DisplayName", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{Provider: coredata.ConnectorProviderSlack})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing DisplayName")
	})

	t.Run("duplicate registration", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		require.NoError(t, r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
		}))
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack-bis",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate registration")
	})

	t.Run("valid Registration round-trips through Get", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		want := &provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
		}
		require.NoError(t, r.Register(want))

		got, ok := r.Get(coredata.ConnectorProviderSlack)
		require.True(t, ok)
		assert.Same(t, want, got)
	})
}

// TestRegistry_All asserts the registry returns the same number of
// entries that have been registered. The builtin registry is the
// canonical source of truth: every coredata.ConnectorProvider has
// exactly one matching Registration, no more.
func TestRegistry_All(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	assert.Len(t, r.All(), len(coredata.ConnectorProviders()))
}

// TestRegistry_ProviderDisplayName covers the fallback path: an
// unregistered provider returns its raw constant string.
func TestRegistry_ProviderDisplayName(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	assert.Equal(t, "Slack", r.ProviderDisplayName(coredata.ConnectorProviderSlack))
	assert.Equal(t, "UNKNOWN", r.ProviderDisplayName(coredata.ConnectorProvider("UNKNOWN")))
}

// TestRegistry_ProviderOAuth2Scopes covers the nil path for an
// unregistered provider and the populated path for a registered one.
func TestRegistry_ProviderOAuth2Scopes(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	assert.NotEmpty(t, r.ProviderOAuth2Scopes(coredata.ConnectorProviderSlack))
	assert.Nil(t, r.ProviderOAuth2Scopes(coredata.ConnectorProvider("UNKNOWN")))
}

// TestRegistry_ProbeURL covers the registered and unregistered paths.
// Slack ships a probe URL in its Registration; an unknown provider
// returns the empty string.
func TestRegistry_ProbeURL(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	assert.NotEmpty(t, r.ProbeURL("SLACK"))
	assert.Empty(t, r.ProbeURL("UNKNOWN"))
}
