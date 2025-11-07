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
	RiskObligation struct {
		RiskID         gid.GID   `db:"risk_id"`
		ObligationID   gid.GID   `db:"obligation_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	RiskObligations []*RiskObligation
)

func (ro RiskObligation) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO risks_obligations (
	risk_id,
	obligation_id,
	organization_id,
	tenant_id,
	created_at
) VALUES (
	@risk_id,
	@obligation_id,
	@organization_id,
	@tenant_id,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"risk_id":         ro.RiskID,
		"obligation_id":   ro.ObligationID,
		"organization_id": ro.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      ro.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert risk obligation: %w", err)
	}

	return nil
}

func (ro RiskObligation) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM risks_obligations
WHERE
	%s
	AND risk_id = @risk_id
	AND obligation_id = @obligation_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_id":       ro.RiskID,
		"obligation_id": ro.ObligationID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete risk obligation: %w", err)
	}

	return nil
}
