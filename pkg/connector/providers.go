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
		"LINEAR": {
			AuthURL:  "https://linear.app/oauth/authorize",
			TokenURL: "https://api.linear.app/oauth/token",
		},
	}
)

// ApplyProviderDefaults sets the redirect URI and applies static provider
// defaults (auth URL, token URL, extra params, token endpoint auth) onto
// an OAuth2Connector. Call this before registering the connector.
func ApplyProviderDefaults(provider string, redirectURI string, c *OAuth2Connector) {
	c.RedirectURI = redirectURI

	if def, ok := providerDefinitions[provider]; ok {
		c.AuthURL = def.AuthURL
		c.TokenURL = def.TokenURL
		c.ExtraAuthParams = def.ExtraAuthParams
		c.TokenEndpointAuth = def.TokenEndpointAuth
		c.SupportsIncrementalAuth = def.SupportsIncrementalAuth
	}
}
