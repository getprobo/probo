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
)

type (
	AuditFramework struct {
		AuditID        gid.GID   `db:"audit_id"`
		FrameworkID    gid.GID   `db:"framework_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	AuditFrameworks []*AuditFramework
)

func (af *AuditFramework) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    audits_frameworks (
        audit_id,
        framework_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @audit_id,
    @framework_id,
    @organization_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"audit_id":        af.AuditID,
		"framework_id":    af.FrameworkID,
		"organization_id": af.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      af.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "audits_frameworks_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert audit framework: %w", err)
	}

	return nil
}

func (af *AuditFramework) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM
    audits_frameworks
WHERE
    %s
    AND audit_id = @audit_id
    AND framework_id = @framework_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"audit_id":     af.AuditID,
		"framework_id": af.FrameworkID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete audit framework: %w", err)
	}

	return nil
}

func (afs *AuditFrameworks) LoadByAuditIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	auditIDs []gid.GID,
) error {
	if len(auditIDs) == 0 {
		*afs = AuditFrameworks{}
		return nil
	}

	q := `
SELECT
    audit_id,
    framework_id,
    organization_id,
    created_at
FROM
    audits_frameworks
WHERE
    %s
    AND audit_id = ANY(@audit_ids)
ORDER BY
    audit_id,
    created_at,
    framework_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"audit_ids": auditIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query audit frameworks: %w", err)
	}

	links, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AuditFramework])
	if err != nil {
		return fmt.Errorf("cannot collect audit frameworks: %w", err)
	}

	*afs = links

	return nil
}
