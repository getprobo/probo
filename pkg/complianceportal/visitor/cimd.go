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

package visitor

import (
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

const (
	VisitorOAuthScope = "openid profile email"
	CIMDMetadataPath  = "/.well-known/oauth-client-metadata"
	OAuthCallbackPath = "/callback"
)

func CIMDClientIDURL(portalBaseURL string) (string, error) {
	return portalEndpointURL(portalBaseURL, CIMDMetadataPath)
}

func OAuthCallbackURL(portalBaseURL string) (string, error) {
	return portalEndpointURL(portalBaseURL, OAuthCallbackPath)
}

func PortalRootURL(rawURL string) (string, error) {
	return portalEndpointURL(rawURL, "")
}

// portalEndpointURL replaces the path on a portal base URL and clears
// query/fragment. Shared by CIMD, OAuth callback, and brand asset URLs.
func portalEndpointURL(portalBaseURL string, path string) (string, error) {
	parsed, err := url.Parse(portalBaseURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse portal URL: %w", err)
	}

	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func PortalBaseURLFromCIMDClientID(clientIDURL string) (string, error) {
	portalURL, err := PortalRootURL(clientIDURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse cimd client_id URL: %w", err)
	}

	return portalURL, nil
}

func BuildClientMetadataDocument(
	portal *coredata.CompliancePortal,
	portalBaseURL string,
) (oauth2.ClientMetadataDocument, error) {
	clientID, err := CIMDClientIDURL(portalBaseURL)
	if err != nil {
		return oauth2.ClientMetadataDocument{}, fmt.Errorf("cannot build cimd client_id URL: %w", err)
	}

	redirectURI, err := OAuthCallbackURL(portalBaseURL)
	if err != nil {
		return oauth2.ClientMetadataDocument{}, fmt.Errorf("cannot build oauth callback URL: %w", err)
	}

	portalRootURL, err := PortalRootURL(portalBaseURL)
	if err != nil {
		return oauth2.ClientMetadataDocument{}, fmt.Errorf("cannot build cimd client_uri URL: %w", err)
	}

	doc := oauth2.ClientMetadataDocument{
		ClientID:                clientID,
		ClientName:              portal.EntityName,
		ClientURI:               portalRootURL,
		RedirectURIs:            []string{redirectURI},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   VisitorOAuthScope,
	}

	if portal.LogoFileID != nil {
		logoURI, err := BrandLogoURL(portalBaseURL)
		if err == nil {
			doc.LogoURI = logoURI
		}
	}

	return doc, nil
}
