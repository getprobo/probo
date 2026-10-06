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

package tasksync

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

func (s *Service) LinearDefaultTeam(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (*LinearTeam, error) {
	connectors, err := s.loadLinearSyncConnectors(ctx, scope, organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	for _, connector := range connectors {
		settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
		if err != nil {
			return nil, fmt.Errorf("cannot read Linear settings: %w", err)
		}

		if settings.DefaultTeamID == "" {
			continue
		}

		return &LinearTeam{
			ID:   settings.DefaultTeamID,
			Name: settings.DefaultTeamName,
			Key:  settings.DefaultTeamKey,
		}, nil
	}

	return nil, nil
}

// Includes connectors that need a reconnect, so an older default cannot win the next read.
func (s *Service) SetLinearDefaultTeam(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	teamID *string,
) error {
	connectors, err := s.loadLinearSyncConnectors(ctx, scope, organizationID)
	if err != nil {
		return fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	if len(connectors) == 0 {
		return ErrLinearNotConnected
	}

	var (
		chosenID gid.GID
		team     *linear.Team
	)

	if teamID != nil && strings.TrimSpace(*teamID) != "" {
		id := strings.TrimSpace(*teamID)

		accounts, err := s.linearAccountsFromConnectors(ctx, scope, connectors)
		if err != nil {
			return fmt.Errorf("cannot load Linear accounts: %w", err)
		}

		account, err := s.linearAccountForTeam(ctx, accounts, id)
		if err != nil {
			return fmt.Errorf("cannot resolve Linear team: %w", err)
		}

		team, err = account.client.TeamByID(ctx, id)
		if err != nil {
			return fmt.Errorf("cannot load Linear team: %w", err)
		}

		if team == nil {
			return ErrLinearTeamNotFound
		}

		chosenID = account.connector.ID
	} else {
		for _, connector := range connectors {
			if err := connector.DecryptConnection(s.encryptionKey); err != nil {
				return fmt.Errorf("cannot decrypt Linear connector: %w", err)
			}
		}
	}

	if err := applyLinearDefaultTeam(connectors, chosenID, team); err != nil {
		return fmt.Errorf("cannot apply Linear default team: %w", err)
	}

	err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		now := time.Now()

		for _, connector := range connectors {
			connector.UpdatedAt = now

			if err := connector.Update(ctx, tx, scope, s.encryptionKey); err != nil {
				return fmt.Errorf("cannot update Linear connector: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot update Linear default team: %w", err)
	}

	return nil
}

func applyLinearDefaultTeam(
	connectors []*coredata.Connector,
	chosenID gid.GID,
	team *linear.Team,
) error {
	for _, connector := range connectors {
		if team != nil && connector.ID == chosenID {
			if err := connector.SetSettings(&coredata.LinearSyncConnectorSettings{
				DefaultTeamID:   team.ID,
				DefaultTeamName: team.Name,
				DefaultTeamKey:  team.Key,
			}); err != nil {
				return fmt.Errorf("cannot store Linear default team: %w", err)
			}

			continue
		}

		connector.RawSettings = nil
	}

	return nil
}

func (s *Service) loadLinearSyncConnectors(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]*coredata.Connector, error) {
	var connectors []*coredata.Connector

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		provider := coredata.ConnectorProviderLinearSync
		filter := coredata.NewConnectorProviderFilter(&provider)

		loaded, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.ConnectorOrderField]{
				Field:     coredata.ConnectorOrderFieldCreatedAt,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.ConnectorOrderField]) ([]*coredata.Connector, error) {
				var batch coredata.Connectors
				if err := batch.LoadByOrganizationIDWithoutDecryptedConnection(
					ctx,
					conn,
					scope,
					organizationID,
					cursor,
					filter,
				); err != nil {
					return nil, err
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		connectors = loaded

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	return connectors, nil
}
