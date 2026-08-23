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

package provider_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

// Register validates which paths are declared, never what they return. Slack
// stands in below only because it is a valid provider constant.
func stubWorkloadIdentity() *provider.WorkloadIdentitySpec {
	return &provider.WorkloadIdentitySpec{
		NewSession: func(
			context.Context,
			*identityfederation.Issuer,
			*coredata.Connector,
		) (cloud.Session, error) {
			return nil, nil
		},
	}
}

// TestEveryProviderRegistered asserts that every
// coredata.ConnectorProvider constant has a matching Registration in
// the registry, that the registration carries the minimum metadata
// (Provider, DisplayName).
//
// Whether the provider can also be reviewed is asserted next to the drivers, in
// pkg/accessreview/drivers: this catalog only knows how to connect one.
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
		})
	}
}

// TestEveryProviderOffersACredentialPath asserts that every builtin
// registration declares at least one way to authenticate. A provider with none
// can be listed but never connected, and nothing else would catch it: each spec
// is optional on its own, so their absence is only wrong in combination.
//
// A settings list on a path the provider does not offer used to need its own
// check. It cannot happen now: the list lives inside the spec, so declaring one
// declares the path.
func TestEveryProviderOffersACredentialPath(t *testing.T) {
	t.Parallel()

	for _, reg := range provider.NewBuiltinRegistry().All() {
		t.Run(string(reg.Provider), func(t *testing.T) {
			t.Parallel()

			assert.Truef(
				t,
				reg.OAuth2 != nil ||
					reg.APIKey != nil ||
					reg.ClientCredentials != nil ||
					reg.WorkloadIdentity != nil,
				"provider %q declares no credential path, so it can never be connected",
				reg.Provider,
			)

			// An OAuth2 path with no way to build an authorize URL sends the
			// customer to an empty redirect.
			if reg.OAuth2 != nil {
				assert.Truef(
					t,
					reg.Endpoints.Auth != "" ||
						reg.OAuth2.BuildAuthURL != nil ||
						reg.OAuth2.BuildAuthURLForSite != nil,
					"provider %q offers OAuth2 but has no authorize URL",
					reg.Provider,
				)
			}
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

	// Presentation replaced four booleans, so the old "pick one" checks are
	// gone: the type now allows only one. What is left to police is the string
	// two presentations consume and the rest ignore.
	t.Run("custom header presentation requires a name", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey:      &provider.APIKeySpec{Presentation: provider.APIKeyCustomHeader},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires Name")
	})

	t.Run("custom scheme presentation requires a name", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey:      &provider.APIKeySpec{Presentation: provider.APIKeyCustomScheme},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires Name")
	})

	// A name on a presentation that never reads it would silently do nothing,
	// which is how an author discovers too late that the key went out as a
	// Bearer token.
	t.Run("bearer presentation rejects a name", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey:      &provider.APIKeySpec{Name: "x-api-key"},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "consumes no Name")
	})

	t.Run("BuildTokenURLForDomain and BuildTokenURLForSite mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			OAuth2: &provider.OAuth2Spec{
				BuildTokenURLForDomain: func(string) (string, error) { return "", nil },
				BuildTokenURLForSite:   func(string) (string, error) { return "", nil },
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	t.Run("duplicate setting key within one list", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey: &provider.APIKeySpec{
				ExtraSettings: []provider.ExtraSetting{
					{Key: "region", Label: "Region"},
					{Key: "region", Label: "Region (again)"},
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `APIKey.ExtraSettings declares duplicate setting key "region"`)
	})

	// One setting both dialogs need is declared in both lists; that is not a
	// duplicate, because each list keys a separate form.
	t.Run("setting key repeated across the two lists", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:          coredata.ConnectorProviderSlack,
			DisplayName:       "Slack",
			APIKey:            &provider.APIKeySpec{ExtraSettings: []provider.ExtraSetting{{Key: "region", Label: "Region"}}},
			ClientCredentials: &provider.ClientCredentialsSpec{ExtraSettings: []provider.ExtraSetting{{Key: "region", Label: "Region"}}},
		})
		require.NoError(t, err)
	})

	t.Run("setting with an empty Key", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey:      &provider.APIKeySpec{ExtraSettings: []provider.ExtraSetting{{Label: "Region"}}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APIKey.ExtraSettings declares a setting with an empty Key or Label")
	})

	t.Run("setting with an empty Label", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:          coredata.ConnectorProviderSlack,
			DisplayName:       "Slack",
			ClientCredentials: &provider.ClientCredentialsSpec{ExtraSettings: []provider.ExtraSetting{{Key: "region"}}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ClientCredentials.ExtraSettings declares a setting with an empty Key or Label")
	})

	t.Run("RequiresResourceID requires a managed key", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey:      &provider.APIKeySpec{RequiresResourceID: true},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "RequiresResourceID requires Managed")
	})

	t.Run("a managed key rules out client credentials", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:          coredata.ConnectorProviderSlack,
			DisplayName:       "Slack",
			APIKey:            &provider.APIKeySpec{Managed: true},
			ClientCredentials: &provider.ClientCredentialsSpec{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "rules out ClientCredentials")
	})

	// Probo holds no credential when it federates, so any path where it does
	// would advertise a field this provider can never use.
	t.Run("workload identity rules out every path Probo holds a credential on", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name string
			reg  *provider.Registration
		}{
			{
				name: "OAuth2",
				reg: &provider.Registration{
					Provider:         coredata.ConnectorProviderSlack,
					DisplayName:      "Slack",
					OAuth2:           &provider.OAuth2Spec{},
					WorkloadIdentity: stubWorkloadIdentity(),
				},
			},
			{
				name: "APIKey",
				reg: &provider.Registration{
					Provider:         coredata.ConnectorProviderSlack,
					DisplayName:      "Slack",
					APIKey:           &provider.APIKeySpec{},
					WorkloadIdentity: stubWorkloadIdentity(),
				},
			},
			{
				name: "ClientCredentials",
				reg: &provider.Registration{
					Provider:          coredata.ConnectorProviderSlack,
					DisplayName:       "Slack",
					ClientCredentials: &provider.ClientCredentialsSpec{},
					WorkloadIdentity:  stubWorkloadIdentity(),
				},
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()

				r := provider.NewRegistry()
				err := r.Register(tc.reg)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "WorkloadIdentity rules out")
			})
		}
	})

	// Without it every capability built on this connector fails at use time, on
	// a provider the catalog advertised as connectable.
	t.Run("workload identity requires a session factory", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:         coredata.ConnectorProviderSlack,
			DisplayName:      "Slack",
			WorkloadIdentity: &provider.WorkloadIdentitySpec{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WorkloadIdentity requires NewSession")
	})

	t.Run("complete workload identity Registration", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:         coredata.ConnectorProviderSlack,
			DisplayName:      "Slack",
			WorkloadIdentity: stubWorkloadIdentity(),
		})
		require.NoError(t, err)
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
	require.NotNil(t, reg.APIKey)
	assert.True(t, reg.APIKey.Managed)
	assert.False(t, reg.AcceptsCustomerAPIKey(), "crisp accepts no customer-pasted key")
	assert.Equal(t, provider.APIKeyBasicUserPass, reg.APIKey.Presentation)
	assert.True(t, reg.APIKey.RequiresResourceID, "crisp needs the plugin ID before it can connect")
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

// Who supplies the key is now one field on one spec, so a provider cannot
// advertise both a Probo-held key and a customer-pasted one — a pairing that
// used to need its own rejection rule because the pasted value would have been
// silently discarded. What is left to pin is that the console can still tell the
// two apart, since only one of them renders a credential field.
func TestRegistrationAcceptsCustomerAPIKey(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	crisp, ok := r.Get(coredata.ConnectorProviderCrisp)
	require.True(t, ok)
	assert.False(t, crisp.AcceptsCustomerAPIKey(), "crisp's key comes from bootstrap config")

	github, ok := r.Get(coredata.ConnectorProviderGitHub)
	require.True(t, ok)
	assert.True(t, github.AcceptsCustomerAPIKey(), "github accepts a pasted token")

	slack, ok := r.Get(coredata.ConnectorProviderSlack)
	require.True(t, ok)
	assert.False(t, slack.AcceptsCustomerAPIKey(), "slack is OAuth2-only")
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
