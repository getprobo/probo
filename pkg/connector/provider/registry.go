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
	"slices"
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

	// BuildTokenURLForDomain and BuildTokenURLForSite both build the token
	// endpoint host, but from different sources (a callback param vs. the
	// signed state). CompleteWithState checks them in order, so setting both
	// is a programmer error with a silent winner. Reject it at startup.
	if reg.BuildTokenURLForDomain != nil && reg.BuildTokenURLForSite != nil {
		return fmt.Errorf("cannot register connector provider %q: BuildTokenURLForDomain and BuildTokenURLForSite are mutually exclusive", reg.Provider)
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
