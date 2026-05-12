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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	FindingAudit struct {
		FindingID      gid.GID   `db:"finding_id"`
		AuditID        gid.GID   `db:"audit_id"`
		ReferenceID    string    `db:"reference_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	FindingAudits []*FindingAudit
)

func (fa FindingAudit) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
    findings_audits (
        finding_id,
        audit_id,
        reference_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @finding_id,
    @audit_id,
    @reference_id,
    @organization_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (finding_id, audit_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"finding_id":      fa.FindingID,
		"audit_id":        fa.AuditID,
		"reference_id":    fa.ReferenceID,
		"organization_id": fa.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      fa.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert finding audit: %w", err)
	}

	return nil
}

func (fa FindingAudit) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	findingID gid.GID,
	auditID gid.GID,
) error {
	q := `
DELETE
FROM
    findings_audits
WHERE
    %s
    AND finding_id = @finding_id
    AND audit_id = @audit_id;
`

	args := pgx.StrictNamedArgs{
		"finding_id": findingID,
		"audit_id":   auditID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete finding audit: %w", err)
	}

	return nil
}
