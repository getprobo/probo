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

	"go.probo.inc/probo/pkg/connector"
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

	// The auth mode is a single value, so the four presentations it replaces can
	// no longer be set together. What is left to check is that the mode and its
	// payload agree: a Name on a mode that ignores it is silently dead, and a
	// mode that needs one without it would send a header with no name.
	if reg.APIKey != nil {
		switch reg.APIKey.Auth.Mode {
		case APIKeyAuthHeader, APIKeyAuthScheme:
			if reg.APIKey.Auth.Name == "" {
				return fmt.Errorf("cannot register connector provider %q: API-key auth mode %q requires a Name", reg.Provider, reg.APIKey.Auth.Mode)
			}
		case APIKeyAuthBearer, APIKeyAuthBasic, APIKeyAuthBasicUserPass:
			if reg.APIKey.Auth.Name != "" {
				return fmt.Errorf("cannot register connector provider %q: API-key auth mode %q ignores Name", reg.Provider, reg.APIKey.Auth.Mode)
			}
		default:
			return fmt.Errorf("cannot register connector provider %q: unknown API-key auth mode %q", reg.Provider, reg.APIKey.Auth.Mode)
		}
	}

	// A Probo-held key ignores any customer credential, so pairing it with the
	// client-credentials path would advertise a credential field whose value is
	// silently discarded. Its former conflict with a customer-supplied API key
	// is now unrepresentable: Managed is a variant of the one API-key path.
	if reg.IsManagedAPIKey() && reg.SupportsClientCredentials() {
		return fmt.Errorf("cannot register connector provider %q: a managed API key is mutually exclusive with ClientCredentials", reg.Provider)
	}

	// The two driver factories take different credentials, and resolveDriver
	// picks between them from the connector's own connection type. A provider
	// declaring both claims to speak two protocols whose connectors it has no
	// way to create, so one factory would simply never run.
	if reg.NewDriver != nil && reg.SupportsWorkloadIdentity() {
		return fmt.Errorf("cannot register connector provider %q: NewDriver and WorkloadIdentity are mutually exclusive", reg.Provider)
	}

	// A workload identity provider holds no credential, so the customer never
	// supplies one and Probo never injects one. Pairing it with a
	// credential-bearing path would advertise a credential field the driver
	// cannot reach — the same silent-winner class rejected above.
	// Note APIKey != nil rather than SupportsAPIKey(), which answers the
	// narrower "does the customer paste a key": a Probo-held key conflicts with
	// workload identity just as much as a customer-pasted one.
	if reg.SupportsWorkloadIdentity() &&
		(reg.APIKey != nil || reg.SupportsClientCredentials()) {
		return fmt.Errorf("cannot register connector provider %q: a workload identity provider cannot also declare APIKey or ClientCredentials", reg.Provider)
	}

	// Neither closure is usable alone: without a session there is nothing to
	// build a driver from, and without a driver nothing ever opens a session.
	// Grouping them already makes a dangling Probe unrepresentable.
	if wi := reg.WorkloadIdentity; wi != nil && (wi.NewSession == nil || wi.NewDriver == nil) {
		return fmt.Errorf("cannot register connector provider %q: WorkloadIdentity requires both NewSession and NewDriver", reg.Provider)
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

	// An OAuth2 provider is one that can produce an authorization URL, whether
	// static or built per flow. Requiring the block in exactly that case is
	// what lets every reader treat a nil OAuth2 as "no OAuth2 path" instead of
	// "no metadata beyond the defaults", which would silently drop a provider's
	// scopes if someone added it without the block.
	if reg.OAuth2 == nil {
		if reg.Endpoints.Auth != "" {
			return fmt.Errorf("cannot register connector provider %q: Endpoints.Auth requires an OAuth2 block", reg.Provider)
		}
	} else {
		if reg.Endpoints.Auth == "" &&
			reg.OAuth2.BuildAuthURL == nil &&
			reg.OAuth2.BuildAuthURLForSite == nil {
			return fmt.Errorf("cannot register connector provider %q: OAuth2 requires Endpoints.Auth, BuildAuthURL or BuildAuthURLForSite", reg.Provider)
		}

		// BuildTokenURLForDomain and BuildTokenURLForSite both build the token
		// endpoint host, but from different sources (a callback param vs. the
		// signed state). CompleteWithState checks them in order, so setting
		// both is a programmer error with a silent winner. Reject it at startup.
		if reg.OAuth2.BuildTokenURLForDomain != nil && reg.OAuth2.BuildTokenURLForSite != nil {
			return fmt.Errorf("cannot register connector provider %q: BuildTokenURLForDomain and BuildTokenURLForSite are mutually exclusive", reg.Provider)
		}
	}

	// A settings list for a path the provider does not offer is now
	// unrepresentable: each list lives inside the block that offers it.
	//
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
		{"APIKey.ExtraSettings", reg.APIKeyExtraSettings()},
		{"ClientCredentials.ExtraSettings", reg.ClientCredentialsExtraSettings()},
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

// NewAPIKeyConnection builds an API-key connection for provider p that presents
// key the way p's registration declares.
//
// This is the only place an APIKeyAuthMode is translated into the connection's
// presentation fields, so a caller cannot present a key one way on the persisted
// connection and another on a verification client. An unknown provider, or one
// with no API-key path, yields the default Bearer presentation.
func (r *Registry) NewAPIKeyConnection(
	p coredata.ConnectorProvider,
	key string,
) *connector.APIKeyConnection {
	conn := &connector.APIKeyConnection{APIKey: key}

	reg, ok := r.Get(p)
	if !ok || reg.APIKey == nil {
		return conn
	}

	switch auth := reg.APIKey.Auth; auth.Mode {
	case APIKeyAuthHeader:
		conn.Header = auth.Name
	case APIKeyAuthBasic:
		conn.BasicAuth = true
	case APIKeyAuthBasicUserPass:
		conn.BasicAuthUserPass = true
	case APIKeyAuthScheme:
		conn.Scheme = auth.Name
	case APIKeyAuthBearer:
		// The connection's zero value already means Bearer.
	}

	return conn
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
	if !ok || !reg.IsManagedAPIKey() {
		return false
	}

	if _, ok := r.ManagedAPIKey(p); !ok {
		return false
	}

	if reg.APIKey.Managed.RequiresResourceID {
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
	if reg, ok := r.Get(p); ok && reg.OAuth2 != nil {
		// Return a copy so callers cannot mutate the shared, concurrently
		// read registration slice held by this long-lived registry.
		return slices.Clone(reg.OAuth2.Scopes)
	}

	return nil
}
