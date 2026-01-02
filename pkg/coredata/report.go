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
	Report struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		ObjectKey      string    `db:"object_key"`
		MimeType       string    `db:"mime_type"`
		Filename       string    `db:"filename"`
		Size           int64     `db:"size"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Reports []*Report
)

func (r *Report) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM reports WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, r.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query report authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (r *Report) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	reportID gid.GID,
) error {
	q := `
SELECT
  id,
  organization_id,
  object_key,
  mime_type,
  filename,
  size,
  created_at,
  updated_at
FROM
  reports
WHERE
  %s
  AND id = @report_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"report_id": reportID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query report: %w", err)
	}

	report, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Report])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect report: %w", err)
	}

	*r = report

	return nil
}

func (r *Reports) LoadByIDs(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	reportIDs []gid.GID,
) error {
	q := `
SELECT
  id,
  organization_id,
  object_key,
  mime_type,
  filename,
  size,
  created_at,
  updated_at
FROM
  reports
WHERE
  %s
  AND id = ANY(@report_ids);
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"report_ids": reportIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query report: %w", err)
	}

	reports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Report])
	if err != nil {
		return fmt.Errorf("cannot collect reports: %w", err)
	}

	*r = reports

	return nil
}

func (r *Report) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO reports (
  id,
  tenant_id,
  organization_id,
  object_key,
  mime_type,
  filename,
  size,
  created_at,
  updated_at
) VALUES (
  @id,
  @tenant_id,
  @organization_id,
  @object_key,
  @mime_type,
  @filename,
  @size,
  @created_at,
  @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              r.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": r.OrganizationID,
		"object_key":      r.ObjectKey,
		"mime_type":       r.MimeType,
		"filename":        r.Filename,
		"size":            r.Size,
		"created_at":      r.CreatedAt,
		"updated_at":      r.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert report: %w", err)
	}

	return nil
}

func (r *Report) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE reports
SET
  object_key = @object_key,
  mime_type = @mime_type,
  filename = @filename,
  size = @size,
  updated_at = @updated_at
WHERE
  %s
  AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         r.ID,
		"object_key": r.ObjectKey,
		"mime_type":  r.MimeType,
		"filename":   r.Filename,
		"size":       r.Size,
		"updated_at": r.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update report: %w", err)
	}

	return nil
}

func (r *Report) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM reports
WHERE
  %s
  AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": r.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete report: %w", err)
	}

	return nil
}

func (r *Report) CursorKey(orderBy ReportOrderField) page.CursorKey {
	switch orderBy {
	case ReportOrderFieldID:
		return page.NewCursorKey(r.ID, r.ID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}
