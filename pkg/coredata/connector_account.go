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

var (
	// ErrMultipleConnectorAccounts is returned when a caller asks for the
	// connector's only account and the connector has more than one.
	ErrMultipleConnectorAccounts = errors.New("connector has more than one account")
)

type (
	ConnectorAccount struct {
		ID                    gid.GID    `db:"id"`
		OrganizationID        gid.GID    `db:"organization_id"`
		ConnectorID           gid.GID    `db:"connector_id"`
		ExternalAccountID     string     `db:"external_account_id"`
		Name                  string     `db:"name"`
		NameSyncedAt          *time.Time `db:"name_synced_at"`
		NameSyncAttempts      int        `db:"name_sync_attempts"`
		NameSyncNextAttemptAt *time.Time `db:"name_sync_next_attempt_at"`
		CreatedAt             time.Time  `db:"created_at"`
		UpdatedAt             time.Time  `db:"updated_at"`
	}

	ConnectorAccounts []*ConnectorAccount
)

func (a ConnectorAccount) CursorKey(orderBy ConnectorAccountOrderField) page.CursorKey {
	switch orderBy {
	case ConnectorAccountOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	case ConnectorAccountOrderFieldExternalAccountID:
		return page.NewCursorKey(a.ID, a.ExternalAccountID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *ConnectorAccount) AuthorizationAttributes(
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

func (a *ConnectorAccount) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
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

	args := pgx.StrictNamedArgs{"id": id}
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

		return fmt.Errorf("cannot collect connector_account: %w", err)
	}

	*a = account

	return nil
}

func ConnectorIDsByAccountIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) (map[gid.GID]gid.GID, error) {
	connectorIDs := make(map[gid.GID]gid.GID, len(ids))
	if len(ids) == 0 {
		return connectorIDs, nil
	}

	q := `
SELECT
    id,
    connector_id
FROM
    connector_accounts
WHERE
    %s
    AND id = ANY(@ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query connector accounts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id, connectorID gid.GID
		if err := rows.Scan(&id, &connectorID); err != nil {
			return nil, fmt.Errorf("cannot scan connector account: %w", err)
		}

		connectorIDs[id] = connectorID
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate connector accounts: %w", err)
	}

	return connectorIDs, nil
}

func (accounts *ConnectorAccounts) LoadByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	ids []gid.GID,
) error {
	*accounts = nil

	if len(ids) == 0 {
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    connector_accounts
WHERE
    %s
    AND id = ANY(@ids)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ids": ids}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query connector_accounts: %w", err)
	}

	loaded, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ConnectorAccount])
	if err != nil {
		return fmt.Errorf("cannot collect connector accounts: %w", err)
	}

	*accounts = loaded

	return nil
}

func (a *ConnectorAccount) LoadByConnectorAndExternalID(
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
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
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

		return fmt.Errorf("cannot collect connector_account: %w", err)
	}

	*a = account

	return nil
}

func (accounts *ConnectorAccounts) LoadByConnectorID(
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
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
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

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ConnectorAccount])
	if err != nil {
		return fmt.Errorf("cannot collect connector_accounts: %w", err)
	}

	*accounts = result

	return nil
}

func (accounts *ConnectorAccounts) CountByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM connector_accounts
WHERE
    %s
    AND connector_id = @connector_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"connector_id": connectorID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count connector_accounts: %w", err)
	}

	return count, nil
}

// CountDistinctExternalAccountIDsByOrganizationAndProvider counts external
// account ids once across every connector of provider in the organization.
// The same id stored on two credentials counts once.
func (accounts *ConnectorAccounts) CountDistinctExternalAccountIDsByOrganizationAndProvider(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	provider ConnectorProvider,
) (int, error) {
	q := `
SELECT COUNT(DISTINCT external_account_id)
FROM connector_accounts
WHERE
    %s
    AND organization_id = @organization_id
    AND connector_id IN (
        SELECT id
        FROM connectors
        WHERE
            %s
            AND organization_id = @organization_id
            AND provider = @provider
    );
`
	q = fmt.Sprintf(q, scope.SQLFragment(), scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
		"provider":        provider,
	}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count distinct connector_accounts: %w", err)
	}

	return count, nil
}

func (a *ConnectorAccount) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) (inserted bool, err error) {
	q := `
INSERT INTO connector_accounts (
    id,
    tenant_id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
)
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @connector_id,
    @external_account_id,
    @name,
    @name_synced_at,
    @name_sync_attempts,
    @name_sync_next_attempt_at,
    @created_at,
    @updated_at
)
ON CONFLICT (connector_id, external_account_id) DO UPDATE SET
    name       = EXCLUDED.name,
    updated_at = EXCLUDED.updated_at
RETURNING
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
`
	originalID := a.ID

	args := pgx.StrictNamedArgs{
		"id":                        a.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           a.OrganizationID,
		"connector_id":              a.ConnectorID,
		"external_account_id":       a.ExternalAccountID,
		"name":                      a.Name,
		"name_synced_at":            a.NameSyncedAt,
		"name_sync_attempts":        a.NameSyncAttempts,
		"name_sync_next_attempt_at": a.NameSyncNextAttemptAt,
		"created_at":                a.CreatedAt,
		"updated_at":                a.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert connector_account: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ConnectorAccount])
	if err != nil {
		return false, fmt.Errorf("cannot collect connector_account: %w", err)
	}

	*a = row

	return originalID == a.ID, nil
}

func UpsertInitialAccount(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	cnnctr *Connector,
	externalID string,
	name string,
) (*ConnectorAccount, error) {
	if externalID == "" {
		return nil, nil
	}

	if name == "" {
		name = externalID
	}

	now := time.Now()
	account := &ConnectorAccount{
		ID:                gid.New(scope.GetTenantID(), ConnectorAccountEntityType),
		OrganizationID:    cnnctr.OrganizationID,
		ConnectorID:       cnnctr.ID,
		ExternalAccountID: externalID,
		Name:              name,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if _, err := account.Upsert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot upsert initial connector account: %w", err)
	}

	return account, nil
}

func (a *ConnectorAccount) LoadStandaloneByConnectorID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	connectorID gid.GID,
) error {
	var accounts ConnectorAccounts

	cursor := page.NewCursor(
		2,
		nil,
		page.Head,
		page.OrderBy[ConnectorAccountOrderField]{
			Field:     ConnectorAccountOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	if err := accounts.LoadByConnectorID(ctx, conn, scope, connectorID, cursor); err != nil {
		return err
	}

	if len(accounts) == 0 {
		return ErrResourceNotFound
	}

	if len(accounts) > 1 {
		return fmt.Errorf("cannot resolve connector account: %w", ErrMultipleConnectorAccounts)
	}

	*a = *accounts[0]

	return nil
}

// SyncStandaloneAccount writes the given tenant onto the connector's single
// account row: insert when none exists, update external_account_id when the
// picker changes. It never inserts a second row. An empty externalID leaves
// the row untouched. The connector row is locked first so two syncs cannot
// both insert.
func SyncStandaloneAccount(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cnnctr *Connector,
	externalID string,
	name string,
) (*ConnectorAccount, error) {
	if externalID == "" {
		return nil, nil
	}

	if name == "" {
		name = externalID
	}

	if err := cnnctr.LockByID(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot lock connector: %w", err)
	}

	existing := &ConnectorAccount{}

	err := existing.LoadStandaloneByConnectorID(ctx, tx, scope, cnnctr.ID)
	switch {
	case errors.Is(err, ErrResourceNotFound):
		return UpsertInitialAccount(ctx, tx, scope, cnnctr, externalID, name)
	case err != nil:
		return nil, err
	}

	if existing.ExternalAccountID == externalID && existing.Name == name {
		return existing, nil
	}

	existing.ExternalAccountID = externalID
	existing.Name = name
	existing.UpdatedAt = time.Now()

	if err := existing.Update(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot update connector account: %w", err)
	}

	return existing, nil
}

func (a *ConnectorAccount) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE connector_accounts
SET
    external_account_id = @external_account_id,
    name = @name,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  a.ID,
		"external_account_id": a.ExternalAccountID,
		"name":                a.Name,
		"updated_at":          a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "idx_connector_accounts_connector_external" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot update connector_account: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (a *ConnectorAccount) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM connector_accounts
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": a.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23503" {
				return ErrResourceInUse
			}
		}

		return fmt.Errorf("cannot delete connector_account: %w", err)
	}

	return nil
}

// RecordNameSyncAttempt charges the claimed account for one name-resolution
// attempt and holds it out of the queue for backoff. It must run in the
// transaction that claimed the row: the claim predicate does not otherwise
// change on failure, so an uncharged attempt is re-claimed immediately.
//
// The deadline is computed from the database clock, which is the clock the
// claim predicate compares against.
func (a *ConnectorAccount) RecordNameSyncAttempt(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	backoff time.Duration,
) error {
	q := `
UPDATE connector_accounts
SET
    name_sync_attempts = name_sync_attempts + 1,
    name_sync_next_attempt_at = NOW() + @backoff::interval
WHERE
    %s
    AND id = @id
RETURNING name_sync_attempts, name_sync_next_attempt_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":      a.ID,
		"backoff": backoff,
	}
	maps.Copy(args, scope.SQLArguments())

	err := conn.QueryRow(ctx, q, args).Scan(&a.NameSyncAttempts, &a.NameSyncNextAttemptAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot record connector account name sync attempt: %w", err)
	}

	return nil
}

// MarkNameSynced retires the account from the name-sync queue and writes the
// resolved instance name. It applies only while the row still carries the
// deadline and attempt count the caller claimed: a reconnect that lands during
// the resolve wins. applied is false in that case, and the error is nil, so
// the worker does not write a stale name.
//
// The deadline is what makes the guard a claim generation. The attempt count
// alone is not unique — a reconnect resets it to zero and the next claim
// charges it straight back to the same number.
func (a *ConnectorAccount) MarkNameSynced(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	syncedAt time.Time,
) (bool, error) {
	q := `
UPDATE connector_accounts
SET
    name = @name,
    name_synced_at = @name_synced_at,
    name_sync_next_attempt_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
    AND name_synced_at IS NULL
    AND name_sync_attempts = @name_sync_attempts
    AND name_sync_next_attempt_at IS NOT DISTINCT FROM @name_sync_next_attempt_at
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        a.ID,
		"name":                      a.Name,
		"name_synced_at":            syncedAt,
		"name_sync_attempts":        a.NameSyncAttempts,
		"name_sync_next_attempt_at": a.NameSyncNextAttemptAt,
		"updated_at":                syncedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot mark connector account name synced: %w", err)
	}

	if result.RowsAffected() == 0 {
		return false, nil
	}

	a.NameSyncedAt = &syncedAt
	a.NameSyncNextAttemptAt = nil
	a.UpdatedAt = syncedAt

	return true, nil
}

// ResetNameSyncByConnectorID clears the synced name and the retry budget for
// every account of connectorID, so the name worker re-resolves their display
// name. No-op when the connector has no account.
//
// Rows already at NULL are included: an account still inside its backoff, or
// one that exhausted its attempts against a misconfigured connector, is exactly
// what the reconnect that fixed it needs to release.
func (accounts *ConnectorAccounts) ResetNameSyncByConnectorID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	connectorID gid.GID,
) error {
	q := `
UPDATE connector_accounts
SET
    name_synced_at = NULL,
    name_sync_attempts = 0,
    name_sync_next_attempt_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND connector_id = @connector_id
    AND (name_synced_at IS NOT NULL OR name_sync_attempts > 0)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"connector_id": connectorID,
		"updated_at":   time.Now(),
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot clear connector account name sync: %w", err)
	}

	return nil
}

// ResetNameSyncByID clears the synced name and the retry budget for one
// account. A relink or an organization change may resolve to a different
// instance, so the name worker has to run again.
func (a *ConnectorAccount) ResetNameSyncByID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE connector_accounts
SET
    name_synced_at = NULL,
    name_sync_attempts = 0,
    name_sync_next_attempt_at = NULL,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
    AND (name_synced_at IS NOT NULL OR name_sync_attempts > 0)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	now := time.Now()

	args := pgx.StrictNamedArgs{
		"id":         a.ID,
		"updated_at": now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot clear connector account name sync: %w", err)
	}

	a.NameSyncedAt = nil
	a.NameSyncAttempts = 0
	a.NameSyncNextAttemptAt = nil
	a.UpdatedAt = now

	return nil
}

// ErrNoConnectorAccountNameSyncAvailable is returned when no connector account
// needs its name resolved.
var ErrNoConnectorAccountNameSyncAvailable = fmt.Errorf("no connector account name sync available")

// LoadNextUnsyncedNameForUpdateSkipLocked claims the next connector account
// that has not yet had its name resolved and is not backing off from a
// previous failure. The row is locked with FOR UPDATE SKIP LOCKED so
// concurrent workers do not pick the same row.
//
// The query is intentionally cross-tenant: the worker serves every tenant and
// re-scopes from the claimed row's own GID.
//
// The backoff gate is what bounds the claim. Nothing else transitions the
// predicate on failure, so without it the same row is re-claimed as fast as
// it can be processed. Callers must charge the attempt with
// RecordNameSyncAttempt in this same transaction.
func (a *ConnectorAccount) LoadNextUnsyncedNameForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_id,
    external_account_id,
    name,
    name_synced_at,
    name_sync_attempts,
    name_sync_next_attempt_at,
    created_at,
    updated_at
FROM
    connector_accounts
WHERE
    name_synced_at IS NULL
    AND (name_sync_next_attempt_at IS NULL OR name_sync_next_attempt_at <= NOW())
ORDER BY
    created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query unsynced connector_accounts: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ConnectorAccount])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoConnectorAccountNameSyncAvailable
		}

		return fmt.Errorf("cannot collect unsynced connector account: %w", err)
	}

	*a = row

	return nil
}
