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

package provider

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// Endpoints groups every host-bearing URL a provider owns, so a deployment
// can reason about — and later override — one struct instead of chasing
// literals across pkg/connector/provider and pkg/accessreview/drivers.
//
// Auth, Token and Probe are complete URLs, query string included when the
// provider needs one. APIBase is a scheme-ful origin, optionally carrying a
// version segment, with NO trailing slash; drivers join paths onto it with
// url.JoinPath.
//
// An empty field means "not expressible here", never "unknown": an
// API-key-only provider has no Auth or Token, a provider with BuildProbeURL
// or a Probe closure has no static Probe, and APIBase is empty whenever the
// data host is per-connection or discovered at runtime.
type Endpoints struct {
	// Auth is the OAuth2 authorization endpoint. It is frequently a
	// DIFFERENT host from APIBase (github.com vs api.github.com,
	// identity.pagerduty.com vs api.pagerduty.com) and must never be derived
	// from it. Empty for API-key-only providers and for providers using
	// BuildAuthURL or BuildAuthURLForSite.
	Auth string

	// Token is the OAuth2 token endpoint, independent of Auth: Monday serves
	// both from auth.monday.com, while HubSpot authorizes on app.hubspot.com
	// and exchanges on api.hubapi.com. Empty for API-key-only providers, for
	// BuildTokenURLForDomain / BuildTokenURLForSite providers, and for
	// 1Password, whose token endpoint is supplied per connection.
	Token string

	// Probe is the connection-check URL for a plain authenticated GET. Empty
	// when BuildProbeURL or the Probe closure is set. When both Probe and
	// APIBase are non-empty their hosts MUST match — Register enforces it, so
	// an APIBase override can never leave the connection check pointed at the
	// real provider while the driver talks to the override.
	Probe string

	// APIBase is the data API root the driver joins paths onto. Set it ONLY
	// for a compile-time-constant host. Leave it EMPTY when the host comes
	// from connector settings (Grafana, Metabase, Okta, SigNoz, Zendesk,
	// Datadog, Langfuse, Segment, PostHog) or is discovered at runtime
	// (DocuSign's per-account base_uri, PostHog's lazy region probe) — the
	// NewDriver closure resolves those, with pkg/accessreview/drivers/posthog.go
	// as the reference implementation.
	APIBase string
}

// Registration is the per-provider metadata + factory bundle. Each
// provider returns one of these from a private constructor (e.g.
// slackRegistration) that NewBuiltinRegistry assembles into the
// runtime *Registry. Fields are grouped by concern: identity, OAuth2
// metadata, supported protocols, extra settings, and factory closures.
type Registration struct {
	// Identity.
	Provider    coredata.ConnectorProvider
	DisplayName string
	// DocumentationURL is the public probo.com docs page for connecting this
	// provider as an access source. Empty for providers with no doc page yet;
	// surfaced (nullable) on ConnectorProviderInfo so the console renders a link.
	DocumentationURL string

	// Endpoints groups every host-bearing URL this provider owns.
	Endpoints Endpoints

	// OAuth2 metadata.
	ExtraAuthParams         map[string]string
	TokenEndpointAuth       string // "post-form" (default), "basic-form", or "basic-json"
	SupportsIncrementalAuth bool
	OAuth2Scopes            []string
	// RequiresPKCE enables RFC 7636 PKCE (S256) on the authorization
	// request and replays the verifier on the token exchange. Default
	// false; non-PKCE providers are unaffected.
	RequiresPKCE bool
	// PublicClient marks an OAuth2 provider that authenticates as a public
	// client (no client_secret) via PKCE, using the Client ID Metadata
	// Document (CIMD) flow. probod auto-registers such providers with no
	// operator credentials: the client_id is the deployment's hosted CIMD
	// URL (baseURL + connector.CIMDMetadataPath) and the state token is
	// signed with a server-derived key. Set TokenEndpointAuth to "none"
	// alongside this.
	PublicClient bool
	// BuildAuthURL derives the authorization URL from an operator-supplied
	// integration slug, for providers (e.g. Vercel) whose AuthURL embeds
	// it as a path segment. It must construct the URL with net/url and
	// escape the slug. Nil for providers with a fully static AuthURL.
	BuildAuthURL func(slug string) (string, error)
	// BuildAuthURLForSite builds the authorize URL for a per-customer
	// site supplied at initiate time (multi-site providers, e.g.
	// Datadog). It MUST validate site against a fixed allow-list and
	// construct the URL with net/url. Nil for single-site providers.
	BuildAuthURLForSite func(site string) (string, error)
	// BuildTokenURLForDomain builds the token endpoint URL from the API
	// domain the provider returns on the OAuth callback (multi-site
	// providers, e.g. Datadog). It MUST validate domain. Nil otherwise.
	BuildTokenURLForDomain func(domain string) (string, error)
	// BuildTokenURLForSite builds the token endpoint URL from the
	// per-customer site/subdomain carried in the signed OAuth state, for
	// multi-site providers whose token host the provider does NOT echo back
	// on the callback (e.g. Zendesk's <subdomain>.zendesk.com). It MUST
	// validate site. A provider sets at most one of BuildTokenURLForDomain /
	// BuildTokenURLForSite. Nil otherwise.
	BuildTokenURLForSite func(site string) (string, error)

	// Protocol support / GraphQL surface.
	SupportsAPIKey            bool
	SupportsClientCredentials bool
	// APIKeyExtraSettings declares the per-provider settings fields the
	// console's API-key connect dialog renders and submits, in render order.
	// It covers a ManagedAPIKey provider too (Crisp): the customer supplies
	// the settings, Probo supplies the key. Nil for a provider with no
	// API-key path.
	APIKeyExtraSettings []ExtraSetting
	// ClientCredentialsExtraSettings declares the settings fields the
	// client-credentials connect dialog renders and submits. The two lists are
	// independent because a different create resolver and a different driver
	// sit behind each path: 1Password needs SCIMBridgeURL on the API key
	// (SCIM-bridge driver) and AccountID + Region on client credentials (Users
	// API driver). A setting genuinely needed on both paths is declared in
	// both lists.
	ClientCredentialsExtraSettings []ExtraSetting
	// APIKeyHeader selects how an API-key connection presents its key
	// on outbound requests. Empty (the default) uses the standard
	// `Authorization: Bearer <key>` scheme; a value such as "x-api-key"
	// sends the raw key in that header instead and omits Authorization
	// (Anthropic). It is consumed when the create-connector resolver
	// builds the APIKeyConnection.
	APIKeyHeader string
	// APIKeyBasicAuth, when true, presents the API key as the username
	// of an HTTP Basic credential with an empty password instead of a
	// Bearer token — required by providers such as Cursor whose Admin
	// API documents `-u <key>:` Basic auth. Mutually exclusive with
	// APIKeyHeader. Consumed when the create-connector resolver builds
	// the APIKeyConnection.
	APIKeyBasicAuth bool
	// APIKeyAuthScheme selects a non-Bearer Authorization scheme for an
	// API-key connection: the key is sent as `Authorization: <scheme>
	// <key>` instead of `Authorization: Bearer <key>`. Required by
	// providers such as Okta whose API tokens use the `SSWS` scheme and
	// reject Bearer. Empty (the default) keeps the standard Bearer
	// scheme. Mutually exclusive with APIKeyHeader and APIKeyBasicAuth.
	// Consumed when the create-connector resolver builds the
	// APIKeyConnection.
	APIKeyAuthScheme string
	// APIKeyBasicAuthUserPass, when true, presents the API key as a complete
	// HTTP Basic credential whose `username:password` pair is already
	// encoded in the key (base64 of the verbatim string) — required by
	// providers such as ClickHouse Cloud (keyId:keySecret) and Langfuse
	// (publicKey:secretKey) whose Basic credential carries a real
	// password, unlike APIKeyBasicAuth's empty-password form. Mutually
	// exclusive with the other API-key auth modes. Consumed when the
	// create-connector resolver builds the APIKeyConnection.
	APIKeyBasicAuthUserPass bool
	// ManagedAPIKey marks a provider whose API key is supplied by Probo
	// from bootstrap config (a single, Probo-held credential shared across
	// all connections) rather than pasted per-connection by the customer.
	// The connection carries only the APIKeyExtraSettings (e.g. a Crisp
	// Website ID); the create-connector resolver injects the managed key
	// registered via (*Registry).SetManagedAPIKey. Such a provider stays
	// hidden from the driver catalog until the operator configures the key,
	// so it ships deactivated and activates with no code change. Orthogonal
	// to the APIKey*/SupportsAPIKey auth-mode flags, which still select how
	// the injected key is presented on the wire.
	ManagedAPIKey bool

	// RequiresManagedResourceID marks a ManagedAPIKey provider that also needs
	// a Probo-supplied resource ID (Crisp's plugin ID, registered via
	// (*Registry).SetManagedResourceID) before a connection can succeed. Such a
	// provider stays out of the driver catalog until BOTH the managed key and
	// the resource ID are configured, so the operator never sees it as
	// connectable while a connect attempt would fail at verify time. See
	// (*Registry).ManagedConnectorReady.
	RequiresManagedResourceID bool

	// BuildProbeURL derives a per-connector probe URL when the API host or
	// path depends on connector settings (e.g. a customer subdomain or
	// instance URL). Nil for providers with a static ProbeURL.
	BuildProbeURL func(*coredata.Connector) (string, error)
	// Probe runs a provider-specific connection check when a plain GET
	// against ProbeURL/BuildProbeURL is insufficient (e.g. GraphQL POST,
	// extra headers, or multi-host region probing). Takes precedence over
	// ProbeURL and BuildProbeURL when set.
	Probe func(context.Context, *http.Client, *coredata.Connector) error

	// Factory closures — wired by Stages 2 and 3.
	// NewDriver and NewNameResolver receive the registration's resolved
	// Endpoints as their last argument rather than closing over it. The
	// closures are written inside a &Registration{...} composite literal and
	// so cannot reference the value being built; capturing a copy declared
	// above the literal would compile and test green while silently ignoring
	// any later override of reg.Endpoints. Passing it at call time makes that
	// failure unrepresentable.
	NewDriver               func(context.Context, *http.Client, *coredata.Connector, *log.Logger, Endpoints) (drivers.Driver, error)
	NewNameResolver         func(context.Context, *http.Client, *coredata.Connector, *log.Logger, Endpoints) drivers.NameResolver
	SetOrganizationSettings func(*coredata.Connector, string) error
}

// ExtraSetting describes one extra per-provider settings field
// surfaced on ConnectorProviderInfo for the frontend to render.
type ExtraSetting struct {
	Key      string
	Label    string
	Required bool
}
