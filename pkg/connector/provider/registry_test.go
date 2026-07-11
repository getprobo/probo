// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/connector"
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

	t.Run("APIKeyBasicAuth and APIKeyHeader mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:        coredata.ConnectorProviderSlack,
			DisplayName:     "Slack",
			APIKeyBasicAuth: true,
			APIKeyHeader:    "x-api-key",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("APIKeyAuthScheme and APIKeyHeader mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:         coredata.ConnectorProviderSlack,
			DisplayName:      "Slack",
			APIKeyAuthScheme: "SSWS",
			APIKeyHeader:     "x-api-key",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("APIKeyBasicAuthUserPass and APIKeyHeader mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:                coredata.ConnectorProviderSlack,
			DisplayName:             "Slack",
			APIKeyBasicAuthUserPass: true,
			APIKeyHeader:            "x-api-key",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("BuildTokenURLForDomain and BuildTokenURLForSite mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:               coredata.ConnectorProviderSlack,
			DisplayName:            "Slack",
			BuildTokenURLForDomain: func(string) (string, error) { return "", nil },
			BuildTokenURLForSite:   func(string) (string, error) { return "", nil },
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
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

// TestRegistry_ManagedAPIKey covers the deactivated default (no key
// configured), a configured key, and that an empty key is a no-op so
// the provider stays deactivated.
func TestRegistry_ManagedAPIKey(t *testing.T) {
	t.Parallel()

	r := provider.NewRegistry()

	key, ok := r.ManagedAPIKey(coredata.ConnectorProviderCrisp)
	assert.False(t, ok)
	assert.Empty(t, key)

	r.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "")
	_, ok = r.ManagedAPIKey(coredata.ConnectorProviderCrisp)
	assert.False(t, ok, "empty key must not configure the provider")

	r.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")
	key, ok = r.ManagedAPIKey(coredata.ConnectorProviderCrisp)
	assert.True(t, ok)
	assert.Equal(t, "identifier:secret", key)
}

func TestRegistry_ManagedResourceID(t *testing.T) {
	t.Parallel()

	r := provider.NewRegistry()

	id, ok := r.ManagedResourceID(coredata.ConnectorProviderCrisp)
	assert.False(t, ok)
	assert.Empty(t, id)

	r.SetManagedResourceID(coredata.ConnectorProviderCrisp, "")
	_, ok = r.ManagedResourceID(coredata.ConnectorProviderCrisp)
	assert.False(t, ok, "empty resource id must not configure the provider")

	r.SetManagedResourceID(coredata.ConnectorProviderCrisp, "plugin-id")
	id, ok = r.ManagedResourceID(coredata.ConnectorProviderCrisp)
	assert.True(t, ok)
	assert.Equal(t, "plugin-id", id)
}

// TestCrispIsManagedAPIKey pins Crisp's Model B shape: it is a managed
// API-key provider that does not accept a customer-pasted key, so the
// driver catalog hides it until the operator configures the plugin
// token.
func TestCrispIsManagedAPIKey(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderCrisp)
	require.True(t, ok)
	assert.True(t, reg.ManagedAPIKey)
	assert.False(t, reg.SupportsAPIKey)
	assert.True(t, reg.APIKeyBasicAuthUserPass)
	assert.True(t, reg.RequiresManagedResourceID, "crisp needs the plugin ID before it can connect")
}

// TestRegistry_ManagedConnectorReady pins that a provider requiring a resource
// ID (Crisp's plugin ID) is reported ready, and thus surfaced in the catalog,
// only once BOTH the managed key and the resource ID are configured.
func TestRegistry_ManagedConnectorReady(t *testing.T) {
	t.Parallel()

	t.Run("crisp needs both key and resource id", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		assert.False(t, r.ManagedConnectorReady(coredata.ConnectorProviderCrisp), "unconfigured")

		r.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")
		assert.False(t, r.ManagedConnectorReady(coredata.ConnectorProviderCrisp), "key set but plugin id missing")

		r.SetManagedResourceID(coredata.ConnectorProviderCrisp, "plugin-id")
		assert.True(t, r.ManagedConnectorReady(coredata.ConnectorProviderCrisp), "key and plugin id set")
	})

	t.Run("non-managed provider is never ready", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		r.SetManagedAPIKey(coredata.ConnectorProviderTally, "some-key")
		assert.False(t, r.ManagedConnectorReady(coredata.ConnectorProviderTally))
	})
}

// TestRegistry_RejectsManagedPlusCustomerCredential pins that a
// ManagedAPIKey registration cannot also advertise a customer-supplied
// credential path, whose value would be silently discarded.
func TestRegistry_RejectsManagedPlusCustomerCredential(t *testing.T) {
	t.Parallel()

	r := provider.NewRegistry()
	err := r.Register(&provider.Registration{
		Provider:       coredata.ConnectorProviderCrisp,
		DisplayName:    "Crisp",
		ManagedAPIKey:  true,
		SupportsAPIKey: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestRegistry_ApplyManagedAPIKey verifies the key is injected fresh into a
// managed provider's connection (so rotation propagates and the key is not
// persisted), while non-managed providers are left untouched.
func TestRegistry_ApplyManagedAPIKey(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	r.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")

	managed := &coredata.Connector{
		Provider:   coredata.ConnectorProviderCrisp,
		Connection: &connector.APIKeyConnection{BasicAuthUserPass: true},
	}
	require.NoError(t, r.ApplyManagedAPIKey(managed))
	assert.Equal(t, "identifier:secret", managed.Connection.(*connector.APIKeyConnection).APIKey)

	// Non-managed provider: the connection is left untouched.
	other := &coredata.Connector{
		Provider:   coredata.ConnectorProviderSlack,
		Connection: &connector.APIKeyConnection{APIKey: "customer-key"},
	}
	require.NoError(t, r.ApplyManagedAPIKey(other))
	assert.Equal(t, "customer-key", other.Connection.(*connector.APIKeyConnection).APIKey)
}

// TestRegistry_ApplyManagedAPIKey_Unconfigured verifies that a managed
// provider whose key was never configured (deactivated) errors rather than
// silently building a keyless client.
func TestRegistry_ApplyManagedAPIKey_Unconfigured(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	managed := &coredata.Connector{
		Provider:   coredata.ConnectorProviderCrisp,
		Connection: &connector.APIKeyConnection{BasicAuthUserPass: true},
	}
	err := r.ApplyManagedAPIKey(managed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
