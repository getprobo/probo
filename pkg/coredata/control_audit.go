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
	"fmt"
	"maps"
	"time"

	"go.probo.inc/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	ControlAudit struct {
		ControlID      gid.GID   `db:"control_id"`
		AuditID        gid.GID   `db:"audit_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	ControlAudits []*ControlAudit
)

func (ca ControlAudit) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    controls_audits (
        control_id,
        audit_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @control_id,
    @audit_id,
    @organization_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (control_id, audit_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"control_id":      ca.ControlID,
		"audit_id":        ca.AuditID,
		"organization_id": ca.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      ca.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (ca ControlAudit) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	auditID gid.GID,
) error {
	q := `
DELETE
FROM
    controls_audits
WHERE
    %s
    AND control_id = @control_id
    AND audit_id = @audit_id;
`

	args := pgx.StrictNamedArgs{
		"control_id": controlID,
		"audit_id":   auditID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (cas *ControlAudits) LoadByControlID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
) error {
	q := `
SELECT
    control_id,
    audit_id,
    created_at
FROM
    controls_audits
WHERE
    %s
    AND control_id = @control_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"control_id": controlID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query control_audits: %w", err)
	}

	controlAudits, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ControlAudit])
	if err != nil {
		return fmt.Errorf("cannot collect control_audits: %w", err)
	}

	*cas = controlAudits
	return nil
}

func (cas *ControlAudits) LoadByAuditID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	auditID gid.GID,
) error {
	q := `
SELECT
    control_id,
    audit_id,
    created_at
FROM
    controls_audits
WHERE
    %s
    AND audit_id = @audit_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_id": auditID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query control_audits: %w", err)
	}

	controlAudits, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ControlAudit])
	if err != nil {
		return fmt.Errorf("cannot collect control_audits: %w", err)
	}

	*cas = controlAudits
	return nil
}
