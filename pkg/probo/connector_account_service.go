// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	// ConnectorAccountService owns which vendor accounts a connector covers.
	//
	// It stores; it never reaches the vendor. Discovery is an outbound call
	// and lives with the session helper in pkg/accessreview, so there is
	// exactly one place a cloud session is opened.
	ConnectorAccountService struct {
		svc *Service
	}

	EnableConnectorAccountsRequest struct {
		OrganizationID gid.GID
		ConnectorID    gid.GID
		Accounts       []ConnectorAccountInput
	}

	// ConnectorAccountInput is one account to record. ExternalAccountID is
	// the identifier the vendor uses and is required: an account Probo
	// records on the user's instruction always has a name in the vendor's
	// own terms. The nameless implicit row is written by connector create,
	// not here.
	ConnectorAccountInput struct {
		ExternalAccountID string
		Name              string
	}

	DisableConnectorAccountRequest struct {
		ConnectorAccountID gid.GID
	}
)

func (r *EnableConnectorAccountsRequest) Validate() error {
	v := validator.New()

	v.Check(r.OrganizationID, "organization_id",
		validator.Required(),
		validator.GID(coredata.OrganizationEntityType),
	)
	v.Check(r.ConnectorID, "connector_id",
		validator.Required(),
		validator.GID(coredata.ConnectorEntityType),
	)
	v.Check(r.Accounts, "accounts", validator.Required())

	v.CheckEach(r.Accounts, "accounts", func(index int, item any) {
		account := item.(ConnectorAccountInput)

		v.Check(account.ExternalAccountID, fmt.Sprintf("accounts[%d].external_account_id", index),
			validator.Required(),
			validator.SafeTextNoNewLine(ConnectorAccountExternalIDMaxLength),
		)
		// Not SafeTextNoNewLine: that requires a value, and a vendor that
		// returns no display name is normal — Enable falls back to the
		// identifier rather than making the caller invent one.
		v.Check(account.Name, fmt.Sprintf("accounts[%d].name", index),
			validator.MaxLen(ConnectorAccountNameMaxLength),
			validator.NoHTML(),
			validator.PrintableText(),
			validator.NoNewLine(),
		)
	})

	return v.Error()
}

func (r *DisableConnectorAccountRequest) Validate() error {
	v := validator.New()

	v.Check(r.ConnectorAccountID, "connector_account_id",
		validator.Required(),
		validator.GID(coredata.ConnectorAccountEntityType),
	)

	return v.Error()
}

const (
	ConnectorAccountExternalIDMaxLength = 255
	ConnectorAccountNameMaxLength       = 255
)

// Enable records accounts under a connector, idempotently per vendor
// identifier: enabling one twice refreshes its name rather than inserting a
// second row.
//
// It never touches a source. Modules subscribe to an account; they do not own
// it, and enabling an account the user has not asked to review must not start
// reviewing it.
func (s *ConnectorAccountService) Enable(
	ctx context.Context,
	scope coredata.Scoper,
	req EnableConnectorAccountsRequest,
) ([]*coredata.ConnectorAccount, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	accounts := make([]*coredata.ConnectorAccount, 0, len(req.Accounts))

	err := s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			cnnctr := &coredata.Connector{}
			if err := cnnctr.LoadMetadataByID(ctx, tx, scope, req.ConnectorID); err != nil {
				return fmt.Errorf("cannot load connector: %w", err)
			}

			if cnnctr.OrganizationID != req.OrganizationID {
				return fmt.Errorf("cannot enable connector accounts: connector belongs to another organization")
			}

			accounts = accounts[:0]
			now := time.Now()

			for _, input := range req.Accounts {
				externalAccountID := input.ExternalAccountID

				name := input.Name
				if name == "" {
					name = externalAccountID
				}

				account := &coredata.ConnectorAccount{
					ID:                gid.New(scope.GetTenantID(), coredata.ConnectorAccountEntityType),
					OrganizationID:    cnnctr.OrganizationID,
					ConnectorID:       cnnctr.ID,
					ExternalAccountID: &externalAccountID,
					Name:              name,
					CreatedAt:         now,
					UpdatedAt:         now,
				}

				if err := account.Upsert(ctx, tx, scope); err != nil {
					return fmt.Errorf("cannot enable connector account: %w", err)
				}

				accounts = append(accounts, account)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// Disable removes one account. An account a source still reviews cannot be
// removed: the foreign key refuses and the caller sees ErrResourceInUse
// rather than a review losing the thing it reviewed.
func (s *ConnectorAccountService) Disable(
	ctx context.Context,
	scope coredata.Scoper,
	req DisableConnectorAccountRequest,
) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return s.svc.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			account := &coredata.ConnectorAccount{}
			if err := account.LoadByID(ctx, tx, scope, req.ConnectorAccountID); err != nil {
				return fmt.Errorf("cannot load connector account: %w", err)
			}

			if err := account.Delete(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot disable connector account: %w", err)
			}

			return nil
		},
	)
}

func (s *ConnectorAccountService) ListForConnectorID(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
	cursor *page.Cursor[coredata.ConnectorAccountOrderField],
) (*page.Page[*coredata.ConnectorAccount, coredata.ConnectorAccountOrderField], error) {
	var accounts coredata.ConnectorAccounts

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return accounts.LoadByConnectorID(ctx, conn, scope, connectorID, cursor)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list connector accounts: %w", err)
	}

	return page.NewPage(accounts, cursor), nil
}

func (s *ConnectorAccountService) CountForConnectorID(
	ctx context.Context,
	scope coredata.Scoper,
	connectorID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) (err error) {
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

func (s *ConnectorAccountService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	connectorAccountID gid.GID,
) (*coredata.ConnectorAccount, error) {
	account := &coredata.ConnectorAccount{}

	err := s.svc.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return account.LoadByID(ctx, conn, scope, connectorAccountID)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get connector account: %w", err)
	}

	return account, nil
}
