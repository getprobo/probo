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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
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
	conn pg.Querier,
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
	conn pg.Tx,
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
