// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"encoding/json"
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
	AuditLogEntry struct {
		ID             gid.GID           `db:"id"`
		OrganizationID gid.GID           `db:"organization_id"`
		ActorID        gid.GID           `db:"actor_id"`
		ActorType      AuditLogActorType `db:"actor_type"`
		Action         string            `db:"action"`
		ResourceType   string            `db:"resource_type"`
		ResourceID     gid.GID           `db:"resource_id"`
		Metadata       json.RawMessage   `db:"metadata"`
		CreatedAt      time.Time         `db:"created_at"`
	}

	AuditLogEntries []*AuditLogEntry
)

func (e AuditLogEntry) CursorKey(orderBy AuditLogEntryOrderField) page.CursorKey {
	switch orderBy {
	case AuditLogEntryOrderFieldCreatedAt:
		return page.NewCursorKey(e.ID, e.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (e *AuditLogEntry) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM audit_log_entries WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, e.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query audit log entry authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (e *AuditLogEntry) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO audit_log_entries (
    id,
    tenant_id,
    organization_id,
    actor_id,
    actor_type,
    action,
    resource_type,
    resource_id,
    metadata,
    created_at
)
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @actor_id,
    @actor_type,
    @action,
    @resource_type,
    @resource_id,
    @metadata,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              e.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": e.OrganizationID,
		"actor_id":        e.ActorID,
		"actor_type":      e.ActorType,
		"action":          e.Action,
		"resource_type":   e.ResourceType,
		"resource_id":     e.ResourceID,
		"metadata":        e.Metadata,
		"created_at":      e.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert audit log entry: %w", err)
	}

	return nil
}

func (e *AuditLogEntry) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    actor_id,
    actor_type,
    action,
    resource_type,
    resource_id,
    metadata,
    created_at
FROM
    audit_log_entries
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
		return fmt.Errorf("cannot query audit log entry: %w", err)
	}

	entry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AuditLogEntry])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect audit log entry: %w", err)
	}

	*e = entry
	return nil
}

func (es *AuditLogEntries) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AuditLogEntryOrderField],
	filter *AuditLogEntryFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    actor_id,
    actor_type,
    action,
    resource_type,
    resource_id,
    metadata,
    created_at
FROM
    audit_log_entries
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audit log entries: %w", err)
	}

	entries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AuditLogEntry])
	if err != nil {
		return fmt.Errorf("cannot collect audit log entries: %w", err)
	}

	*es = entries
	return nil
}

func (es *AuditLogEntries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *AuditLogEntryFilter,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM audit_log_entries
WHERE %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count audit log entries: %w", err)
	}

	return count, nil
}
