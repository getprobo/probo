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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	AuditReport struct {
		AuditID   gid.GID   `db:"audit_id"`
		ReportID  gid.GID   `db:"report_id"`
		CreatedAt time.Time `db:"created_at"`
	}

	AuditReports []*AuditReport
)

func (ar AuditReport) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    audit_reports (
        audit_id,
        report_id,
        tenant_id,
        created_at
    )
VALUES (
    @audit_id,
    @report_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (audit_id, report_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"audit_id":   ar.AuditID,
		"report_id":  ar.ReportID,
		"tenant_id":  scope.GetTenantID(),
		"created_at": ar.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (ar AuditReport) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
	reportID gid.GID,
) error {
	q := `
DELETE
FROM
    audit_reports
WHERE
    audit_id = @audit_id
    AND report_id = @report_id
    AND tenant_id = @tenant_id;
`

	args := pgx.StrictNamedArgs{
		"audit_id":  auditID,
		"report_id": reportID,
		"tenant_id": scope.GetTenantID(),
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (ars *AuditReports) LoadByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
) error {
	q := `
SELECT
    audit_id,
    report_id,
    created_at
FROM
    audit_reports
WHERE
    audit_id = @audit_id
    AND tenant_id = @tenant_id
ORDER BY
    created_at DESC;
`

	args := pgx.StrictNamedArgs{
		"audit_id":  auditID,
		"tenant_id": scope.GetTenantID(),
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return err
	}
	defer rows.Close()

	var auditReports []*AuditReport
	for rows.Next() {
		ar := &AuditReport{}
		if err := rows.Scan(&ar.AuditID, &ar.ReportID, &ar.CreatedAt); err != nil {
			return err
		}
		auditReports = append(auditReports, ar)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	*ars = auditReports
	return nil
}

func (ar AuditReport) DeleteByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
) error {
	q := `
DELETE
FROM
    audit_reports
WHERE
    audit_id = @audit_id
    AND tenant_id = @tenant_id;
`

	args := pgx.StrictNamedArgs{
		"audit_id":  auditID,
		"tenant_id": scope.GetTenantID(),
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}