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

package complianceportal

import (
	"fmt"
	"net/url"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
)

const (
	VisitorOAuthScope = "openid profile email"
	CIMDMetadataPath  = "/.well-known/oauth-client-metadata"
	OAuthCallbackPath = "/callback"
)

func CIMDClientIDURL(portalBaseURL string) (string, error) {
	parsed, err := url.Parse(portalBaseURL)
	if err != nil {
		return "", err
	}

	parsed.Path = CIMDMetadataPath
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func OAuthCallbackURL(portalBaseURL string) (string, error) {
	parsed, err := url.Parse(portalBaseURL)
	if err != nil {
		return "", err
	}

	parsed.Path = OAuthCallbackPath
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func PortalBaseURLFromCIMDClientID(clientIDURL string) (string, error) {
	parsed, err := url.Parse(clientIDURL)
	if err != nil {
		return "", fmt.Errorf("cannot parse cimd client_id URL: %w", err)
	}

	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}

func BuildClientMetadataDocument(
	portal *coredata.TrustCenter,
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

	doc := oauth2.ClientMetadataDocument{
		ClientID:                clientID,
		ClientName:              portal.Title,
		RedirectURIs:            []string{redirectURI},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
		Scope:                   VisitorOAuthScope,
	}

	if portal.WebsiteURL != nil && *portal.WebsiteURL != "" {
		doc.ClientURI = *portal.WebsiteURL
	} else {
		doc.ClientURI = portalBaseURL
	}

	if portal.LogoFileID != nil {
		parsedBaseURL, err := baseurl.Parse(portalBaseURL)
		if err == nil {
			logoURI, err := parsedBaseURL.
				WithPath("/api/files/v1/public/" + portal.LogoFileID.String()).
				String()
			if err == nil {
				doc.LogoURI = logoURI
			}
		}
	}

	return doc, nil
}
