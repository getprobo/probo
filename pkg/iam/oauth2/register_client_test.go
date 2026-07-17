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

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

func TestRegisterClient_RejectsNonWebURIs(t *testing.T) {
	t.Parallel()

	svc := NewService(nil, nil, "", log.NewLogger())

	t.Run(
		"non-web client_uri",
		func(t *testing.T) {
			t.Parallel()

			clientURI := uri.URI("javascript://example.com/%0Aalert(1)")

			_, _, err := svc.RegisterClient(
				context.Background(),
				&RegisterClientRequest{
					ClientName:   "Acme",
					Visibility:   coredata.OAuth2ClientVisibilityPublic,
					RedirectURIs: []uri.URI{"https://example.com/callback"},
					ClientURI:    &clientURI,
				},
			)
			require.Error(t, err)
		},
	)

	t.Run(
		"non-web logo_uri",
		func(t *testing.T) {
			t.Parallel()

			logoURI := uri.URI("data://example.com/image")

			_, _, err := svc.RegisterClient(
				context.Background(),
				&RegisterClientRequest{
					ClientName:   "Acme",
					Visibility:   coredata.OAuth2ClientVisibilityPublic,
					RedirectURIs: []uri.URI{"https://example.com/callback"},
					LogoURI:      &logoURI,
				},
			)
			require.Error(t, err)
		},
	)
}
