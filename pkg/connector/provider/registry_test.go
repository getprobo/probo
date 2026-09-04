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
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
)

func stubNewCloudSession(
	context.Context,
	*identityfederation.Issuer,
	*coredata.Connector,
) (cloud.Session, error) {
	return nil, nil
}

func stubNewCloudDriver(
	context.Context,
	cloud.Session,
	*coredata.Connector,
	*log.Logger,
) (drivers.Driver, error) {
	return nil, nil
}

func stubNewDriver(
	context.Context,
	*http.Client,
	*coredata.Connector,
	*log.Logger,
	provider.Endpoints,
) (drivers.Driver, error) {
	return nil, nil
}

func stubWorkloadIdentity() *provider.WorkloadIdentityConfig {
	return &provider.WorkloadIdentityConfig{
		NewSession: stubNewCloudSession,
		NewDriver:  stubNewCloudDriver,
		Probe: func(context.Context, cloud.Session, *coredata.Connector) error {
			return nil
		},
	}
}

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

			// A workload identity provider builds its driver from a cloud
			// session rather than an *http.Client, so it wires the factory
			// inside its connect path and leaves NewDriver nil — Register
			// rejects a provider that sets both.
			if reg.SupportsWorkloadIdentity() {
				assert.NotNilf(t, reg.WorkloadIdentity.NewDriver, "provider %q has nil WorkloadIdentity.NewDriver", p)

				return
			}

			assert.NotNilf(t, reg.NewDriver, "provider %q has nil NewDriver", p)
		})
	}
}

// TestEveryProviderSettingsReachADialog asserts that every builtin
// registration declares its extra settings on a connect path it actually
// offers. A list on an unoffered path is a dead declaration: no dialog reads
// it, so the settings never reach the create mutation and the connect attempt
// fails on a field the customer was never asked for. Register rejects the same
// condition at startup; this pins it per provider so the failure names the
// offender rather than panicking inside NewBuiltinRegistry.
func TestEveryProviderSettingsReachADialog(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()

	for _, reg := range r.All() {
		t.Run(string(reg.Provider), func(t *testing.T) {
			t.Parallel()

			if len(reg.APIKeyExtraSettings()) > 0 {
				assert.Truef(
					t,
					reg.SupportsAPIKey() || reg.IsManagedAPIKey(),
					"provider %q declares APIKeyExtraSettings but offers no API-key path",
					reg.Provider,
				)
			}

			if len(reg.ClientCredentialsExtraSettings()) > 0 {
				assert.Truef(
					t,
					reg.SupportsClientCredentials(),
					"provider %q declares ClientCredentialsExtraSettings but offers no client-credentials path",
					reg.Provider,
				)
			}

			if len(reg.WorkloadIdentityExtraSettings()) > 0 {
				assert.Truef(
					t,
					reg.SupportsWorkloadIdentity(),
					"provider %q declares WorkloadIdentityExtraSettings but offers no workload-identity path",
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

	// The three "these two presentations are mutually exclusive" cases this
	// replaces are gone: APIKeyAuth.Mode holds one value, so two presentations
	// cannot be written down. What remains checkable is mode/payload agreement.
	t.Run("API-key auth mode and its payload must agree", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			auth provider.APIKeyAuth
			want string
		}{
			"header without a name": {
				auth: provider.APIKeyAuth{Mode: provider.APIKeyAuthHeader},
				want: `mode "header" requires a Name`,
			},
			"scheme without a name": {
				auth: provider.APIKeyAuth{Mode: provider.APIKeyAuthScheme},
				want: `mode "scheme" requires a Name`,
			},
			"bearer with a name": {
				auth: provider.APIKeyAuth{Name: "x-api-key"},
				want: `mode "" ignores Name`,
			},
			"basic with a name": {
				auth: provider.APIKeyAuth{Mode: provider.APIKeyAuthBasic, Name: "x-api-key"},
				want: `mode "basic" ignores Name`,
			},
			"unknown mode": {
				auth: provider.APIKeyAuth{Mode: provider.APIKeyAuthMode("wat")},
				want: `unknown API-key auth mode "wat"`,
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := provider.NewRegistry().Register(&provider.Registration{
					Provider:    coredata.ConnectorProviderSlack,
					DisplayName: "Slack",
					APIKey:      &provider.APIKeyConfig{Auth: tc.auth},
				})
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want)
			})
		}
	})

	// A KeyFormat is a promise to the customer, shown as the field's
	// placeholder and echoed in the rejection: Register refuses the two ways
	// that promise can be a lie.
	t.Run("KeyFormat rules", func(t *testing.T) {
		t.Parallel()

		for name, tc := range map[string]struct {
			apiKey *provider.APIKeyConfig
			want   string
		}{
			"example its own pattern rejects": {
				apiKey: &provider.APIKeyConfig{
					KeyFormat: &provider.KeyFormat{
						Pattern: regexp.MustCompile(`^pk-lf-.+:sk-lf-.+$`),
						Example: "sk-lf-…",
					},
				},
				want: "example does not match its own pattern",
			},
			"example missing": {
				apiKey: &provider.APIKeyConfig{
					KeyFormat: &provider.KeyFormat{Pattern: regexp.MustCompile(`.`)},
				},
				want: "needs both a Pattern and an Example",
			},
			"pattern missing": {
				apiKey: &provider.APIKeyConfig{
					KeyFormat: &provider.KeyFormat{Example: "pk-lf-…"},
				},
				want: "needs both a Pattern and an Example",
			},
			"shape declared for a key the customer never types": {
				apiKey: &provider.APIKeyConfig{
					Managed: &provider.ManagedAPIKey{},
					KeyFormat: &provider.KeyFormat{
						Pattern: regexp.MustCompile(`^k-.+$`),
						Example: "k-…",
					},
				},
				want: "no customer-supplied key to shape",
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := provider.NewRegistry().Register(&provider.Registration{
					Provider:    coredata.ConnectorProviderSlack,
					DisplayName: "Slack",
					APIKey:      tc.apiKey,
				})
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want)
			})
		}
	})

	t.Run("BuildTokenURLForDomain and BuildTokenURLForSite mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			Endpoints:   provider.Endpoints{Auth: "https://slack.com/oauth/v2/authorize"},
			OAuth2: &provider.OAuth2Config{
				BuildTokenURLForDomain: func(string) (string, error) { return "", nil },
				BuildTokenURLForSite:   func(string) (string, error) { return "", nil },
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mutually exclusive")
	})

	// The OAuth2 block and an authorization URL imply each other, so that a nil
	// block reliably means "no OAuth2 path" rather than "no metadata".
	t.Run("OAuth2 block and Endpoints.Auth imply each other", func(t *testing.T) {
		t.Parallel()

		t.Run("auth without block", func(t *testing.T) {
			t.Parallel()

			err := provider.NewRegistry().Register(&provider.Registration{
				Provider:    coredata.ConnectorProviderSlack,
				DisplayName: "Slack",
				Endpoints:   provider.Endpoints{Auth: "https://slack.com/oauth/v2/authorize"},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Endpoints.Auth requires an OAuth2 block")
		})

		t.Run("block without any authorization URL", func(t *testing.T) {
			t.Parallel()

			err := provider.NewRegistry().Register(&provider.Registration{
				Provider:    coredata.ConnectorProviderSlack,
				DisplayName: "Slack",
				OAuth2:      &provider.OAuth2Config{Scopes: []string{"users:read"}},
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "OAuth2 requires Endpoints.Auth, BuildAuthURL or BuildAuthURLForSite")
		})
	})

	// The two "a settings list needs the path that renders it" rules this
	// replaces are gone: each list lives inside the block that offers the path,
	// so a list without its path cannot be written down.
	t.Run("duplicate setting key within one list", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey: &provider.APIKeyConfig{
				ExtraSettings: []provider.ExtraSetting{
					{Key: "region", Label: "Region"},
					{Key: "region", Label: "Region (again)"},
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `APIKey.ExtraSettings declares duplicate setting key "region"`)
	})

	// One setting more than one dialog needs is declared in each of those
	// lists; that is not a duplicate, because each list keys a separate form.
	t.Run("setting key repeated across lists", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey: &provider.APIKeyConfig{
				ExtraSettings: []provider.ExtraSetting{{Key: "region", Label: "Region"}},
			},
			ClientCredentials: &provider.ClientCredentialsConfig{
				ExtraSettings: []provider.ExtraSetting{{Key: "region", Label: "Region"}},
			},
		})
		require.NoError(t, err)
	})

	t.Run("setting with an empty Key", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey: &provider.APIKeyConfig{
				ExtraSettings: []provider.ExtraSetting{{Label: "Region"}},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "APIKey.ExtraSettings declares a setting with an empty Key or Label")
	})

	t.Run("setting with an empty Key on WorkloadIdentity", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			WorkloadIdentity: &provider.WorkloadIdentityConfig{
				NewSession: stubNewCloudSession,
				NewDriver:  stubNewCloudDriver,
				ExtraSettings: []provider.ExtraSetting{
					{Label: "Account ID"},
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "WorkloadIdentity.ExtraSettings declares a setting with an empty Key or Label")
	})

	t.Run("setting with an empty Label", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			ClientCredentials: &provider.ClientCredentialsConfig{
				ExtraSettings: []provider.ExtraSetting{{Key: "region"}},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ClientCredentials.ExtraSettings declares a setting with an empty Key or Label")
	})

	t.Run("a managed API key excludes client credentials", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
			APIKey: &provider.APIKeyConfig{
				Managed: &provider.ManagedAPIKey{RequiresResourceID: true},
			},
			ClientCredentials: &provider.ClientCredentialsConfig{},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "a managed API key is mutually exclusive with ClientCredentials")
	})

	t.Run("NewDriver and WorkloadIdentity mutually exclusive", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		err := r.Register(&provider.Registration{
			Provider:         coredata.ConnectorProviderSlack,
			DisplayName:      "Slack",
			NewDriver:        stubNewDriver,
			WorkloadIdentity: stubWorkloadIdentity(),
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NewDriver and WorkloadIdentity are mutually exclusive")
	})

	t.Run("workload identity excludes credential-bearing paths", func(t *testing.T) {
		t.Parallel()

		for name, mutate := range map[string]func(*provider.Registration){
			"APIKey": func(reg *provider.Registration) {
				reg.APIKey = &provider.APIKeyConfig{}
			},
			"ClientCredentials": func(reg *provider.Registration) {
				reg.ClientCredentials = &provider.ClientCredentialsConfig{}
			},
			"managed APIKey": func(reg *provider.Registration) {
				reg.APIKey = &provider.APIKeyConfig{Managed: &provider.ManagedAPIKey{}}
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				reg := &provider.Registration{
					Provider:         coredata.ConnectorProviderSlack,
					DisplayName:      "Slack",
					WorkloadIdentity: stubWorkloadIdentity(),
				}
				mutate(reg)

				err := provider.NewRegistry().Register(reg)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "cannot also declare APIKey or ClientCredentials")
			})
		}
	})

	// Grouping the closures makes a dangling Probe unrepresentable, so the only
	// half-configured shape left to reject is a missing required closure.
	t.Run("WorkloadIdentity requires both closures", func(t *testing.T) {
		t.Parallel()

		for name, wi := range map[string]*provider.WorkloadIdentityConfig{
			"driver without session": {NewDriver: stubNewCloudDriver},
			"session without driver": {NewSession: stubNewCloudSession},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				err := provider.NewRegistry().Register(&provider.Registration{
					Provider:         coredata.ConnectorProviderSlack,
					DisplayName:      "Slack",
					WorkloadIdentity: wi,
				})
				require.Error(t, err)
				assert.Contains(t, err.Error(), "requires both NewSession and NewDriver")
			})
		}
	})

	// SupportsWorkloadIdentity is derived from the connect path rather than
	// declared, so the two can never disagree.
	t.Run("workload identity Registration round-trips", func(t *testing.T) {
		t.Parallel()

		r := provider.NewRegistry()
		reg := &provider.Registration{
			Provider:         coredata.ConnectorProviderSlack,
			DisplayName:      "Slack",
			WorkloadIdentity: stubWorkloadIdentity(),
		}
		require.NoError(t, r.Register(reg))
		assert.True(t, reg.SupportsWorkloadIdentity())

		httpOnly := &provider.Registration{
			Provider:    coredata.ConnectorProviderGitHub,
			DisplayName: "GitHub",
			NewDriver:   stubNewDriver,
		}
		require.NoError(t, r.Register(httpOnly))
		assert.False(t, httpOnly.SupportsWorkloadIdentity())
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

	// Managed and customer-supplied are the two variants of the one API-key
	// path, so these two predicates read the same field in opposite directions
	// and can never both be true.
	assert.True(t, reg.IsManagedAPIKey())
	assert.False(t, reg.SupportsAPIKey(), "the customer pastes no key for crisp")
	assert.Equal(
		t,
		provider.APIKeyAuth{Mode: provider.APIKeyAuthBasicUserPass},
		reg.APIKey.Auth,
	)
	assert.True(t, reg.APIKey.Managed.RequiresResourceID, "crisp needs the plugin ID before it can connect")
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

// A managed key can no longer be paired with a customer-supplied one: Managed
// is a variant of the single API-key path, so the conflict the old
// TestRegistry_RejectsManagedPlusCustomerCredential asserted cannot be written
// down. Its surviving half — managed versus client credentials — is covered by
// TestRegistry_Register.

// TestRegistry_APIKeyFor verifies a managed provider resolves the
// Probo-held key without mutating the stored connection, while a
// non-managed provider returns the key already on the connection.
func TestRegistry_APIKeyFor(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	r.SetManagedAPIKey(coredata.ConnectorProviderCrisp, "identifier:secret")

	managed := &connector.APIKeyConnection{BasicAuthUserPass: true}
	key, err := r.APIKeyFor(coredata.ConnectorProviderCrisp, managed)
	require.NoError(t, err)
	assert.Equal(t, "identifier:secret", key)
	assert.Empty(t, managed.APIKey)

	other := &connector.APIKeyConnection{APIKey: "customer-key"}
	key, err = r.APIKeyFor(coredata.ConnectorProviderSlack, other)
	require.NoError(t, err)
	assert.Equal(t, "customer-key", key)
	assert.Equal(t, "customer-key", other.APIKey)
}

// TestRegistry_APIKeyFor_Unconfigured verifies that a managed provider
// whose key was never configured (deactivated) errors rather than
// silently building a keyless client.
func TestRegistry_APIKeyFor_Unconfigured(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	managed := &connector.APIKeyConnection{BasicAuthUserPass: true}
	_, err := r.APIKeyFor(coredata.ConnectorProviderCrisp, managed)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
	assert.Empty(t, managed.APIKey)
}
