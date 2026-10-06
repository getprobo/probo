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

type LinearSettings struct {
	DefaultTeam     *LinearTeam
	ImportNewIssues bool
}

type LinearSyncSettingsUpdate struct {
	DefaultTeamID   **string
	ImportNewIssues *bool
}

func (s *Service) LinearSettings(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (LinearSettings, error) {
	connectors, err := s.loadLinearSyncConnectors(ctx, scope, organizationID)
	if err != nil {
		return LinearSettings{}, fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	for _, connector := range connectors {
		settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
		if err != nil {
			return LinearSettings{}, fmt.Errorf("cannot read Linear settings: %w", err)
		}

		if settings.DefaultTeamID == "" {
			continue
		}

		return LinearSettings{
			DefaultTeam: &LinearTeam{
				ID:   settings.DefaultTeamID,
				Name: settings.DefaultTeamName,
				Key:  settings.DefaultTeamKey,
			},
			ImportNewIssues: settings.ImportNewIssues,
		}, nil
	}

	return LinearSettings{}, nil
}

// Includes connectors that need a reconnect, so an older default cannot win the next read.
func (s *Service) UpdateLinearSettings(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	update LinearSyncSettingsUpdate,
) error {
	if update.DefaultTeamID == nil && update.ImportNewIssues == nil {
		return ErrLinearSettingsRequired
	}

	connectors, err := s.loadLinearSyncConnectors(ctx, scope, organizationID)
	if err != nil {
		return fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	if len(connectors) == 0 {
		return ErrLinearNotConnected
	}

	if update.DefaultTeamID != nil {
		return s.updateLinearDefaultTeam(ctx, scope, connectors, *update.DefaultTeamID, update.ImportNewIssues)
	}

	return s.updateLinearIssueImport(ctx, scope, connectors, *update.ImportNewIssues)
}

func (s *Service) updateLinearDefaultTeam(
	ctx context.Context,
	scope coredata.Scoper,
	connectors []*coredata.Connector,
	teamID *string,
	importNewIssues *bool,
) error {
	clearTeam := teamID == nil || strings.TrimSpace(*teamID) == ""
	if clearTeam {
		if importNewIssues != nil && *importNewIssues {
			return ErrLinearDefaultTeamRequired
		}

		if err := applyLinearDefaultTeam(connectors, gid.GID{}, nil, linearConnectorIdentity{}); err != nil {
			return fmt.Errorf("cannot apply Linear default team: %w", err)
		}

		return s.persistConnectorSettings(ctx, scope, connectors)
	}

	id := strings.TrimSpace(*teamID)

	accounts, err := s.linearAccountsFromConnectors(ctx, scope, connectors)
	if err != nil {
		return fmt.Errorf("cannot load Linear accounts: %w", err)
	}

	account, err := s.linearAccountForTeam(ctx, accounts, id)
	if err != nil {
		return fmt.Errorf("cannot resolve Linear team: %w", err)
	}

	team, err := account.client.TeamByID(ctx, id)
	if err != nil {
		return fmt.Errorf("cannot load Linear team: %w", err)
	}

	if team == nil {
		return ErrLinearTeamNotFound
	}

	linearOrganizationID, err := account.client.OrganizationID(ctx)
	if err != nil {
		return fmt.Errorf("cannot load Linear organization: %w", err)
	}

	appActorID, err := account.client.ViewerID(ctx)
	if err != nil {
		return fmt.Errorf("cannot load Linear viewer: %w", err)
	}

	if err := applyLinearDefaultTeam(connectors, account.connector.ID, team, linearConnectorIdentity{
		OrganizationID: linearOrganizationID,
		AppActorID:     appActorID,
	}); err != nil {
		return fmt.Errorf("cannot apply Linear default team: %w", err)
	}

	if importNewIssues != nil {
		if err := setImportNewIssues(account.connector, *importNewIssues); err != nil {
			return err
		}
	}

	return s.persistConnectorSettings(ctx, scope, connectors)
}

func applyLinearDefaultTeam(
	connectors []*coredata.Connector,
	chosenID gid.GID,
	team *linear.Team,
	identity linearConnectorIdentity,
) error {
	for _, connector := range connectors {
		if team != nil && connector.ID == chosenID {
			settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
			if err != nil {
				return fmt.Errorf("cannot read Linear settings: %w", err)
			}

			settings.DefaultTeamID = team.ID
			settings.DefaultTeamName = team.Name
			settings.DefaultTeamKey = team.Key
			settings.LinearOrganizationID = identity.OrganizationID
			settings.LinearAppActorID = identity.AppActorID

			if err := connector.SetSettings(&settings); err != nil {
				return fmt.Errorf("cannot store Linear default team: %w", err)
			}

			continue
		}

		connector.RawSettings = nil
	}

	return nil
}

type linearConnectorIdentity struct {
	OrganizationID string
	AppActorID     string
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

func (s *Service) updateLinearIssueImport(
	ctx context.Context,
	scope coredata.Scoper,
	connectors []*coredata.Connector,
	active bool,
) error {
	var chosen *coredata.Connector

	var settings coredata.LinearSyncConnectorSettings

	for _, connector := range connectors {
		current, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
		if err != nil {
			return fmt.Errorf("cannot read Linear settings: %w", err)
		}

		if current.DefaultTeamID == "" {
			continue
		}

		chosen = connector
		settings = current

		break
	}

	if chosen == nil {
		if !active {
			return nil
		}

		return ErrLinearDefaultTeamRequired
	}

	if active && (settings.LinearOrganizationID == "" || settings.LinearAppActorID == "") {
		identity, err := s.linearIdentityForTeam(ctx, scope, connectors, settings.DefaultTeamID)
		if err != nil {
			return err
		}

		settings.LinearOrganizationID = identity.OrganizationID
		settings.LinearAppActorID = identity.AppActorID

		if err := chosen.SetSettings(&settings); err != nil {
			return fmt.Errorf("cannot store Linear issue import: %w", err)
		}
	}

	if err := setImportNewIssues(chosen, active); err != nil {
		return err
	}

	return s.persistConnectorSettings(ctx, scope, []*coredata.Connector{chosen})
}

func setImportNewIssues(connector *coredata.Connector, active bool) error {
	settings, err := coredata.ConnectorSettings[coredata.LinearSyncConnectorSettings](connector)
	if err != nil {
		return fmt.Errorf("cannot read Linear settings: %w", err)
	}

	settings.ImportNewIssues = active

	if err := connector.SetSettings(&settings); err != nil {
		return fmt.Errorf("cannot store Linear issue import: %w", err)
	}

	return nil
}

func (s *Service) persistConnectorSettings(
	ctx context.Context,
	scope coredata.Scoper,
	connectors []*coredata.Connector,
) error {
	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		now := time.Now()

		for _, connector := range connectors {
			connector.UpdatedAt = now

			if err := connector.UpdateSettings(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot update Linear connector settings: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("cannot update Linear settings: %w", err)
	}

	return nil
}

func (s *Service) linearIdentityForTeam(
	ctx context.Context,
	scope coredata.Scoper,
	connectors []*coredata.Connector,
	teamID string,
) (linearConnectorIdentity, error) {
	accounts, err := s.linearAccountsFromConnectors(ctx, scope, connectors)
	if err != nil {
		return linearConnectorIdentity{}, fmt.Errorf("cannot load Linear accounts: %w", err)
	}

	account, err := s.linearAccountForTeam(ctx, accounts, teamID)
	if err != nil {
		return linearConnectorIdentity{}, fmt.Errorf("cannot resolve Linear team: %w", err)
	}

	organizationID, err := account.client.OrganizationID(ctx)
	if err != nil {
		return linearConnectorIdentity{}, fmt.Errorf("cannot load Linear organization: %w", err)
	}

	appActorID, err := account.client.ViewerID(ctx)
	if err != nil {
		return linearConnectorIdentity{}, fmt.Errorf("cannot load Linear viewer: %w", err)
	}

	return linearConnectorIdentity{
		OrganizationID: organizationID,
		AppActorID:     appActorID,
	}, nil
}
