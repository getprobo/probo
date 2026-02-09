// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	SCIMBridge struct {
		ID                  gid.GID         `db:"id"`
		OrganizationID      gid.GID         `db:"organization_id"`
		ScimConfigurationID gid.GID         `db:"scim_configuration_id"`
		ConnectorID         *gid.GID        `db:"connector_id"`
		Type                SCIMBridgeType  `db:"type"`
		State               SCIMBridgeState `db:"state"`
		ExcludedUserNames   []string        `db:"excluded_user_names"`
		LastSyncedAt        *time.Time      `db:"last_synced_at"`
		NextSyncAt          *time.Time      `db:"next_sync_at"`
		SyncError           *string         `db:"sync_error"`
		ConsecutiveFailures int             `db:"consecutive_failures"`
		TotalSyncCount      int             `db:"total_sync_count"`
		TotalFailureCount   int             `db:"total_failure_count"`
		CreatedAt           time.Time       `db:"created_at"`
		UpdatedAt           time.Time       `db:"updated_at"`
	}

	SCIMBridges []*SCIMBridge
)

var ErrNoSCIMBridgeAvailable = errors.New("no SCIM bridge available for sync")

func (s *SCIMBridge) CursorKey(orderBy SCIMBridgeOrderField) page.CursorKey {
	switch orderBy {
	case SCIMBridgeOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *SCIMBridge) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM iam_scim_bridges WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query scim bridge authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *SCIMBridge) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	bridgeID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    connector_id,
    type,
    state,
    excluded_user_names,
    last_synced_at,
    next_sync_at,
    sync_error,
    consecutive_failures,
    total_sync_count,
    total_failure_count,
    created_at,
    updated_at
FROM
    iam_scim_bridges
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": bridgeID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_bridges: %w", err)
	}

	bridge, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMBridge])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_bridge: %w", err)
	}

	*s = bridge

	return nil
}

func (s *SCIMBridge) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    connector_id,
    type,
    state,
    excluded_user_names,
    last_synced_at,
    next_sync_at,
    sync_error,
    consecutive_failures,
    total_sync_count,
    total_failure_count,
    created_at,
    updated_at
FROM
    iam_scim_bridges
WHERE
    %s
    AND organization_id = @organization_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_bridges: %w", err)
	}

	bridge, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMBridge])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_bridge: %w", err)
	}

	*s = bridge

	return nil
}

func (s *SCIMBridge) LoadBySCIMConfigurationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	scimConfigurationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    connector_id,
    type,
    state,
    excluded_user_names,
    last_synced_at,
    next_sync_at,
    sync_error,
    consecutive_failures,
    total_sync_count,
    total_failure_count,
    created_at,
    updated_at
FROM
    iam_scim_bridges
WHERE
    %s
    AND scim_configuration_id = @scim_configuration_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"scim_configuration_id": scimConfigurationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_bridges: %w", err)
	}

	bridge, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMBridge])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect scim_bridge: %w", err)
	}

	*s = bridge

	return nil
}

func (s *SCIMBridge) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO iam_scim_bridges (
    id,
    tenant_id,
    organization_id,
    scim_configuration_id,
    connector_id,
    type,
    state,
    excluded_user_names,
    last_synced_at,
    next_sync_at,
    sync_error,
    consecutive_failures,
    total_sync_count,
    total_failure_count,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @scim_configuration_id,
    @connector_id,
    @type,
    @state,
    @excluded_user_names,
    @last_synced_at,
    @next_sync_at,
    @sync_error,
    @consecutive_failures,
    @total_sync_count,
    @total_failure_count,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                    s.ID,
		"tenant_id":             scope.GetTenantID(),
		"organization_id":       s.OrganizationID,
		"scim_configuration_id": s.ScimConfigurationID,
		"connector_id":          s.ConnectorID,
		"type":                  s.Type,
		"state":                 s.State,
		"excluded_user_names":   s.ExcludedUserNames,
		"last_synced_at":        s.LastSyncedAt,
		"next_sync_at":          s.NextSyncAt,
		"sync_error":            s.SyncError,
		"consecutive_failures":  s.ConsecutiveFailures,
		"total_sync_count":      s.TotalSyncCount,
		"total_failure_count":   s.TotalFailureCount,
		"created_at":            s.CreatedAt,
		"updated_at":            s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert scim_bridge: %w", err)
	}

	return nil
}

func (s *SCIMBridge) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE iam_scim_bridges
SET
    connector_id = @connector_id,
    state = @state,
    excluded_user_names = @excluded_user_names,
    last_synced_at = @last_synced_at,
    next_sync_at = @next_sync_at,
    sync_error = @sync_error,
    consecutive_failures = @consecutive_failures,
    total_sync_count = @total_sync_count,
    total_failure_count = @total_failure_count,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                   s.ID,
		"connector_id":         s.ConnectorID,
		"state":                s.State,
		"excluded_user_names":  s.ExcludedUserNames,
		"last_synced_at":       s.LastSyncedAt,
		"next_sync_at":         s.NextSyncAt,
		"sync_error":           s.SyncError,
		"consecutive_failures": s.ConsecutiveFailures,
		"total_sync_count":     s.TotalSyncCount,
		"total_failure_count":  s.TotalFailureCount,
		"updated_at":           s.UpdatedAt,
	}

	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update scim_bridge: %w", err)
	}

	return nil
}

func (s *SCIMBridge) LoadNextForSyncSkipLocked(
	ctx context.Context,
	conn pg.Conn,
	staleSyncThreshold time.Duration,
) error {
	staleCutoff := time.Now().Add(-staleSyncThreshold)

	q := `
SELECT
    id,
    organization_id,
    scim_configuration_id,
    connector_id,
    type,
    state,
    excluded_user_names,
    last_synced_at,
    next_sync_at,
    sync_error,
    consecutive_failures,
    total_sync_count,
    total_failure_count,
    created_at,
    updated_at
FROM
    iam_scim_bridges
WHERE
    (state IN (@state_active, @state_failed) AND (next_sync_at IS NULL OR next_sync_at <= NOW()))
    OR (state = @state_syncing AND updated_at < @stale_cutoff)
ORDER BY
    next_sync_at ASC NULLS FIRST
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"state_active":  SCIMBridgeStateActive,
		"state_failed":  SCIMBridgeStateFailed,
		"state_syncing": SCIMBridgeStateSyncing,
		"stale_cutoff":  staleCutoff,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_scim_bridges: %w", err)
	}

	bridge, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SCIMBridge])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoSCIMBridgeAvailable
		}
		return fmt.Errorf("cannot collect scim_bridge: %w", err)
	}

	*s = bridge
	return nil
}

func (s *SCIMBridge) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM iam_scim_bridges
WHERE
    %s
    AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete scim_bridge: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}
