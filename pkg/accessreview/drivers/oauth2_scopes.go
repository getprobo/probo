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

package drivers

import "go.probo.inc/probo/pkg/coredata"

// providerOAuth2Scopes maps each access review provider to the OAuth2 scopes
// the corresponding driver requires to list user accounts. The map is the
// single source of truth for access-review OAuth2 scopes — surfaced via
// GraphQL so the frontend never hardcodes scope strings.
var providerOAuth2Scopes = map[coredata.ConnectorProvider][]string{
	coredata.ConnectorProviderHubSpot:  {"settings.users.read"},
	coredata.ConnectorProviderGitHub:   {"read:org"},
	coredata.ConnectorProviderSentry:   {"org:read", "member:read"},
	coredata.ConnectorProviderBrex:     {"openid", "offline_access"},
	coredata.ConnectorProviderDocuSign: {"signature"},
	coredata.ConnectorProviderLinear:   {"read"},
	coredata.ConnectorProviderGoogleWorkspace: {
		"https://www.googleapis.com/auth/admin.directory.user.readonly",
		"https://www.googleapis.com/auth/admin.directory.group.member.readonly",
	},
	// Notion and Intercom intentionally omitted: Notion uses extra-auth-params
	// instead of scopes, Intercom configures scopes at the app level.
}

// ProviderOAuth2Scopes returns the OAuth2 scopes the access review driver
// for the given provider needs. Returns nil for providers that do not need
// any scopes (Notion, Intercom) or for non-access-review providers.
func ProviderOAuth2Scopes(provider coredata.ConnectorProvider) []string {
	return providerOAuth2Scopes[provider]
}
