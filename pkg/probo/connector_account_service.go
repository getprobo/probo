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

package probo

import (
	"context"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type (
	EnableConnectorAccount struct {
		ExternalAccountID string
		Name              string
	}
)

func (s *ConnectorService) GetAccount(
	ctx context.Context,
	scope coredata.Scoper,
	accountID gid.GID,
) (*coredata.ConnectorAccount, error) {
	account := &coredata.ConnectorAccount{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return account.LoadByID(ctx, conn, scope, accountID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector account: %w", err)
	}

	return account, nil
}

func (s *ConnectorService) ListAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	cursor *page.Cursor[coredata.ConnectorAccountOrderField],
) (*page.Page[*coredata.ConnectorAccount, coredata.ConnectorAccountOrderField], error) {
	var accounts coredata.ConnectorAccounts

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			cnnctr := &coredata.Connector{}
			if err := cnnctr.LoadMetadataByID(ctx, conn, scope, connectorID); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			return accounts.LoadByConnectorID(ctx, conn, scope, connectorID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list connector accounts: %w", err)
	}

	return page.NewPage(accounts, cursor), nil
}

func (s *ConnectorService) CountAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var err error

			accounts := coredata.ConnectorAccounts{}
			count, err = accounts.CountByConnectorID(ctx, conn, scope, connectorID)

			return err
		},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot count connector accounts: %w", err)
	}

	return count, nil
}

func (s *ConnectorService) EnableAccounts(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	accounts []EnableConnectorAccount,
) ([]*coredata.ConnectorAccount, error) {
	v := validator.New()
	v.Check(connectorID, "connector_id", validator.Required(), validator.GID(coredata.ConnectorEntityType))
	v.CheckEach(accounts, "accounts", func(index int, item any) {
		account := item.(EnableConnectorAccount)
		prefix := fmt.Sprintf("accounts[%d]", index)
		v.Check(
			account.ExternalAccountID,
			prefix+".external_account_id",
			validator.Required(),
			validator.SafeTextNoNewLine(TitleMaxLength),
		)

		if account.Name != "" {
			v.Check(account.Name, prefix+".name", validator.SafeTextNoNewLine(TitleMaxLength))
		}
	})

	externalAccountIDs := make([]string, len(accounts))
	for i, account := range accounts {
		externalAccountIDs[i] = account.ExternalAccountID
	}

	v.Check(externalAccountIDs, "accounts.external_account_id", validator.NoDuplicates())

	if err := v.Error(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	enabled := make([]*coredata.ConnectorAccount, 0, len(accounts))

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{}
			if err := cnnctr.LoadMetadataByID(ctx, tx, scope, connectorID); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			now := time.Now()

			for _, item := range accounts {
				name := item.Name
				if name == "" {
					name = item.ExternalAccountID
				}

				account := &coredata.ConnectorAccount{
					ID:                gid.New(scope.GetTenantID(), coredata.ConnectorAccountEntityType),
					OrganizationID:    cnnctr.OrganizationID,
					ConnectorID:       connectorID,
					ExternalAccountID: item.ExternalAccountID,
					Name:              name,
					CreatedAt:         now,
					UpdatedAt:         now,
				}

				if _, err := account.Upsert(ctx, tx, scope); err != nil {
					return err
				}

				enabled = append(enabled, account)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot enable connector accounts: %w", err)
	}

	return enabled, nil
}

func (s *ConnectorService) DisableAccount(
	ctx context.Context,
	scope coredata.Scoper,
	accountID gid.GID,
) error {
	v := validator.New()
	v.Check(accountID, "connector_account_id", validator.Required(), validator.GID(coredata.ConnectorAccountEntityType))

	if err := v.Error(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.ConnectorAccount{}
			if err := account.LoadByID(ctx, tx, scope, accountID); err != nil {
				return fmt.Errorf("cannot load connector account: %w", err)
			}

			sources := coredata.AccessReviewSources{}

			count, err := sources.CountByConnectorAccountID(ctx, tx, scope, accountID)
			if err != nil {
				return fmt.Errorf("cannot count access sources for connector account: %w", err)
			}

			if count > 0 {
				return coredata.ErrResourceInUse
			}

			return account.Delete(ctx, tx, scope)
		},
	)
	if err != nil {
		return fmt.Errorf("cannot disable connector account: %w", err)
	}

	return nil
}
