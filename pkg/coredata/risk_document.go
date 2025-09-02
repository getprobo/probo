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

	"github.com/getprobo/probo/pkg/gid"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	RiskDocument struct {
		RiskID     gid.GID   `db:"risk_id"`
		DocumentID gid.GID   `db:"document_id"`
		SnapshotID *gid.GID  `db:"snapshot_id"`
		CreatedAt  time.Time `db:"created_at"`
	}

	RiskDocuments []*RiskDocument
)

func (rp RiskDocument) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    risks_documents (
        risk_id,
        document_id,
        tenant_id,
        created_at
    )
VALUES (
    @risk_id,
    @document_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"risk_id":     rp.RiskID,
		"document_id": rp.DocumentID,
		"tenant_id":   scope.GetTenantID(),
		"created_at":  rp.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (rp RiskDocument) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	documentID gid.GID,
) error {
	q := `
DELETE
FROM
    risks_documents
WHERE
    %s
    AND risk_id = @risk_id
    AND document_id = @document_id;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_id":     riskID,
		"document_id": documentID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (rd RiskDocuments) InsertRiskSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	snapshot_risks AS (
		SELECT id, source_id
		FROM risks
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	)
INSERT INTO risks_documents (
	tenant_id,
	snapshot_id,
	risk_id,
	document_id,
	created_at
)
SELECT
	@tenant_id,
	@snapshot_id,
	sr.id,
	rd.document_id,
	rd.created_at
FROM risks_documents rd
INNER JOIN snapshot_risks sr ON sr.source_id = rd.risk_id
WHERE %s AND rd.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert risk document snapshots: %w", err)
	}

	return nil
}
