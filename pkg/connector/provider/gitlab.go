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

func gitlabRegistration() *Registration {
	return &Registration{
		Provider:     coredata.ConnectorProviderGitLab,
		DisplayName:  "GitLab",
		AuthURL:      "https://gitlab.com/oauth/authorize",
		TokenURL:     "https://gitlab.com/oauth/token",
		ProbeURL:     "https://gitlab.com/api/v4/user",
		OAuth2Scopes: []string{"read_api"},
		NewDriver: func(_ context.Context, c *http.Client, conn *coredata.Connector, _ *log.Logger) (drivers.Driver, error) {
			s, err := coredata.ConnectorSettings[coredata.GitLabConnectorSettings](conn)
			if err != nil {
				return nil, fmt.Errorf("cannot read gitlab connector settings: %w", err)
			}

			if s.GroupID == "" {
				return nil, fmt.Errorf("cannot create gitlab driver: group_id is required")
			}

			return drivers.NewGitLabDriver(c, s.GroupID), nil
		},
		NewNameResolver: func(ctx context.Context, c *http.Client, conn *coredata.Connector, logger *log.Logger) drivers.NameResolver {
			s, err := coredata.ConnectorSettings[coredata.GitLabConnectorSettings](conn)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot read gitlab connector settings", log.Error(err))
				return nil
			}

			return drivers.NewGitLabNameResolver(c, s.GroupID)
		},
		SetOrganizationSettings: func(c *coredata.Connector, groupID string) error {
			return c.SetSettings(&coredata.GitLabConnectorSettings{GroupID: groupID})
		},
	}
}
