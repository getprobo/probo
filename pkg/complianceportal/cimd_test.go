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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCIMDClientIDURL(t *testing.T) {
	t.Parallel()

	clientID, err := CIMDClientIDURL("https://acme.example.com/overview")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/.well-known/oauth-client-metadata", clientID)
}

func TestOAuthCallbackURL(t *testing.T) {
	t.Parallel()

	callbackURL, err := OAuthCallbackURL("https://acme.example.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/callback", callbackURL)
}

func TestPortalBaseURLFromCIMDClientID(t *testing.T) {
	t.Parallel()

	baseURL, err := PortalBaseURLFromCIMDClientID(
		"https://acme.example.com/.well-known/oauth-client-metadata",
	)
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com", baseURL)
}

func TestBrandLogoURL(t *testing.T) {
	t.Parallel()

	logoURL, err := BrandLogoURL("https://acme.example.com/page")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/brand/logo", logoURL)

	darkLogoURL, err := BrandDarkLogoURL("https://acme.example.com/")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/brand/dark-logo", darkLogoURL)
}

func TestBuildClientMetadataDocument(t *testing.T) {
	t.Parallel()

	websiteURL := "https://www.acme.com"
	portal := &coredata.TrustCenter{
		Title:      "Acme Trust Center",
		WebsiteURL: &websiteURL,
	}

	doc, err := BuildClientMetadataDocument(
		portal,
		"https://acme.example.com/.well-known/oauth-client-metadata",
	)
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/.well-known/oauth-client-metadata", doc.ClientID)
	assert.Equal(t, "Acme Trust Center", doc.ClientName)
	assert.Equal(t, []string{"https://acme.example.com/callback"}, doc.RedirectURIs)
	assert.Equal(t, "https://acme.example.com", doc.ClientURI)
	assert.Equal(t, VisitorOAuthScope, doc.Scope)
	assert.Empty(t, doc.LogoURI)
}

func TestBuildClientMetadataDocument_LogoURIUsesBrandLogoEndpoint(t *testing.T) {
	t.Parallel()

	logoFileID := gid.MustParseGID("WR-qMrB5AAEAGQAAAZ9mIO8B8vDFQ-i3")
	portal := &coredata.TrustCenter{
		Title:      "Acme Trust Center",
		LogoFileID: &logoFileID,
	}

	doc, err := BuildClientMetadataDocument(portal, "https://acme.example.com")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.example.com/brand/logo", doc.LogoURI)
}
