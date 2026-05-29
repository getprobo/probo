// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package provider

import (
	"context"
	"net/http"

	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

// Registration is the per-provider metadata + factory bundle. Each
// provider returns one of these from a private constructor (e.g.
// slackRegistration) that NewBuiltinRegistry assembles into the
// runtime *Registry. Fields are grouped by concern: identity, OAuth2
// metadata, supported protocols, extra settings, and factory closures.
type Registration struct {
	// Identity.
	Provider    coredata.ConnectorProvider
	DisplayName string

	// OAuth2 metadata.
	AuthURL                 string
	TokenURL                string
	ExtraAuthParams         map[string]string
	TokenEndpointAuth       string // "post-form" (default), "basic-form", or "basic-json"
	SupportsIncrementalAuth bool
	OAuth2Scopes            []string
	ProbeURL                string
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

	// Protocol support / GraphQL surface.
	SupportsAPIKey            bool
	SupportsClientCredentials bool
	ExtraSettings             []ExtraSetting
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

	// Factory closures — wired by Stages 2 and 3.
	NewDriver               func(context.Context, *http.Client, *coredata.Connector, *log.Logger) (drivers.Driver, error)
	NewNameResolver         func(context.Context, *http.Client, *coredata.Connector, *log.Logger) drivers.NameResolver
	SetOrganizationSettings func(*coredata.Connector, string) error
}

// ExtraSetting describes one extra per-provider settings field
// surfaced on ConnectorProviderInfo for the frontend to render.
type ExtraSetting struct {
	Key      string
	Label    string
	Required bool
}
