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

package connector

import (
	"maps"
	"strings"

	"go.gearno.de/kit/httpclient"
)

// CallbackPath is the HTTP path for the OAuth2 callback endpoint.
const CallbackPath = "/api/console/v1/connectors/complete"

// providerDefinition holds the static OAuth2 properties for a provider.
// These are intrinsic to the provider and do not vary between deployments.
// Scopes are not part of this — they are passed by the caller at
// initiate time via InitiateOptions, since the same provider may be used
// in multiple contexts requiring different scope sets.
type providerDefinition struct {
	AuthURL                 string
	TokenURL                string
	ExtraAuthParams         map[string]string
	TokenEndpointAuth       string // "post-form" (default), "basic-form", or "basic-json"
	SupportsIncrementalAuth bool
	// RequiresPKCE enables RFC 7636 PKCE (S256) on the authorization
	// request and replays the verifier on the token exchange. Default
	// false; existing providers are unaffected.
	RequiresPKCE bool
	// TokenExtraParams are merged into the token-exchange request body
	// (form-encoded for "post-form"/"basic-form", JSON for "basic-json").
	// Used by providers like Lever that require an `audience` parameter.
	TokenExtraParams map[string]string
}

// providerDefinitions maps provider names to their static OAuth2 definitions.
// Only ClientID and ClientSecret come from deployment config.
var (
	providerDefinitions = map[string]providerDefinition{
		"SLACK": {
			AuthURL:  "https://slack.com/oauth/v2/authorize",
			TokenURL: "https://slack.com/api/oauth.v2.access",
		},
		"HUBSPOT": {
			AuthURL:  "https://app.hubspot.com/oauth/authorize",
			TokenURL: "https://api.hubapi.com/oauth/v1/token",
		},
		"DOCUSIGN": {
			AuthURL:           "https://account.docusign.com/oauth/auth",
			TokenURL:          "https://account.docusign.com/oauth/token",
			TokenEndpointAuth: "basic-form",
		},
		"NOTION": {
			AuthURL:           "https://api.notion.com/v1/oauth/authorize",
			TokenURL:          "https://api.notion.com/v1/oauth/token",
			ExtraAuthParams:   map[string]string{"owner": "user"},
			TokenEndpointAuth: "basic-json",
		},
		"GITHUB": {
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
		"SENTRY": {
			AuthURL:  "https://sentry.io/oauth/authorize/",
			TokenURL: "https://sentry.io/oauth/token/",
		},
		"INTERCOM": {
			AuthURL:  "https://app.intercom.com/oauth",
			TokenURL: "https://api.intercom.io/auth/eagle/token",
		},
		"BREX": {
			AuthURL:  "https://accounts-api.brex.com/oauth2/default/v1/authorize",
			TokenURL: "https://accounts-api.brex.com/oauth2/default/v1/token",
		},
		"GOOGLE_WORKSPACE": {
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			ExtraAuthParams: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
			SupportsIncrementalAuth: true,
		},
		"MICROSOFT_365": {
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			ExtraAuthParams: map[string]string{
				"prompt": "consent",
			},
		},
		"LINEAR": {
			AuthURL:  "https://linear.app/oauth/authorize",
			TokenURL: "https://api.linear.app/oauth/token",
		},
		"GITLAB": {
			AuthURL:  "https://gitlab.com/oauth/authorize",
			TokenURL: "https://gitlab.com/oauth/token",
		},
		// Bitbucket scopes are pinned on the OAuth consumer at registration
		// time (`account` for workspace membership). They are not passed in
		// the authorize URL and not configured here.
		"BITBUCKET": {
			AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
			TokenURL: "https://bitbucket.org/site/oauth2/access_token",
		},
		"HEROKU": {
			AuthURL:  "https://id.heroku.com/oauth/authorize",
			TokenURL: "https://id.heroku.com/oauth/token",
		},
		"PAGERDUTY": {
			AuthURL:      "https://identity.pagerduty.com/oauth/authorize",
			TokenURL:     "https://identity.pagerduty.com/oauth/token",
			RequiresPKCE: true,
		},
		"ASANA": {
			AuthURL:  "https://app.asana.com/-/oauth_authorize",
			TokenURL: "https://app.asana.com/-/oauth_token",
		},
		"SNYK": {
			AuthURL:      "https://app.snyk.io/oauth2/authorize",
			TokenURL:     "https://api.snyk.io/oauth2/token",
			RequiresPKCE: true,
		},
		"NETLIFY": {
			AuthURL:  "https://app.netlify.com/authorize",
			TokenURL: "https://api.netlify.com/oauth/token",
		},
		"RAMP": {
			AuthURL:           "https://app.ramp.com/v1/authorize",
			TokenURL:          "https://api.ramp.com/developer/v1/token",
			TokenEndpointAuth: "basic-form",
		},
		"CLICKUP": {
			AuthURL:  "https://app.clickup.com/api",
			TokenURL: "https://api.clickup.com/api/v2/oauth/token",
		},
		// Vercel uses a templated AuthURL: the operator supplies an
		// `integration-slug` config field which is resolved into the
		// "{integration_slug}" placeholder by ApplyProviderDefaults.
		// Vercel does not use OAuth scopes — capabilities are pinned on
		// the integration registration in the Vercel dashboard.
		"VERCEL": {
			AuthURL:  "https://vercel.com/integrations/{integration_slug}/new",
			TokenURL: "https://api.vercel.com/v2/oauth/access_token",
		},
		"MONDAY": {
			AuthURL:  "https://auth.monday.com/oauth2/authorize",
			TokenURL: "https://auth.monday.com/oauth2/token",
		},
		// Lever runs on Auth0: the `audience` parameter is required in
		// BOTH the authorize URL and the token-exchange POST body. The
		// trailing slash on the audience value is mandatory.
		"LEVER": {
			AuthURL:  "https://auth.lever.co/authorize",
			TokenURL: "https://auth.lever.co/oauth/token",
			ExtraAuthParams: map[string]string{
				"audience": "https://api.lever.co/v1/",
				"prompt":   "consent",
			},
			TokenExtraParams: map[string]string{
				"audience": "https://api.lever.co/v1/",
			},
		},
		// Deel: the token endpoint path is "/oauth2/tokens" (plural) —
		// Deel's docs are inconsistent on the singular vs plural form.
		// The API base host (api.letsdeel.com) differs from the auth host
		// (app.deel.com). Deel's token endpoint requires HTTP Basic auth
		// (base64(client_id:client_secret)); credentials placed in the
		// form body are rejected with 401 invalid basic credentials.
		"DEEL": {
			AuthURL:           "https://app.deel.com/oauth2/authorize",
			TokenURL:          "https://app.deel.com/oauth2/tokens",
			TokenEndpointAuth: "basic-form",
		},
	}
)

// ApplyProviderDefaults sets the redirect URI and applies static provider
// defaults (auth URL, token URL, extra params, token endpoint auth) onto
// an OAuth2Connector, and wires an SSRF-protected HTTP client for the
// token exchange request. Call this before registering the connector.
func ApplyProviderDefaults(provider string, redirectURI string, c *OAuth2Connector) {
	c.RedirectURI = redirectURI
	c.HTTPClient = httpclient.DefaultClient(httpclient.WithSSRFProtection())

	if def, ok := providerDefinitions[provider]; ok {
		c.AuthURL = def.AuthURL
		c.TokenURL = def.TokenURL
		c.TokenEndpointAuth = def.TokenEndpointAuth
		c.SupportsIncrementalAuth = def.SupportsIncrementalAuth
		c.RequiresPKCE = def.RequiresPKCE

		// Deep copy ExtraAuthParams and TokenExtraParams so per-connector
		// mutations (e.g. incremental auth, scope overrides) cannot alias
		// back into the shared providerDefinitions map.
		if len(def.ExtraAuthParams) > 0 {
			extra := make(map[string]string, len(def.ExtraAuthParams))
			maps.Copy(extra, def.ExtraAuthParams)
			c.ExtraAuthParams = extra
		}
		if len(def.TokenExtraParams) > 0 {
			tokenExtra := make(map[string]string, len(def.TokenExtraParams))
			maps.Copy(tokenExtra, def.TokenExtraParams)
			c.TokenExtraParams = tokenExtra
		}

		// Resolve operator-supplied placeholders in the static AuthURL
		// (for example Vercel's "{integration_slug}"). Providers without
		// placeholders are unaffected; the loop is a no-op when
		// AuthURLParams is empty.
		for k, v := range c.AuthURLParams {
			c.AuthURL = strings.ReplaceAll(c.AuthURL, "{"+k+"}", v)
		}
	}
}
