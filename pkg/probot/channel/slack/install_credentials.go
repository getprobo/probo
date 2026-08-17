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

	"go.gearno.de/kit/log"
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

			return nil
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

	if installation.AccessTokenExpiresAt == nil ||
		installation.AccessTokenExpiresAt.After(time.Now().Add(tokenRefreshLeeway)) {
		return &installation, credentials, nil
	}

	if credentials.RefreshToken == nil {
		return nil, coredata.SlackbotInstallationCredentials{}, fmt.Errorf(
			"slack installation access token expired without refresh token",
		)
	}

	refreshed, err := s.refreshToken(ctx, *credentials.RefreshToken)
	if err != nil {
		return nil, coredata.SlackbotInstallationCredentials{}, fmt.Errorf(
			"cannot refresh Slack installation token: %w",
			err,
		)
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

	var persistErr error
	for range credentialPersistAttempts {
		persistErr = s.persistInstallationCredentials(
			ctx,
			scope,
			&installation,
			credentials,
		)
		if persistErr == nil {
			return &installation, credentials, nil
		}
	}

	s.logger.ErrorCtx(
		ctx,
		"cannot persist refreshed Slack credentials, disabling installation",
		log.Error(persistErr),
		log.String("organization_id", organizationID.String()),
	)

	if disableErr := s.disableInstallation(
		ctx,
		scope,
		&installation,
	); disableErr != nil {
		return nil, coredata.SlackbotInstallationCredentials{}, fmt.Errorf(
			"cannot disable Slack installation after credential persist failure: %w",
			disableErr,
		)
	}

	return nil, coredata.SlackbotInstallationCredentials{}, fmt.Errorf(
		"cannot persist refreshed Slack credentials: %w",
		persistErr,
	)
}

func (s *InstallationService) persistInstallationCredentials(
	ctx context.Context,
	scope coredata.Scoper,
	installation *coredata.SlackbotInstallation,
	credentials coredata.SlackbotInstallationCredentials,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var current coredata.SlackbotInstallation

			if err := current.LoadByOrganizationIDForUpdate(
				ctx,
				tx,
				scope,
				installation.OrganizationID,
			); err != nil {
				return fmt.Errorf("cannot load Slack installation for credential persist: %w", err)
			}

			if current.ID != installation.ID {
				return fmt.Errorf("cannot persist Slack credentials for replaced installation")
			}

			current.Scopes = installation.Scopes
			current.AccessTokenExpiresAt = installation.AccessTokenExpiresAt
			current.UpdatedAt = installation.UpdatedAt

			if err := current.UpdateCredentials(
				ctx,
				tx,
				scope,
				s.encryptionKey,
				credentials,
			); err != nil {
				return fmt.Errorf("cannot persist Slack installation credentials: %w", err)
			}

			*installation = current

			return nil
		},
	)
}

func (s *InstallationService) disableInstallation(
	ctx context.Context,
	scope coredata.Scoper,
	installation *coredata.SlackbotInstallation,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			var current coredata.SlackbotInstallation

			if err := current.LoadByOrganizationIDForUpdate(
				ctx,
				tx,
				scope,
				installation.OrganizationID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load Slack installation to disable: %w", err)
			}

			if current.ID != installation.ID {
				return nil
			}

			current.Status = coredata.SlackbotInstallationStatusDisabled
			current.UpdatedAt = time.Now()

			if err := current.UpdateStatus(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Slack installation status: %w", err)
			}

			return nil
		},
	)
}
