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
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
)

type (
	LogExport struct {
		ID             gid.GID         `db:"id"`
		OrganizationID gid.GID         `db:"organization_id"`
		Type           LogExportType   `db:"type"`
		Status         LogExportStatus `db:"status"`
		FromTime       time.Time       `db:"from_time"`
		ToTime         time.Time       `db:"to_time"`
		FileID         *gid.GID        `db:"file_id"`
		Error          *string         `db:"error"`
		RecipientEmail mail.Addr       `db:"recipient_email"`
		RecipientName  string          `db:"recipient_name"`
		CreatedAt      time.Time       `db:"created_at"`
		StartedAt      *time.Time      `db:"started_at"`
		CompletedAt    *time.Time      `db:"completed_at"`
	}

	LogExports []*LogExport
)

func (le *LogExport) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM log_exports WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, le.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query log export authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (le *LogExport) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO log_exports (
	id,
	tenant_id,
	organization_id,
	type,
	status,
	from_time,
	to_time,
	recipient_email,
	recipient_name,
	created_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@type,
	@status,
	@from_time,
	@to_time,
	@recipient_email,
	@recipient_name,
	@created_at
)`
	args := pgx.StrictNamedArgs{
		"id":              le.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": le.OrganizationID,
		"type":            le.Type,
		"status":          le.Status,
		"from_time":       le.FromTime,
		"to_time":         le.ToTime,
		"recipient_email": le.RecipientEmail,
		"recipient_name":  le.RecipientName,
		"created_at":      le.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (le *LogExport) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	log_exports
SET
	status = @status,
	error = @error,
	file_id = @file_id,
	started_at = @started_at,
	completed_at = @completed_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{
		"status":       le.Status,
		"error":        le.Error,
		"file_id":      le.FileID,
		"started_at":   le.StartedAt,
		"completed_at": le.CompletedAt,
		"id":           le.ID,
	}
	maps.Copy(args, scope.SQLArguments())
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (le *LogExport) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	type,
	status,
	from_time,
	to_time,
	file_id,
	error,
	recipient_email,
	recipient_name,
	created_at,
	started_at,
	completed_at
FROM
	log_exports
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}

	le2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[LogExport])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect log export: %w", err)
	}

	*le = le2
	return nil
}

func (le *LogExport) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	id,
	organization_id,
	type,
	status,
	from_time,
	to_time,
	file_id,
	error,
	recipient_email,
	recipient_name,
	created_at,
	started_at,
	completed_at
FROM
	log_exports
WHERE
	status = @status
ORDER BY
	created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"status": LogExportStatusPending,
	}
	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}

	le2, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[LogExport])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect log export: %w", err)
	}

	*le = le2
	return nil
}

func ResetStaleLogExports(
	ctx context.Context,
	conn pg.Conn,
	staleAfter time.Duration,
) error {
	q := `
UPDATE log_exports
SET
	status = @pending_status,
	started_at = NULL
WHERE
	status = @processing_status
	AND started_at < @stale_threshold
`
	args := pgx.StrictNamedArgs{
		"pending_status":    LogExportStatusPending,
		"processing_status": LogExportStatusProcessing,
		"stale_threshold":   time.Now().Add(-staleAfter),
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}
