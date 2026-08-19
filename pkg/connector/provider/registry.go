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

// Package provider holds one Go file per connector provider. Each file
// exposes a private constructor that returns a *Registration; the
// builtin set is assembled by NewBuiltinRegistry, which probod calls
// once at startup and threads as an explicit *Registry into every
// consumer. The registry carries no package-level state.
//
// pkg/connector/provider is a sub-package of pkg/connector. The
// child may import its parent (it does — for the *OAuth2Connector
// type in apply.go); the parent must not import this child. Cycles
// with pkg/coredata are avoided because the back-edge runs:
// provider -> connector -> coredata -> (no further imports back).
package provider

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"

	"go.probo.inc/probo/pkg/coredata"
)

// Registry holds the per-provider *Registration set used by the rest
// of the system to look up display names, OAuth2 metadata, driver
// constructors, and so on. It is safe for concurrent use.
//
// All consumers receive a *Registry constructed by NewBuiltinRegistry
// at probod startup; no package-level singleton exists.
type Registry struct {
	mu        sync.RWMutex
	providers map[coredata.ConnectorProvider]*Registration
	// managedAPIKeys holds the Probo-supplied API key for providers with
	// ManagedAPIKey registrations (e.g. Crisp's marketplace plugin token).
	// Populated by probod from bootstrap config via SetManagedAPIKey; empty
	// until the operator configures the credential.
	managedAPIKeys map[coredata.ConnectorProvider]string
	// managedResourceIDs holds an optional Probo-supplied resource identifier
	// for a ManagedAPIKey provider, distinct from the credential. Crisp needs
	// it: the plugin token's Basic identifier is not the plugin ID, yet the
	// per-website plugin API (used for ownership verification) requires the
	// plugin ID in the path. Populated by probod via SetManagedResourceID;
	// empty for providers that need no such identifier.
	managedResourceIDs map[coredata.ConnectorProvider]string
}

// NewRegistry returns an empty *Registry. Production code uses
// NewBuiltinRegistry; tests and specialised callers can construct an
// empty Registry and register only the providers they need.
func NewRegistry() *Registry {
	return &Registry{
		providers:          make(map[coredata.ConnectorProvider]*Registration),
		managedAPIKeys:     make(map[coredata.ConnectorProvider]string),
		managedResourceIDs: make(map[coredata.ConnectorProvider]string),
	}
}

// Register adds a Registration to r. It returns an error on nil or
// incomplete Registration metadata or on duplicate registration so
// callers (in particular NewBuiltinRegistry) can decide whether the
// condition is a programmer error worth crashing on or a recoverable
// state worth surfacing.
func (r *Registry) Register(reg *Registration) error {
	if reg == nil {
		return fmt.Errorf("cannot register connector provider: nil Registration")
	}

	if reg.Provider == "" {
		return fmt.Errorf("cannot register connector provider: missing Provider")
	}

	if reg.DisplayName == "" {
		return fmt.Errorf("cannot register connector provider %q: missing DisplayName", reg.Provider)
	}

	// APIKeyBasicAuth, APIKeyBasicAuthUserPass, APIKeyHeader, and
	// APIKeyAuthScheme select different presentations of the same key;
	// setting more than one is a programmer error with a silent winner
	// (Client checks BasicAuth, then BasicAuthUserPass, then Header, then
	// Scheme). Reject it at startup.
	apiKeyModes := 0

	if reg.APIKeyBasicAuth {
		apiKeyModes++
	}

	if reg.APIKeyBasicAuthUserPass {
		apiKeyModes++
	}

	if reg.APIKeyHeader != "" {
		apiKeyModes++
	}

	if reg.APIKeyAuthScheme != "" {
		apiKeyModes++
	}

	if apiKeyModes > 1 {
		return fmt.Errorf("cannot register connector provider %q: APIKeyBasicAuth, APIKeyBasicAuthUserPass, APIKeyHeader, and APIKeyAuthScheme are mutually exclusive", reg.Provider)
	}

	// ManagedAPIKey injects a Probo-held key and ignores any customer
	// credential, so pairing it with SupportsAPIKey/SupportsClientCredentials
	// would advertise a credential field whose value is silently discarded —
	// the same silent-winner class rejected above. Reject it at startup.
	if reg.ManagedAPIKey && (reg.SupportsAPIKey || reg.SupportsClientCredentials) {
		return fmt.Errorf("cannot register connector provider %q: ManagedAPIKey is mutually exclusive with SupportsAPIKey and SupportsClientCredentials", reg.Provider)
	}

	// RequiresManagedResourceID only has meaning for a ManagedAPIKey provider:
	// ManagedConnectorReady consults it exclusively on that path, so setting it
	// on a non-managed provider is a silently ineffective flag. Reject it at
	// startup rather than let the requirement quietly do nothing.
	if reg.RequiresManagedResourceID && !reg.ManagedAPIKey {
		return fmt.Errorf("cannot register connector provider %q: RequiresManagedResourceID requires ManagedAPIKey", reg.Provider)
	}

	// A workload identity provider holds no credential, so a key-based path
	// would advertise a field it can never use, and the two families select
	// different driver factories — the silent-winner class rejected above.
	if reg.SupportsWorkloadIdentity &&
		(reg.SupportsAPIKey || reg.SupportsClientCredentials || reg.ManagedAPIKey) {
		return fmt.Errorf("cannot register connector provider %q: SupportsWorkloadIdentity is mutually exclusive with SupportsAPIKey, SupportsClientCredentials, and ManagedAPIKey", reg.Provider)
	}

	// NewCloudSession without the flag is silently ineffective: the driver
	// catalog gate and the mutual exclusions above both read the flag, so the
	// provider would advertise a key-based path it cannot serve.
	if reg.NewCloudSession != nil && !reg.SupportsWorkloadIdentity {
		return fmt.Errorf("cannot register connector provider %q: NewCloudSession requires SupportsWorkloadIdentity", reg.Provider)
	}

	// Open fills a workload identity Handle through NewCloudSession, so without
	// it every capability on that Handle fails at use time — on a provider the
	// catalog advertised as connectable.
	if reg.SupportsWorkloadIdentity && reg.NewCloudSession == nil {
		return fmt.Errorf("cannot register connector provider %q: SupportsWorkloadIdentity requires NewCloudSession", reg.Provider)
	}

	// A Probe on a different host from APIBase (or Identity) would let a
	// deployment move the driver to another host while the connection check
	// keeps hitting the real provider — a half-migrated connector that
	// reports healthy. All three describe the same provider's API or identity
	// surface, so their hosts must agree. This is what turns an override that
	// moves Probe without moving the matching field (an operator forgetting
	// DocuSign's Identity, say) into a boot failure instead of the silent
	// split applyEndpointOverride's own check cannot see, because it only
	// runs before Register on the still-being-assembled Endpoints.
	if reg.Endpoints.Probe != "" && (reg.Endpoints.APIBase != "" || reg.Endpoints.Identity != "") {
		probe, err := url.Parse(reg.Endpoints.Probe)
		if err != nil {
			return fmt.Errorf("cannot register connector provider %q: cannot parse Probe: %w", reg.Provider, err)
		}

		if reg.Endpoints.APIBase != "" {
			base, err := url.Parse(reg.Endpoints.APIBase)
			if err != nil {
				return fmt.Errorf("cannot register connector provider %q: cannot parse APIBase: %w", reg.Provider, err)
			}

			if !strings.EqualFold(base.Host, probe.Host) {
				return fmt.Errorf("cannot register connector provider %q: Probe host %q does not match APIBase host %q", reg.Provider, probe.Host, base.Host)
			}
		}

		if reg.Endpoints.Identity != "" {
			identity, err := url.Parse(reg.Endpoints.Identity)
			if err != nil {
				return fmt.Errorf("cannot register connector provider %q: cannot parse Identity: %w", reg.Provider, err)
			}

			if !strings.EqualFold(identity.Host, probe.Host) {
				return fmt.Errorf("cannot register connector provider %q: Probe host %q does not match Identity host %q", reg.Provider, probe.Host, identity.Host)
			}
		}
	}

	// BuildTokenURLForDomain and BuildTokenURLForSite both build the token
	// endpoint host, but from different sources (a callback param vs. the
	// signed state). CompleteWithState checks them in order, so setting both
	// is a programmer error with a silent winner. Reject it at startup.
	if reg.BuildTokenURLForDomain != nil && reg.BuildTokenURLForSite != nil {
		return fmt.Errorf("cannot register connector provider %q: BuildTokenURLForDomain and BuildTokenURLForSite are mutually exclusive", reg.Provider)
	}

	// A per-path settings list for a path the provider cannot offer is a dead
	// declaration: no dialog will ever render it. ManagedAPIKey counts as an
	// API-key path — the customer supplies the settings, Probo the key.
	if len(reg.APIKeyExtraSettings) > 0 && !reg.SupportsAPIKey && !reg.ManagedAPIKey {
		return fmt.Errorf("cannot register connector provider %q: APIKeyExtraSettings requires SupportsAPIKey or ManagedAPIKey", reg.Provider)
	}

	if len(reg.ClientCredentialsExtraSettings) > 0 && !reg.SupportsClientCredentials {
		return fmt.Errorf("cannot register connector provider %q: ClientCredentialsExtraSettings requires SupportsClientCredentials", reg.Provider)
	}

	// The console keys both its form state and its submitted values by setting
	// key within one dialog, so a duplicate key silently collapses two fields
	// into one and an empty key produces an unlabelled field bound to nothing.
	// Reject both at startup. A key repeated across the two lists is fine and
	// intended: that is how a dual-path provider declares one setting both
	// dialogs need.
	for _, list := range []struct {
		field    string
		settings []ExtraSetting
	}{
		{"APIKeyExtraSettings", reg.APIKeyExtraSettings},
		{"ClientCredentialsExtraSettings", reg.ClientCredentialsExtraSettings},
	} {
		seen := make(map[string]bool, len(list.settings))

		for _, s := range list.settings {
			if s.Key == "" || s.Label == "" {
				return fmt.Errorf("cannot register connector provider %q: %s declares a setting with an empty Key or Label", reg.Provider, list.field)
			}

			if seen[s.Key] {
				return fmt.Errorf("cannot register connector provider %q: %s declares duplicate setting key %q", reg.Provider, list.field, s.Key)
			}

			seen[s.Key] = true
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, dup := r.providers[reg.Provider]; dup {
		return fmt.Errorf("cannot register connector provider %q: duplicate registration", reg.Provider)
	}

	r.providers[reg.Provider] = reg

	return nil
}

// Get returns the Registration for the given provider, or false if
// no provider is registered under that key.
func (r *Registry) Get(p coredata.ConnectorProvider) (*Registration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	reg, ok := r.providers[p]

	return reg, ok
}

// For returns the Registration behind an opened connector, or false when the
// registry knows no such provider. It is Get for a caller holding a Handle,
// which owns no Registration of its own.
func (r *Registry) For(h *Handle) (*Registration, bool) {
	return r.Get(h.Connector.Provider)
}

// All returns every Registration currently in r. Order is not stable;
// callers must sort when determinism matters.
func (r *Registry) All() []*Registration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]*Registration, 0, len(r.providers))
	for _, reg := range r.providers {
		out = append(out, reg)
	}

	return out
}

// PublicClients returns every Registration flagged PublicClient (CIMD,
// no client_secret). probod uses this to auto-register their OAuth2
// connectors with a deployment-derived client_id and state-signing key.
// Order is not stable.
func (r *Registry) PublicClients() []*Registration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []*Registration

	for _, reg := range r.providers {
		if reg.PublicClient {
			out = append(out, reg)
		}
	}

	return out
}

// ProviderDisplayName returns the human-readable label for the
// provider, falling back to the raw constant string when no display
// name is registered.
func (r *Registry) ProviderDisplayName(p coredata.ConnectorProvider) string {
	if reg, ok := r.Get(p); ok && reg.DisplayName != "" {
		return reg.DisplayName
	}

	return string(p)
}

// APIKeyHeader returns the request header an API-key connection for the
// given provider must use to present its key. Empty means the default
// `Authorization: Bearer` scheme; a value such as "x-api-key" means the
// raw key is sent in that header instead. Returns empty for unknown
// providers and for providers that do not customise the scheme.
func (r *Registry) APIKeyHeader(p coredata.ConnectorProvider) string {
	if reg, ok := r.Get(p); ok {
		return reg.APIKeyHeader
	}

	return ""
}

// APIKeyUsesBasicAuth reports whether an API-key connection for the
// given provider must present its key as an HTTP Basic auth username
// (empty password) instead of a Bearer token. Returns false for unknown
// providers and for providers that use the default Bearer scheme.
func (r *Registry) APIKeyUsesBasicAuth(p coredata.ConnectorProvider) bool {
	if reg, ok := r.Get(p); ok {
		return reg.APIKeyBasicAuth
	}

	return false
}

// APIKeyAuthScheme returns the non-Bearer Authorization scheme an API-key
// connection for the given provider must use to present its key (e.g.
// "SSWS" for Okta). Empty means the default `Authorization: Bearer`
// scheme. Returns empty for unknown providers and for providers that do
// not customise the scheme.
func (r *Registry) APIKeyAuthScheme(p coredata.ConnectorProvider) string {
	if reg, ok := r.Get(p); ok {
		return reg.APIKeyAuthScheme
	}

	return ""
}

// APIKeyUsesBasicAuthUserPass reports whether an API-key connection for the
// given provider must present its key as a complete HTTP Basic credential
// (`username:password` already encoded in the key, base64'd verbatim)
// instead of a Bearer token. Returns false for unknown providers and for
// providers that use the default Bearer scheme.
func (r *Registry) APIKeyUsesBasicAuthUserPass(p coredata.ConnectorProvider) bool {
	if reg, ok := r.Get(p); ok {
		return reg.APIKeyBasicAuthUserPass
	}

	return false
}

// SetManagedAPIKey records the Probo-supplied API key for a
// ManagedAPIKey provider (e.g. Crisp). probod calls this from bootstrap
// config so the create-connector resolver can inject the key and the
// driver catalog can surface the provider. An empty key is treated as
// "not configured": it is not stored, keeping the provider hidden.
func (r *Registry) SetManagedAPIKey(p coredata.ConnectorProvider, key string) {
	if key == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.managedAPIKeys[p] = key
}

// ManagedAPIKey returns the Probo-supplied API key configured for a
// ManagedAPIKey provider and whether one is set. The boolean is false
// (and the string empty) until the operator configures the credential
// via bootstrap, which is what keeps such a provider deactivated.
func (r *Registry) ManagedAPIKey(p coredata.ConnectorProvider) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key, ok := r.managedAPIKeys[p]

	return key, ok
}

// SetManagedResourceID records an optional Probo-supplied resource
// identifier for a ManagedAPIKey provider (e.g. the Crisp plugin ID used
// by the per-website plugin API). probod calls this from bootstrap config
// alongside SetManagedAPIKey. An empty id is treated as "not configured":
// it is not stored.
func (r *Registry) SetManagedResourceID(p coredata.ConnectorProvider, id string) {
	if id == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.managedResourceIDs[p] = id
}

// ManagedResourceID returns the Probo-supplied resource identifier
// configured for a ManagedAPIKey provider and whether one is set. The
// boolean is false (and the string empty) until the operator configures it
// via bootstrap.
func (r *Registry) ManagedResourceID(p coredata.ConnectorProvider) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.managedResourceIDs[p]

	return id, ok
}

// ManagedConnectorReady reports whether a ManagedAPIKey provider is fully
// configured for this deployment: its Probo-held key is set and, when the
// provider also requires a resource ID (RequiresManagedResourceID, e.g. the
// Crisp plugin ID), that is set too. A provider that is not ready is kept out
// of the driver catalog, since connecting it would fail at verify time. It is
// false for non-managed and unregistered providers.
func (r *Registry) ManagedConnectorReady(p coredata.ConnectorProvider) bool {
	reg, ok := r.Get(p)
	if !ok || !reg.ManagedAPIKey {
		return false
	}

	if _, ok := r.ManagedAPIKey(p); !ok {
		return false
	}

	if reg.RequiresManagedResourceID {
		if _, ok := r.ManagedResourceID(p); !ok {
			return false
		}
	}

	return true
}

// ProviderOAuth2Scopes returns the OAuth2 scopes the access review
// driver for the given provider needs to list user accounts. Returns
// nil for providers that do not need any scopes (Notion, Intercom)
// or for non-access-review providers.
func (r *Registry) ProviderOAuth2Scopes(p coredata.ConnectorProvider) []string {
	if reg, ok := r.Get(p); ok {
		// Return a copy so callers cannot mutate the shared, concurrently
		// read registration slice held by this long-lived registry.
		return slices.Clone(reg.OAuth2Scopes)
	}

	return nil
}
