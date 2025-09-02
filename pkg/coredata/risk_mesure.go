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
	RiskMeasure struct {
		RiskID     gid.GID   `db:"risk_id"`
		MeasureID  gid.GID   `db:"measure_id"`
		SnapshotID *gid.GID  `db:"snapshot_id"`
		CreatedAt  time.Time `db:"created_at"`
	}

	RiskMeasures []*RiskMeasure
)

func (rm RiskMeasure) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    risks_measures (
        risk_id,
        measure_id,
        tenant_id,
        created_at
    )
VALUES (
    @risk_id,
    @measure_id,
    @tenant_id,
    @created_at
);
`

	args := pgx.StrictNamedArgs{
		"risk_id":    rm.RiskID,
		"measure_id": rm.MeasureID,
		"tenant_id":  scope.GetTenantID(),
		"created_at": rm.CreatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

func (rm RiskMeasure) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
	measureID gid.GID,
) error {
	q := `
DELETE
FROM
    risks_measures
WHERE
    %s
    AND risk_id = @risk_id
    AND measure_id = @measure_id
	AND snapshot_id IS NULL;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_id":    riskID,
		"measure_id": measureID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (rm RiskMeasures) InsertRiskSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	source_risks AS (
		SELECT id
		FROM risks
		WHERE organization_id = @organization_id AND snapshot_id IS NULL
	),
	snapshot_risks AS (
		SELECT id, source_id
		FROM risks
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	),
	snapshot_measures AS (
		SELECT id, source_id
		FROM measures
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	),
	source_risk_measures AS (
		SELECT risk_id, measure_id, snapshot_id, created_at
		FROM risks_measures
		WHERE %s AND risk_id = ANY(SELECT id FROM source_risks) AND snapshot_id IS NULL
	)
INSERT INTO risks_measures (
	tenant_id,
	risk_id,
	measure_id,
	snapshot_id,
	created_at
)
SELECT
	@tenant_id,
	sr.id,
	sm.id,
	@snapshot_id,
	rm.created_at
FROM source_risk_measures rm
JOIN snapshot_risks sr ON sr.source_id = rm.risk_id
JOIN snapshot_measures sm ON sm.source_id = rm.measure_id
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
		return fmt.Errorf("cannot insert risk measure snapshots: %w", err)
	}

	return nil
}
