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

package connect_v1

import (
	"context"

	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
)

func oauthClientBranding(
	ctx context.Context,
	r *Resolver,
	clientID string,
) (*types.OAuthClientBranding, error) {
	branding, err := r.iam.OAuth2ServerService.ClientBranding(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if branding == nil {
		return nil, nil
	}

	return oauthClientBrandingFromIAM(branding)
}

func oauthClientBrandingFromIAM(
	branding *oauth2.ClientBranding,
) (*types.OAuthClientBranding, error) {
	result := &types.OAuthClientBranding{
		Name:      branding.Name,
		ClientURL: branding.ClientURL,
	}

	if branding.LogoURL != nil {
		result.Logo = &types.File{
			DownloadURL: *branding.LogoURL,
		}
	}

	return result, nil
}
