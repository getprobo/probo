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
	"errors"
	"fmt"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/task/sync/linear"
)

type linearAccount struct {
	client    *linear.Client
	connector *coredata.Connector
}

func (s *Service) linearAccountsForOrganization(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) ([]linearAccount, error) {
	var metas []*coredata.Connector

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

			metas = loaded

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot load Linear connectors: %w", err)
	}

	if len(metas) == 0 {
		return nil, ErrLinearNotConnected
	}

	accounts := make([]linearAccount, 0, len(metas))

	var reconnectErr error

	for _, meta := range metas {
		if err := meta.DecryptConnection(s.encryptionKey); err != nil {
			return nil, fmt.Errorf("cannot decrypt Linear connector: %w", err)
		}

		client, dbConnector, err := s.linearClientForConnector(ctx, scope, meta)
		if errors.Is(err, ErrLinearReconnectRequired) {
			if reconnectErr == nil {
				reconnectErr = err
			}

			continue
		}

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, linearAccount{
			client:    client,
			connector: dbConnector,
		})
	}

	if len(accounts) == 0 {
		if reconnectErr != nil {
			return nil, reconnectErr
		}

		return nil, ErrLinearNotConnected
	}

	return accounts, nil
}

func (s *Service) linearAccountForTeam(
	ctx context.Context,
	accounts []linearAccount,
	teamID string,
) (*linearAccount, error) {
	var listErr error

	for i := range accounts {
		found, err := accounts[i].client.HasTeam(ctx, teamID)
		if err != nil {
			listErr = err
			if s.logger != nil {
				s.logger.WarnCtx(
					ctx,
					"cannot resolve Linear team for connector",
					log.String("connector_id", accounts[i].connector.ID.String()),
					log.Error(err),
				)
			}

			continue
		}

		if found {
			return &accounts[i], nil
		}
	}

	if listErr != nil {
		return nil, fmt.Errorf("cannot resolve Linear team: %w", listErr)
	}

	return nil, ErrLinearTeamNotFound
}

func linearTeamExists(teams []linear.Team, teamID string) bool {
	for _, team := range teams {
		if team.ID == teamID {
			return true
		}
	}

	return false
}
