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

package provider

import (
	"context"
	"fmt"
	"net/http"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/coredata"
)

func bitbucketRegistration() *Registration {
	// Bitbucket scopes are pinned on the OAuth consumer at registration
	// time (`account` for workspace membership). They are not passed in
	// the authorize URL.
	return &Registration{
		Provider:    coredata.ConnectorProviderBitbucket,
		DisplayName: "Bitbucket",
		AuthURL:     "https://bitbucket.org/site/oauth2/authorize",
		TokenURL:    "https://bitbucket.org/site/oauth2/access_token",
		ProbeURL:    "https://api.bitbucket.org/2.0/user",
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.BitbucketConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read bitbucket connector settings: %w", err)
			}

			if s.Workspace == "" {
				return nil, fmt.Errorf("cannot create bitbucket driver: workspace is required")
			}

			return drivers.NewBitbucketDriver(c, s.Workspace), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.BitbucketConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read bitbucket connector settings", log.Error(err))
				return nil
			}

			return drivers.NewBitbucketNameResolver(c, s.Workspace)
		},
		SetOrganizationSettings: func(c *coredata.Connector, workspace string) error {
			return c.SetSettings(&coredata.BitbucketConnectorSettings{Workspace: workspace})
		},
	}
}
