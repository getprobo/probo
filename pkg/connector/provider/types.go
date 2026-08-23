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

	"go.probo.inc/probo/pkg/cloud"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/identityfederation"
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

	// Identity is the host a provider's driver resolves its real data host
	// from, for providers that split identity from data. DocuSign is the
	// case: this endpoint's response carries the per-account `base_uri` that
	// every eSignature call then targets, so moving Identity moves the whole
	// connector while moving APIBase alone could not — DocuSign has no static
	// data root to move. Empty for every provider whose data host is either
	// static (APIBase) or supplied per connection.
	Identity string

	// APIBase is the data API root the driver joins paths onto, or — for a
	// GraphQL provider (Linear, Monday, Railway), which exposes a single
	// endpoint and nothing to join — that endpoint, used verbatim. Set it
	// ONLY for a compile-time-constant host. Leave it EMPTY when the host
	// comes from connector settings (Grafana, Metabase, Okta, SigNoz,
	// Zendesk, Datadog, Langfuse, Segment, PostHog) or is discovered at
	// runtime (DocuSign's per-account base_uri, PostHog's lazy region probe)
	// — the NewDriver closure resolves those, with
	// pkg/accessreview/drivers/posthog.go as the reference implementation.
	APIBase string
}

// Registration is everything the deployment knows about one provider before
// any connector exists: what to call it, which hosts it lives on, and which
// credential paths it offers. Each provider returns one from a private
// constructor (e.g. slackRegistration) that NewBuiltinRegistry assembles into
// the runtime *Registry.
//
// The credential paths are grouped into one pointer each rather than spread as
// flags, so "which paths does this provider offer" is answered by which pointers
// are set. That is also what reduces the mutual-exclusion checks in Register to
// the single rule that actually matters: federating into a customer's cloud
// rules out every path where Probo holds a credential.
//
// Nothing here describes what a caller does once connected. That is the
// consuming domain's business and lives in its own registry — see
// pkg/accessreview/drivers.
type Registration struct {
	Provider    coredata.ConnectorProvider
	DisplayName string
	// DocumentationURL is the public probo.com docs page for connecting this
	// provider as an access source. Empty for providers with no doc page yet;
	// surfaced (nullable) on ConnectorProviderInfo so the console renders a link.
	DocumentationURL string

	// Endpoints groups every host-bearing URL this provider owns.
	Endpoints Endpoints
	// EndpointOverrideUnsupported explains why a deployment cannot repoint this
	// provider, for providers that reach hosts declared outside Endpoints. An
	// override could move what Endpoints describes while those other calls kept
	// going to the real vendor, so the whole provider is refused rather than
	// partly moved. Empty for a provider whose every host comes from Endpoints.
	EndpointOverrideUnsupported string

	// OAuth2 describes the authorization-code path. Nil for a provider that
	// offers none, which is what tells the console there is no "Connect with
	// …" button to render.
	OAuth2 *OAuth2Spec
	// APIKey describes the path where a key is presented on every request,
	// whether the customer pastes it or Probo supplies it. Nil for a provider
	// that offers none.
	APIKey *APIKeySpec
	// ClientCredentials describes the OAuth2 client-credentials path, which is
	// independent of OAuth2 above: a provider can offer either, both, or
	// neither, and each needs its own settings because a different create
	// resolver sits behind it.
	ClientCredentials *ClientCredentialsSpec
	// WorkloadIdentity describes the path where Probo holds no credential at
	// all and federates in with OIDC. Mutually exclusive with the three above.
	WorkloadIdentity *WorkloadIdentitySpec

	// BuildProbeURL derives a per-connector probe URL when the API host or
	// path depends on connector settings (e.g. a customer subdomain or
	// instance URL). Nil for providers with a static Endpoints.Probe.
	//
	// It receives the registration's resolved Endpoints rather than closing
	// over them: a builder whose host IS constant (Neon, Render, Qovery — only
	// the path varies per connector) must compose from ep.APIBase, or an
	// APIBase override would move the driver while the connection check kept
	// hitting the real provider. Register can only enforce that agreement on
	// the static Endpoints.Probe, so passing the Endpoints in is what extends
	// the guarantee to the built URLs.
	BuildProbeURL func(*coredata.Connector, Endpoints) (string, error)
	// Probe runs a provider-specific connection check when a plain GET against
	// Endpoints.Probe or BuildProbeURL is insufficient (e.g. GraphQL POST,
	// extra headers, or multi-host region probing). Takes precedence over both
	// when set.
	Probe func(context.Context, *Handle) error

	// SetOrganizationSettings scopes a connector to one org/workspace/team by
	// writing the provider's own settings shape. It takes no credential, so it
	// stays here while listing the choices belongs to the consuming domain.
	// Nil for a provider whose scope is captured during the OAuth callback
	// (PagerDuty's subdomain, Vercel's team, Datadog's domain, Zendesk's
	// subdomain) and for one that has no scope at all.
	SetOrganizationSettings func(*coredata.Connector, string) error
}

// AcceptsCustomerAPIKey reports whether the customer pastes the key themselves,
// which is what decides whether the console renders a credential field. A
// managed provider offers the same dialog for its settings but no key field:
// Probo supplies the key and any pasted one would be discarded.
func (reg *Registration) AcceptsCustomerAPIKey() bool {
	return reg.APIKey != nil && !reg.APIKey.Managed
}

// OAuth2Spec is the authorization-code metadata a deployment's OAuth app needs
// on top of its own client ID and secret.
type OAuth2Spec struct {
	// Scopes are the scopes a review of this provider needs. Empty for a
	// provider that needs none (Notion, Intercom).
	Scopes []string
	// ExclusiveScopes marks a provider whose authorization server refuses any
	// scope outside Scopes, so a reconnect must request exactly that set rather
	// than the union with the connector's earlier grant. Asana is one: once its
	// app moved to Full Permissions, replaying a stored "users:read" fails the
	// whole authorize with forbidden_scopes.
	ExclusiveScopes bool
	// SupportsIncrementalAuth allows an authorize request to ask the provider
	// to keep scopes it already granted.
	SupportsIncrementalAuth bool
	// RequiresPKCE enables RFC 7636 PKCE (S256) on the authorization request
	// and replays the verifier on the token exchange.
	RequiresPKCE bool
	// PublicClient marks a provider that authenticates as a public client (no
	// client_secret) via PKCE, using the Client ID Metadata Document (CIMD)
	// flow. probod auto-registers such a provider with no operator
	// credentials: the client_id is the deployment's hosted CIMD URL (baseURL
	// + connector.CIMDMetadataPath) and the state token is signed with a
	// server-derived key. Set TokenEndpointAuth to "none" alongside this.
	PublicClient bool
	// TokenEndpointAuth is "post-form" (default), "basic-form", "basic-json",
	// or "none".
	TokenEndpointAuth string
	// ExtraAuthParams are provider-specific query parameters the authorize
	// request must carry.
	ExtraAuthParams map[string]string
	// BuildAuthURL derives the authorization URL from an operator-supplied
	// integration slug, for providers (e.g. Vercel) whose auth URL embeds it as
	// a path segment. It must construct the URL with net/url and escape the
	// slug. Nil for providers with a fully static Endpoints.Auth.
	BuildAuthURL func(slug string) (string, error)
	// BuildAuthURLForSite builds the authorize URL for a per-customer site
	// supplied at initiate time (multi-site providers, e.g. Datadog). It MUST
	// validate site against a fixed allow-list and construct the URL with
	// net/url. Nil for single-site providers.
	BuildAuthURLForSite func(site string) (string, error)
	// BuildTokenURLForDomain builds the token endpoint URL from the API domain
	// the provider returns on the OAuth callback (multi-site providers, e.g.
	// Datadog). It MUST validate domain. Nil otherwise.
	BuildTokenURLForDomain func(domain string) (string, error)
	// BuildTokenURLForSite builds the token endpoint URL from the per-customer
	// site/subdomain carried in the signed OAuth state, for multi-site
	// providers whose token host the provider does NOT echo back on the
	// callback (e.g. Zendesk's <subdomain>.zendesk.com). A provider sets at
	// most one of BuildTokenURLForDomain / BuildTokenURLForSite. Nil otherwise.
	BuildTokenURLForSite func(site string) (string, error)
}

// APIKeyPresentation selects how an API key is put on the wire. The zero value
// sends `Authorization: Bearer <key>`, which is what most providers want.
//
// It is one field rather than the four booleans it replaces because the
// presentations are alternatives, not options: with flags, setting two produced
// a silent winner that Register had to count its way out of.
type APIKeyPresentation string

const (
	// APIKeyBearer sends `Authorization: Bearer <key>`.
	APIKeyBearer APIKeyPresentation = ""
	// APIKeyBasic sends the key as a Basic username with an empty password,
	// which is what a provider documenting `-u <key>:` wants (Cursor).
	APIKeyBasic APIKeyPresentation = "BASIC"
	// APIKeyBasicUserPass sends the key as a complete Basic credential whose
	// `username:password` pair is already inside it (ClickHouse Cloud's
	// keyId:keySecret, Langfuse's publicKey:secretKey).
	APIKeyBasicUserPass APIKeyPresentation = "BASIC_USER_PASS"
	// APIKeyCustomHeader sends the raw key in the header named by Name, and no
	// Authorization header at all (Anthropic's x-api-key).
	APIKeyCustomHeader APIKeyPresentation = "CUSTOM_HEADER"
	// APIKeyCustomScheme sends `Authorization: <Name> <key>`, for a provider
	// whose tokens use a scheme of their own and reject Bearer (Okta's SSWS).
	APIKeyCustomScheme APIKeyPresentation = "CUSTOM_SCHEME"
)

// APIKeySpec is the path where a key travels on every request.
type APIKeySpec struct {
	Presentation APIKeyPresentation
	// Name is the header for APIKeyCustomHeader and the Authorization scheme
	// for APIKeyCustomScheme. Empty for every other presentation, which
	// consumes no name.
	Name string
	// Managed marks a provider whose key Probo supplies from bootstrap config
	// (a single credential shared across all connections) rather than the
	// customer pasting one. The connection then carries only ExtraSettings
	// (e.g. a Crisp Website ID); the create resolver injects the key
	// registered via (*Registry).SetManagedAPIKey. Such a provider stays out
	// of the catalog until the operator configures it, so it ships deactivated
	// and activates with no code change.
	Managed bool
	// RequiresResourceID marks a Managed provider that also needs a
	// Probo-supplied resource ID (Crisp's plugin ID, registered via
	// (*Registry).SetManagedResourceID) before a connection can succeed. Such
	// a provider stays out of the catalog until BOTH are configured, so the
	// operator never sees it as connectable while a connect attempt would fail
	// at verify time. See (*Registry).ManagedConnectorReady.
	RequiresResourceID bool
	// ExtraSettings declares the per-provider settings fields the console's
	// API-key connect dialog renders and submits, in render order. It covers a
	// Managed provider too (Crisp): the customer supplies the settings, Probo
	// supplies the key.
	ExtraSettings []ExtraSetting
}

// ClientCredentialsSpec is the OAuth2 client-credentials path.
//
// Its settings are separate from APIKeySpec.ExtraSettings because a different
// create resolver and a different driver sit behind each: 1Password needs
// SCIMBridgeURL on the API key (SCIM-bridge driver) and AccountID + Region on
// client credentials (Users API driver). A setting genuinely needed on both
// paths is declared in both.
type ClientCredentialsSpec struct {
	ExtraSettings []ExtraSetting
}

// WorkloadIdentitySpec is the path where Probo stores no credential and mints
// one per use by federating into the customer's cloud with OIDC.
type WorkloadIdentitySpec struct {
	// NewSession performs the credential exchange. It has no counterpart on the
	// other paths because those carry their credential in the connection row,
	// whereas this one holds none.
	//
	// It takes the issuer at call time because the closure sits in a literal
	// assembled at startup, before any issuer exists.
	NewSession func(
		ctx context.Context,
		issuer *identityfederation.Issuer,
		conn *coredata.Connector,
	) (cloud.Session, error)
}

// ExtraSetting describes one extra per-provider settings field
// surfaced on ConnectorProviderInfo for the frontend to render.
type ExtraSetting struct {
	Key      string
	Label    string
	Required bool
}
