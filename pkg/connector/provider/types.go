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

// Registration is the per-provider metadata + factory bundle. Each
// provider returns one of these from a private constructor (e.g.
// slackRegistration) that NewBuiltinRegistry assembles into the
// runtime *Registry.
//
// Fields are grouped by concern: identity, endpoints, connect paths, the
// connection check, and the factory closures. Each connect path is one
// nil-able block, so which paths a provider offers is answered by which blocks
// are present rather than by a parallel set of booleans that could disagree
// with them.
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
	// EndpointOverrideUnsupported explains why a deployment cannot repoint this
	// provider, for providers that reach hosts declared outside Endpoints. An
	// override could move what Endpoints describes while those other calls kept
	// going to the real vendor, so the whole provider is refused rather than
	// partly moved. Empty for a provider whose every host comes from Endpoints.
	EndpointOverrideUnsupported string

	// OAuth2 is the OAuth2 connect path. Non-nil for exactly those providers
	// that offer one, even when every field inside takes its default, so that
	// a nil block reliably means "no OAuth2 path" — Register enforces the
	// agreement with Endpoints.Auth.
	OAuth2 *OAuth2Config

	// APIKey is the API-key connect path, covering both a customer-pasted key
	// and a Probo-held one. Nil for a provider with no API-key path.
	APIKey *APIKeyConfig

	// ClientCredentials is the OAuth2 client-credentials connect path. Nil for
	// a provider with no such path.
	ClientCredentials *ClientCredentialsConfig

	// WorkloadIdentity is the connect path for a provider reached by federated
	// workload identity. Nil for every provider that uses a stored credential.
	WorkloadIdentity *WorkloadIdentityConfig

	// BuildProbeURL derives a per-connector probe URL when the API host or
	// path depends on connector settings (e.g. a customer subdomain or
	// instance URL). Nil for providers with a static ProbeURL.
	//
	// Like NewDriver, it receives the registration's resolved Endpoints
	// rather than closing over them: a builder whose host IS constant (Neon,
	// Render, Qovery — only the path varies per connector) must compose from
	// ep.APIBase, or an APIBase override would move the driver while the
	// connection check kept hitting the real provider. Register can only
	// enforce that agreement on the static Endpoints.Probe, so passing the
	// Endpoints in is what extends the guarantee to the built URLs.
	BuildProbeURL func(*coredata.Connector, Endpoints) (string, error)
	// Probe runs a provider-specific connection check when a plain GET
	// against ProbeURL/BuildProbeURL is insufficient (e.g. GraphQL POST,
	// extra headers, or multi-host region probing). Takes precedence over
	// ProbeURL and BuildProbeURL when set. It receives the registration's
	// resolved Endpoints for the same reason BuildProbeURL does.
	Probe func(context.Context, *http.Client, *coredata.Connector, Endpoints) error
	// ClassifyRejection refines a 403 by reading the provider's own
	// explanation of it: it reports whether the provider accepted the
	// credential and refused the operation (true, the customer fixes the plan
	// or the permissions) or refused the credential itself (false, the
	// customer fixes the key). Nil for a provider that does not explain
	// itself, which leaves 403 meaning refused-operation.
	//
	// It receives the response body and must map it onto that one bit:
	// provider text is never carried any further than this closure.
	ClassifyRejection func(body []byte) bool

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

// APIKeyAuthMode selects how an API key is presented on outbound requests. The
// modes are mutually exclusive by construction, which is what the four
// independent booleans and strings this replaces could only achieve by counting
// them at startup.
type APIKeyAuthMode string

const (
	// APIKeyAuthBearer sends `Authorization: Bearer <key>`. The default, used by
	// every OAuth-style and standard API-key connector.
	APIKeyAuthBearer APIKeyAuthMode = ""

	// APIKeyAuthHeader sends the raw key in the header named by
	// APIKeyAuth.Name and omits Authorization entirely — required by providers
	// such as Anthropic that reject Bearer auth and return 400 when both
	// x-api-key and Authorization are present.
	APIKeyAuthHeader APIKeyAuthMode = "header"

	// APIKeyAuthBasic presents the key as the username of an HTTP Basic
	// credential with an empty password — required by providers such as Cursor
	// whose Admin API documents `-u <key>:` and rejects Bearer.
	APIKeyAuthBasic APIKeyAuthMode = "basic"

	// APIKeyAuthBasicUserPass presents the key as a complete HTTP Basic
	// credential whose `username:password` pair is already encoded in it —
	// required by ClickHouse Cloud (keyId:keySecret) and Langfuse
	// (publicKey:secretKey), whose credential carries a real password that
	// APIKeyAuthBasic's empty-password form cannot express.
	APIKeyAuthBasicUserPass APIKeyAuthMode = "basic-user-pass"

	// APIKeyAuthScheme sends `Authorization: <Name> <key>` — required by
	// providers such as Okta whose API tokens use the SSWS scheme.
	APIKeyAuthScheme APIKeyAuthMode = "scheme"
)

// APIKeyAuth is how a provider's API key goes on the wire. It is consumed by
// (*Registry).NewAPIKeyConnection, which is the single place the mode is
// translated into the connection's presentation fields.
type APIKeyAuth struct {
	Mode APIKeyAuthMode

	// Name is the request header name for APIKeyAuthHeader and the
	// Authorization scheme for APIKeyAuthScheme. The two never coexist, so one
	// field serves both. Empty and unused for every other mode; Register
	// rejects a Name that would be ignored or a mode that needs one and lacks
	// it.
	Name string
}

// APIKeyConfig is the API-key connect path of a provider.
type APIKeyConfig struct {
	// Auth is how the key is presented on the wire.
	Auth APIKeyAuth

	// ExtraSettings declares the per-provider settings fields the console's
	// API-key connect dialog renders and submits, in render order. It covers a
	// Managed provider too (Crisp): the customer supplies the settings, Probo
	// supplies the key.
	ExtraSettings []ExtraSetting

	// Managed is non-nil for a provider whose key Probo supplies from bootstrap
	// config (a single, Probo-held credential shared across all connections)
	// rather than one the customer pastes per connection. The connection then
	// carries only ExtraSettings (e.g. a Crisp Website ID); the
	// create-connector resolver injects the key registered via
	// (*Registry).SetManagedAPIKey. Such a provider stays hidden from the
	// driver catalog until the operator configures the key, so it ships
	// deactivated and activates with no code change.
	//
	// Orthogonal to Auth, which still selects how the injected key goes on the
	// wire.
	Managed *ManagedAPIKey
}

// ManagedAPIKey is the Probo-held variant of the API-key path. Nesting it under
// APIKeyConfig.Managed is what makes RequiresResourceID unreachable for a
// customer-supplied key, a pairing Register used to police at startup.
type ManagedAPIKey struct {
	// RequiresResourceID marks a provider that also needs a Probo-supplied
	// resource ID (Crisp's plugin ID, registered via
	// (*Registry).SetManagedResourceID) before a connection can succeed. Such a
	// provider stays out of the driver catalog until BOTH the key and the
	// resource ID are configured, so the operator never sees it as connectable
	// while a connect attempt would fail at verify time. See
	// (*Registry).ManagedConnectorReady.
	RequiresResourceID bool
}

// ClientCredentialsConfig is the OAuth2 client-credentials connect path.
type ClientCredentialsConfig struct {
	// ExtraSettings declares the settings fields the client-credentials connect
	// dialog renders and submits. It is independent of APIKeyConfig.ExtraSettings
	// because a different create resolver and a different driver sit behind each
	// path: 1Password needs SCIMBridgeURL on the API key (SCIM-bridge driver)
	// and AccountID + Region on client credentials (Users API driver). A setting
	// genuinely needed on both paths is declared in both lists.
	ExtraSettings []ExtraSetting
}

// OAuth2Config is the OAuth2 connect path of a provider: everything the
// authorization-code flow needs beyond the Auth and Token endpoints, which live
// in Endpoints because a deployment can override them.
//
// The whole block is consumed by (*Registry).ApplyOAuth2Defaults, which copies
// it onto a connector.OAuth2Connector at startup. It has no reader outside this
// package.
type OAuth2Config struct {
	// Scopes are the scopes the access-review driver needs to list accounts.
	// Nil for a provider that needs none (Notion, Intercom).
	Scopes []string

	// ExtraAuthParams are provider-specific query parameters added to the
	// authorization request. Copied per connector, never aliased.
	ExtraAuthParams map[string]string

	// TokenEndpointAuth selects how the client authenticates at the token
	// endpoint: "post-form" (default), "basic-form", "basic-json", or "none"
	// alongside PublicClient.
	TokenEndpointAuth string

	// SupportsIncrementalAuth allows an authorize request to ask for the union
	// of the connector's existing grant and the newly required scopes.
	SupportsIncrementalAuth bool

	// ExclusiveScopes marks a provider whose authorization server refuses any
	// scope outside Scopes, so a reconnect must request exactly that set rather
	// than the union with the connector's earlier grant. Asana is one: once its
	// app moved to Full Permissions, replaying a stored "users:read" fails the
	// whole authorize with forbidden_scopes.
	ExclusiveScopes bool

	// RequiresPKCE enables RFC 7636 PKCE (S256) on the authorization request
	// and replays the verifier on the token exchange. Default false; non-PKCE
	// providers are unaffected.
	RequiresPKCE bool

	// PublicClient marks a provider that authenticates as a public client (no
	// client_secret) via PKCE, using the Client ID Metadata Document (CIMD)
	// flow. probod auto-registers such providers with no operator credentials:
	// the client_id is the deployment's hosted CIMD URL (baseURL +
	// connector.CIMDMetadataPath) and the state token is signed with a
	// server-derived key. Set TokenEndpointAuth to "none" alongside this.
	PublicClient bool

	// BuildAuthURL derives the authorization URL from an operator-supplied
	// integration slug, for providers (e.g. Vercel) whose AuthURL embeds it as
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
	// callback (e.g. Zendesk's <subdomain>.zendesk.com). It MUST validate site.
	// A provider sets at most one of BuildTokenURLForDomain /
	// BuildTokenURLForSite. Nil otherwise.
	BuildTokenURLForSite func(site string) (string, error)
}

// WorkloadIdentityConfig is the connect path of a provider Probo reaches by
// minting a short-lived OIDC assertion (pkg/identityfederation) that the
// customer's cloud STS exchanges for temporary credentials.
//
// Grouping the three closures is what makes a half-configured cloud provider
// unrepresentable: a Probe with no driver, or a driver with no way to open a
// session, cannot be written down. Such a provider also needs no operator
// configuration at all — the customer grants access in their own cloud account
// — which is why the driver catalog admits it on the presence of this block.
type WorkloadIdentityConfig struct {
	// NewSession opens authenticated access to the cloud account this connector
	// points at, by exchanging an assertion the issuer mints for the connector's
	// organization. Which role, region and account to reach is per-provider
	// knowledge read from the connector's settings, which is why it lives here
	// rather than in a cross-cloud switch the access-review service would own.
	//
	// The framework calls it once and hands the session to NewDriver, Probe,
	// and NewNameResolver, mirroring how it hands one *http.Client to
	// Registration.NewDriver and Registration.Probe. Required.
	NewSession func(context.Context, *identityfederation.Issuer, *coredata.Connector) (cloud.Session, error)

	// NewDriver builds the access-review driver from a cloud session rather
	// than an *http.Client, because cloud SDK credentials sign requests the SDK
	// builds itself. Endpoints is absent because a cloud SDK resolves its own
	// endpoints from the region in the session.
	//
	// cloud.Session deliberately exposes only the cloud and the account: the
	// two SDKs diverge too far to share driver code, so a closure asserts the
	// concrete session type (*cloud/aws.Session) to reach its SDK config.
	// Required.
	NewDriver func(context.Context, cloud.Session, *coredata.Connector, *log.Logger) (drivers.Driver, error)

	// Probe is the connection check, replacing the HTTP Probe/BuildProbeURL
	// path. Nil means the check is skipped, matching an empty Endpoints.Probe.
	Probe func(context.Context, cloud.Session, *coredata.Connector) error

	// NewNameResolver builds the source-name resolver from a cloud session
	// rather than an *http.Client. Nil means the worker keeps the generic
	// provider display name, matching a nil Registration.NewNameResolver.
	NewNameResolver func(context.Context, cloud.Session, *coredata.Connector, *log.Logger) drivers.NameResolver

	// ExtraSettings declares the per-provider settings fields the console's
	// workload-identity connect dialog renders and submits, in render order.
	// Empty when the provider needs none beyond the grant in the customer's
	// own cloud account.
	ExtraSettings []ExtraSetting
}

// The three Supports* predicates below are derived from the presence of a
// connect path rather than declared beside it, so a flag can never disagree with
// the block it describes.

// SupportsWorkloadIdentity reports whether this provider is reached by federated
// workload identity rather than by a stored credential.
func (r *Registration) SupportsWorkloadIdentity() bool {
	return r.WorkloadIdentity != nil
}

// SupportsAPIKey reports whether the CUSTOMER supplies this provider's API key,
// which is what decides whether the console renders a key field. It is false for
// a managed provider, whose key Probo injects — so the old pairing this replaces
// (a managed key alongside a customer-supplied one) is now unrepresentable:
// both answers come from the same field, in opposite directions.
func (r *Registration) SupportsAPIKey() bool {
	return r.APIKey != nil && r.APIKey.Managed == nil
}

// SupportsClientCredentials reports whether this provider can be connected with
// an OAuth2 client-credentials grant.
func (r *Registration) SupportsClientCredentials() bool {
	return r.ClientCredentials != nil
}

// IsManagedAPIKey reports whether Probo, rather than the customer, supplies this
// provider's API key.
func (r *Registration) IsManagedAPIKey() bool {
	return r.APIKey != nil && r.APIKey.Managed != nil
}

// APIKeyExtraSettings returns the API-key dialog's settings fields, or nil when
// the provider has no API-key path.
func (r *Registration) APIKeyExtraSettings() []ExtraSetting {
	if r.APIKey == nil {
		return nil
	}

	return r.APIKey.ExtraSettings
}

// ClientCredentialsExtraSettings returns the client-credentials dialog's
// settings fields, or nil when the provider has no such path.
func (r *Registration) ClientCredentialsExtraSettings() []ExtraSetting {
	if r.ClientCredentials == nil {
		return nil
	}

	return r.ClientCredentials.ExtraSettings
}

// WorkloadIdentityExtraSettings returns the workload-identity dialog's
// settings fields, or nil when the provider has no such path.
func (r *Registration) WorkloadIdentityExtraSettings() []ExtraSetting {
	if r.WorkloadIdentity == nil {
		return nil
	}

	return r.WorkloadIdentity.ExtraSettings
}

// ExtraSetting describes one extra per-provider settings field
// surfaced on ConnectorProviderInfo for the frontend to render.
type ExtraSetting struct {
	Key      string
	Label    string
	Required bool
}
