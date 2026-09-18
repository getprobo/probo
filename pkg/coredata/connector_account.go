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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	// ConnectorAccount is one vendor account a connector credential covers.
	//
	// ExternalAccountID is the identifier the vendor uses — an AWS account
	// id, a GCP project number, an Azure subscription, a GitHub org login.
	// It is nil when the credential is bound to a single tenant the vendor
	// never names, which is most SaaS providers: opening a session with an
	// empty account id then reproduces the settings-implied behaviour.
	//
	// It is never the empty string. A blank would read as a slug we failed
	// to capture, and one empty string per connector is admitted by the
	// unique index, so it could never serve as a general placeholder anyway.
	ConnectorAccount struct {
		ID                gid.GID   `db:"id"`
		OrganizationID    gid.GID   `db:"organization_id"`
		ConnectorID       gid.GID   `db:"connector_id"`
		ExternalAccountID *string   `db:"external_account_id"`
		Name              string    `db:"name"`
		CreatedAt         time.Time `db:"created_at"`
		UpdatedAt         time.Time `db:"updated_at"`
	}

	ConnectorAccounts []*ConnectorAccount
)

func (ca ConnectorAccount) CursorKey(orderBy ConnectorAccountOrderField) page.CursorKey {
	switch orderBy {
	case ConnectorAccountOrderFieldCreatedAt:
		return page.NewCursorKey(ca.ID, ca.CreatedAt)
	case ConnectorAccountOrderFieldExternalAccountID:
		return page.NewCursorKey(ca.ID, ca.ExternalAccountID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (ca *ConnectorAccount) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM connector_accounts WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

// Upsert records an account the vendor names, refreshing its display name
// when it is already known. The generated ID is preserved across upserts
// because it is not part of the conflict target, so a source already
// pointing at the account keeps pointing at the same row.
//
// It requires a non-nil ExternalAccountID: ON CONFLICT never matches a NULL,
// so the implicit row has its own arbiter in UpsertImplicit.
func (ca *ConnectorAccount) Upsert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	if ca.ExternalAccountID == nil {
		return fmt.Errorf("cannot upsert connector account: no external account id, use UpsertImplicit")
	}

	return ca.upsert(ctx, conn, scope, "(connector_id, external_account_id)")
}

// UpsertImplicit records the single account of a credential the vendor never
// names, arbitrating on the partial unique index instead of the composite
// one so a second call updates the first row rather than inserting beside it.
func (ca *ConnectorAccount) UpsertImplicit(ctx context.Context, conn pg.Tx, scope Scoper) error {
	if ca.ExternalAccountID != nil {
		return fmt.Errorf("cannot upsert implicit connector account: it has an external account id, use Upsert")
	}

	return ca.upsert(ctx, conn, scope, "(connector_id) WHERE external_account_id IS NULL")
}

func (ca *ConnectorAccount) upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	conflictTarget string,
) error {
	q := `
INSERT INTO connector_accounts (
    id,
    tenant_id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @connector_id,
    @external_account_id,
    @name,
    @created_at,
    @updated_at
)
ON CONFLICT %s DO UPDATE SET
    name = EXCLUDED.name,
    updated_at = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    created_at,
    updated_at;
`

	q = fmt.Sprintf(q, conflictTarget)

	args := pgx.StrictNamedArgs{
		"id":                  ca.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     ca.OrganizationID,
		"connector_id":        ca.ConnectorID,
		"external_account_id": ca.ExternalAccountID,
		"name":                ca.Name,
		"created_at":          ca.CreatedAt,
		"updated_at":          ca.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert connector account: %w", err)
	}

	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ConnectorAccount])
	if err != nil {
		return fmt.Errorf("cannot collect connector account: %w", err)
	}

	*ca = account

	return nil
}

func (ca *ConnectorAccount) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorAccountID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    created_at,
    updated_at
FROM
    connector_accounts
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": connectorAccountID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connector_accounts: %w", err)
	}

	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ConnectorAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect connector account: %w", err)
	}

	*ca = account

	return nil
}

// LoadByConnectorIDAndExternalAccountID resolves the account a vendor
// identifier names under one connector. Returns ErrResourceNotFound when the
// connector does not cover it.
func (ca *ConnectorAccount) LoadByConnectorIDAndExternalAccountID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
	externalAccountID string,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    created_at,
    updated_at
FROM
    connector_accounts
WHERE
    %s
    AND connector_id = @connector_id
    AND external_account_id = @external_account_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id":        connectorID,
		"external_account_id": externalAccountID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connector_accounts: %w", err)
	}

	account, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ConnectorAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect connector account: %w", err)
	}

	*ca = account

	return nil
}

// Update rewrites the vendor identifier and display name of an account.
//
// This is how a changed organization pick lands: the row a source already
// points at takes the new slug, so the source keeps its account and nothing
// has to be repointed.
func (ca *ConnectorAccount) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE connector_accounts
SET
    external_account_id = @external_account_id,
    name = @name,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  ca.ID,
		"external_account_id": ca.ExternalAccountID,
		"name":                ca.Name,
		"updated_at":          ca.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update connector account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// Delete removes one account. A source still referencing it makes the
// foreign key refuse, reported as ErrResourceInUse: the account a review is
// attached to cannot disappear under it.
func (ca *ConnectorAccount) Delete(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
DELETE FROM connector_accounts
WHERE
    %s
    AND id = @id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": ca.ID}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23503" {
				return ErrResourceInUse
			}
		}

		return fmt.Errorf("cannot delete connector account: %w", err)
	}

	return nil
}

func (cas *ConnectorAccounts) LoadByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
	cursor *page.Cursor[ConnectorAccountOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    created_at,
    updated_at
FROM
    connector_accounts
WHERE
    %s
    AND connector_id = @connector_id
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connector_accounts: %w", err)
	}

	accounts, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ConnectorAccount])
	if err != nil {
		return fmt.Errorf("cannot collect connector accounts: %w", err)
	}

	*cas = accounts

	return nil
}

func (cas *ConnectorAccounts) CountByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    connector_accounts
WHERE
    %s
    AND connector_id = @connector_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query connector_accounts: %w", err)
	}

	count, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[int])
	if err != nil {
		return 0, fmt.Errorf("cannot collect connector account count: %w", err)
	}

	return count, nil
}
