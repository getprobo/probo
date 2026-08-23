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

// Register adds a Registration to r. It returns an error on nil or incomplete
// metadata, on a contradictory credential path, or on a duplicate, so callers
// (in particular NewBuiltinRegistry) can decide whether the condition is a
// programmer error worth crashing on or a recoverable state worth surfacing.
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

	if err := reg.validateCredentialPaths(); err != nil {
		return err
	}

	if err := reg.validateProbeHost(); err != nil {
		return err
	}

	if err := reg.validateExtraSettings(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, dup := r.providers[reg.Provider]; dup {
		return fmt.Errorf("cannot register connector provider %q: duplicate registration", reg.Provider)
	}

	r.providers[reg.Provider] = reg

	return nil
}

// validateCredentialPaths rejects a provider whose declared paths contradict
// each other. Grouping each path into its own spec leaves only the rules that
// say something about the domain.
func (reg *Registration) validateCredentialPaths() error {
	// Federating into a customer's cloud means Probo holds no credential, so
	// any path where it does would advertise a field this provider can never
	// use — and the two families need different consumer code.
	if reg.WorkloadIdentity != nil {
		if reg.OAuth2 != nil || reg.APIKey != nil || reg.ClientCredentials != nil {
			return fmt.Errorf("cannot register connector provider %q: WorkloadIdentity rules out OAuth2, APIKey, and ClientCredentials", reg.Provider)
		}

		// Without it every capability built on this connector fails at use
		// time, on a provider the catalog advertised as connectable.
		if reg.WorkloadIdentity.NewSession == nil {
			return fmt.Errorf("cannot register connector provider %q: WorkloadIdentity requires NewSession", reg.Provider)
		}
	}

	if reg.APIKey != nil {
		// Name is consumed by exactly two presentations. Set anywhere else it
		// is silently ignored; missing on these two it yields a request with an
		// empty header name or scheme.
		needsName := reg.APIKey.Presentation == APIKeyCustomHeader ||
			reg.APIKey.Presentation == APIKeyCustomScheme

		if needsName && reg.APIKey.Name == "" {
			return fmt.Errorf("cannot register connector provider %q: APIKey presentation %q requires Name", reg.Provider, reg.APIKey.Presentation)
		}

		if !needsName && reg.APIKey.Name != "" {
			return fmt.Errorf("cannot register connector provider %q: APIKey presentation %q consumes no Name", reg.Provider, reg.APIKey.Presentation)
		}

		// ManagedConnectorReady consults RequiresResourceID only on the managed
		// path, so setting it elsewhere is a requirement that quietly does
		// nothing.
		if reg.APIKey.RequiresResourceID && !reg.APIKey.Managed {
			return fmt.Errorf("cannot register connector provider %q: APIKey.RequiresResourceID requires Managed", reg.Provider)
		}

		// A managed key is injected at use time and any customer credential is
		// discarded, so offering another key-based path would advertise a field
		// whose value is thrown away.
		if reg.APIKey.Managed && reg.ClientCredentials != nil {
			return fmt.Errorf("cannot register connector provider %q: a managed APIKey rules out ClientCredentials", reg.Provider)
		}
	}

	// CompleteWithState checks the two token-URL builders in order, so setting
	// both is a programmer error with a silent winner.
	if reg.OAuth2 != nil &&
		reg.OAuth2.BuildTokenURLForDomain != nil &&
		reg.OAuth2.BuildTokenURLForSite != nil {
		return fmt.Errorf("cannot register connector provider %q: BuildTokenURLForDomain and BuildTokenURLForSite are mutually exclusive", reg.Provider)
	}

	return nil
}

// validateProbeHost rejects a Probe on a different host from the API or
// identity surface it checks.
//
// Such a split would let a deployment move the driver to another host while the
// connection check keeps hitting the real provider — a half-migrated connector
// that reports healthy. This is what turns an override that moves Probe without
// moving the matching field (an operator forgetting DocuSign's Identity, say)
// into a boot failure instead of the silent split applyEndpointOverride cannot
// see, because that only runs before Register on the still-being-assembled
// Endpoints.
func (reg *Registration) validateProbeHost() error {
	if reg.Endpoints.Probe == "" {
		return nil
	}

	probe, err := url.Parse(reg.Endpoints.Probe)
	if err != nil {
		return fmt.Errorf("cannot register connector provider %q: cannot parse Probe: %w", reg.Provider, err)
	}

	for _, surface := range []struct {
		field string
		value string
	}{
		{"APIBase", reg.Endpoints.APIBase},
		{"Identity", reg.Endpoints.Identity},
	} {
		if surface.value == "" {
			continue
		}

		parsed, err := url.Parse(surface.value)
		if err != nil {
			return fmt.Errorf("cannot register connector provider %q: cannot parse %s: %w", reg.Provider, surface.field, err)
		}

		if !strings.EqualFold(parsed.Host, probe.Host) {
			return fmt.Errorf("cannot register connector provider %q: Probe host %q does not match %s host %q", reg.Provider, probe.Host, surface.field, parsed.Host)
		}
	}

	return nil
}

// validateExtraSettings rejects a settings list the console cannot render.
//
// The console keys both its form state and its submitted values by setting key
// within one dialog, so a duplicate key silently collapses two fields into one
// and an empty key produces an unlabelled field bound to nothing. A key
// repeated ACROSS the two lists is fine and intended: that is how a dual-path
// provider declares one setting both dialogs need.
func (reg *Registration) validateExtraSettings() error {
	lists := []struct {
		field    string
		settings []ExtraSetting
	}{}

	if reg.APIKey != nil {
		lists = append(lists, struct {
			field    string
			settings []ExtraSetting
		}{"APIKey.ExtraSettings", reg.APIKey.ExtraSettings})
	}

	if reg.ClientCredentials != nil {
		lists = append(lists, struct {
			field    string
			settings []ExtraSetting
		}{"ClientCredentials.ExtraSettings", reg.ClientCredentials.ExtraSettings})
	}

	for _, list := range lists {
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

// PublicClients returns every OAuth2 provider that authenticates as a public
// client (CIMD, no client_secret). probod uses this to auto-register their
// OAuth2 connectors with a deployment-derived client_id and state-signing key.
// Order is not stable.
func (r *Registry) PublicClients() []*Registration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var out []*Registration

	for _, reg := range r.providers {
		if reg.OAuth2 != nil && reg.OAuth2.PublicClient {
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

// apiKeySpec returns the provider's API-key path, or nil when it offers none.
// The accessors below read through it so a caller building an APIKeyConnection
// need not know the registration's shape.
func (r *Registry) apiKeySpec(p coredata.ConnectorProvider) *APIKeySpec {
	reg, ok := r.Get(p)
	if !ok {
		return nil
	}

	return reg.APIKey
}

// APIKeyHeader returns the request header an API-key connection for the given
// provider must send its raw key in, and "" when the key travels in the
// Authorization header instead.
func (r *Registry) APIKeyHeader(p coredata.ConnectorProvider) string {
	spec := r.apiKeySpec(p)
	if spec == nil || spec.Presentation != APIKeyCustomHeader {
		return ""
	}

	return spec.Name
}

// APIKeyAuthScheme returns the non-Bearer Authorization scheme an API-key
// connection for the given provider must use (e.g. "SSWS" for Okta), and "" for
// the default Bearer scheme.
func (r *Registry) APIKeyAuthScheme(p coredata.ConnectorProvider) string {
	spec := r.apiKeySpec(p)
	if spec == nil || spec.Presentation != APIKeyCustomScheme {
		return ""
	}

	return spec.Name
}

// APIKeyUsesBasicAuth reports whether the key is presented as a Basic auth
// username with an empty password instead of a Bearer token.
func (r *Registry) APIKeyUsesBasicAuth(p coredata.ConnectorProvider) bool {
	spec := r.apiKeySpec(p)

	return spec != nil && spec.Presentation == APIKeyBasic
}

// APIKeyUsesBasicAuthUserPass reports whether the key is presented as a
// complete Basic credential whose `username:password` pair is already inside it.
func (r *Registry) APIKeyUsesBasicAuthUserPass(p coredata.ConnectorProvider) bool {
	spec := r.apiKeySpec(p)

	return spec != nil && spec.Presentation == APIKeyBasicUserPass
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
	spec := r.apiKeySpec(p)
	if spec == nil || !spec.Managed {
		return false
	}

	if _, ok := r.ManagedAPIKey(p); !ok {
		return false
	}

	if spec.RequiresResourceID {
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
	reg, ok := r.Get(p)
	if !ok || reg.OAuth2 == nil {
		return nil
	}

	// Return a copy so callers cannot mutate the shared, concurrently read
	// registration slice held by this long-lived registry.
	return slices.Clone(reg.OAuth2.Scopes)
}
