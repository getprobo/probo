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

package oauth2

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

func TestClientBrandingFromClient(t *testing.T) {
	t.Parallel()

	t.Run("exposes http and https client urls", func(t *testing.T) {
		t.Parallel()

		clientURL := uri.URI("https://example.com")
		logoURL := uri.URI("https://example.com/logo.png")

		branding := ClientBrandingFromClient(
			&coredata.OAuth2Client{
				ClientName: "Acme",
				ClientURI:  &clientURL,
				LogoURI:    &logoURL,
			},
		)

		require.NotNil(t, branding)
		assert.Equal(t, "Acme", branding.Name)
		assert.Equal(t, "https://example.com", *branding.ClientURL)
		assert.Equal(t, "https://example.com/logo.png", *branding.LogoURL)
	})

	t.Run("omits non-web client and logo urls", func(t *testing.T) {
		t.Parallel()

		clientURL := uri.URI("javascript://example.com/%0Aalert(1)")
		logoURL := uri.URI("data://example.com/image")

		branding := ClientBrandingFromClient(
			&coredata.OAuth2Client{
				ClientName: "Acme",
				ClientURI:  &clientURL,
				LogoURI:    &logoURL,
			},
		)

		require.NotNil(t, branding)
		assert.Equal(t, "Acme", branding.Name)
		assert.Nil(t, branding.ClientURL)
		assert.Nil(t, branding.LogoURL)
	})
}

func TestClientBranding_EmptyClientID(t *testing.T) {
	t.Parallel()

	svc := NewService(nil, nil, "", log.NewLogger())

	branding, err := svc.ClientBranding(context.Background(), "")
	require.NoError(t, err)
	assert.Nil(t, branding)
}

func TestClientBranding_InvalidClientID(t *testing.T) {
	t.Parallel()

	svc := NewService(nil, nil, "", log.NewLogger())

	branding, err := svc.ClientBranding(context.Background(), "not-a-cimd-client")
	require.NoError(t, err)
	assert.Nil(t, branding)
}

func TestClientBranding_CompliancePortalFlag(t *testing.T) {
	t.Parallel()

	const (
		portalClientID     = "https://portal.example.com/.well-known/oauth-client-metadata"
		thirdPartyClientID = "https://chatgpt.com/oauth/client.json"
	)

	t.Run(
		"skip consent marks compliance portal",
		func(t *testing.T) {
			t.Parallel()

			svc := NewService(
				nil,
				nil,
				"",
				log.NewLogger(),
				WithCIMDAllow(
					func(_ context.Context, clientIDURL string) (CIMDAllowance, error) {
						if clientIDURL == portalClientID {
							return CIMDAllowanceAllowedSkipConsent, nil
						}

						return CIMDAllowanceDenied, nil
					},
				),
			)

			assert.True(t, svc.isCompliancePortalClient(context.Background(), portalClientID))
		},
	)

	t.Run(
		"allowed cimd is not a compliance portal",
		func(t *testing.T) {
			t.Parallel()

			svc := NewService(
				nil,
				nil,
				"",
				log.NewLogger(),
				WithCIMDAllow(
					func(_ context.Context, _ string) (CIMDAllowance, error) {
						return CIMDAllowanceAllowed, nil
					},
				),
			)

			assert.False(t, svc.isCompliancePortalClient(context.Background(), thirdPartyClientID))
		},
	)
}
