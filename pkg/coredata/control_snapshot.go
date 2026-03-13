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

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	ControlSnapshot struct {
		ControlID      gid.GID   `db:"control_id"`
		SnapshotID     gid.GID   `db:"snapshot_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	ControlSnapshots []*ControlSnapshot
)

func (cs ControlSnapshot) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    controls_snapshots (
        control_id,
        snapshot_id,
        organization_id,
        tenant_id,
        created_at
    )
VALUES (
    @control_id,
    @snapshot_id,
    @organization_id,
    @tenant_id,
    @created_at
)
ON CONFLICT (control_id, snapshot_id) DO NOTHING;
`

	args := pgx.StrictNamedArgs{
		"control_id":      cs.ControlID,
		"snapshot_id":     cs.SnapshotID,
		"organization_id": cs.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"created_at":      cs.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (cs ControlSnapshot) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	controlID gid.GID,
	snapshotID gid.GID,
) error {
	q := `
DELETE
FROM
    controls_snapshots
WHERE
    %s
    AND control_id = @control_id
    AND snapshot_id = @snapshot_id;
`

	args := pgx.StrictNamedArgs{
		"control_id":  controlID,
		"snapshot_id": snapshotID,
	}
	maps.Copy(args, scope.SQLArguments())
	q = fmt.Sprintf(q, scope.SQLFragment())

	_, err := conn.Exec(ctx, q, args)
	return err
}

