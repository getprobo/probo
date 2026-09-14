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
	"errors"

	"go.probo.inc/probo/pkg/coredata"
)

type ClientBranding struct {
	Name      string
	ClientURL *string
	LogoURL   *string
}

func (s *Service) ClientBranding(ctx context.Context, clientIDRaw string) (*ClientBranding, error) {
	if clientIDRaw == "" {
		return nil, nil
	}

	client, err := s.resolveClient(ctx, nil, clientIDRaw)
	if err != nil {
		if _, ok := errors.AsType[*OAuth2Error](err); ok {
			return nil, nil
		}

		return nil, err
	}

	return ClientBrandingFromClient(client), nil
}

func ClientBrandingFromClient(client *coredata.OAuth2Client) *ClientBranding {
	if client == nil || client.ClientName == "" {
		return nil
	}

	branding := &ClientBranding{
		Name: client.ClientName,
	}

	// Only expose absolute http(s) URLs. Metadata may historically contain
	// non-web schemes; branding must not surface those as links or image src.
	if client.ClientURI != nil && client.ClientURI.IsHTTP() {
		clientURL := client.ClientURI.String()
		branding.ClientURL = &clientURL
	}

	if client.LogoURI != nil && client.LogoURI.IsHTTP() {
		logoURL := client.LogoURI.String()
		branding.LogoURL = &logoURL
	}

	return branding
}
