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
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
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

func (e *AuditLogEntry) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM audit_log_entries WHERE id = ANY(@resource_ids::text[])`

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

// BulkInsert writes all entries to audit_log_entries in a single PostgreSQL
// COPY operation. Used by the authorizer's batch path to fold N per-resource
// INSERTs into a single round trip.
func (es AuditLogEntries) BulkInsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	if len(es) == 0 {
		return nil
	}

	rows := make([][]any, 0, len(es))
	for _, e := range es {
		rows = append(
			rows,
			[]any{
				e.ID,
				scope.GetTenantID(),
				e.OrganizationID,
				e.ActorID,
				e.ActorType,
				e.Action,
				e.ResourceType,
				e.ResourceID,
				e.Metadata,
				e.CreatedAt,
			},
		)
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"audit_log_entries"},
		[]string{
			"id",
			"tenant_id",
			"organization_id",
			"actor_id",
			"actor_type",
			"action",
			"resource_type",
			"resource_id",
			"metadata",
			"created_at",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("cannot bulk insert audit log entries: %w", err)
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
