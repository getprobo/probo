// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package slack

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func (s *InstallationService) loadUsableCredentials(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*coredata.SlackbotInstallation, coredata.SlackbotInstallationCredentials, error) {
	var (
		installation coredata.SlackbotInstallation
		credentials  coredata.SlackbotInstallationCredentials
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := installation.LoadByOrganizationIDForUpdate(
				ctx,
				tx,
				scope,
				organizationID,
			); err != nil {
				return fmt.Errorf("cannot load Slack installation: %w", err)
			}

			if installation.Status != coredata.SlackbotInstallationStatusActive {
				return ErrSlackbotNotInstalled
			}

			var err error

			credentials, err = installation.DecryptCredentials(s.encryptionKey)
			if err != nil {
				return fmt.Errorf("cannot decrypt Slack installation credentials: %w", err)
			}

			if installation.AccessTokenExpiresAt == nil ||
				installation.AccessTokenExpiresAt.After(time.Now().Add(tokenRefreshLeeway)) {
				return nil
			}

			if credentials.RefreshToken == nil {
				return fmt.Errorf("slack installation access token expired without refresh token")
			}

			refreshed, err := s.refreshToken(ctx, *credentials.RefreshToken)
			if err != nil {
				return fmt.Errorf("cannot refresh Slack installation token: %w", err)
			}

			credentials.AccessToken = refreshed.AccessToken
			if refreshed.RefreshToken != "" {
				credentials.RefreshToken = new(refreshed.RefreshToken)
			}

			installation.Scopes = splitScopes(refreshed.Scope)
			if refreshed.ExpiresIn > 0 {
				installation.AccessTokenExpiresAt = new(
					time.Now().Add(time.Duration(refreshed.ExpiresIn) * time.Second),
				)
			}

			installation.UpdatedAt = time.Now()

			return installation.UpdateCredentials(
				ctx,
				tx,
				scope,
				s.encryptionKey,
				credentials,
			)
		},
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, coredata.SlackbotInstallationCredentials{}, ErrSlackbotNotInstalled
		}

		return nil, coredata.SlackbotInstallationCredentials{}, fmt.Errorf(
			"cannot load usable Slack credentials: %w",
			err,
		)
	}

	return &installation, credentials, nil
}
