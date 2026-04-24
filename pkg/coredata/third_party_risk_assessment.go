// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	// RiskAssessment represents a point-in-time risk assessment for a thirdParty
	ThirdPartyRiskAssessment struct {
		ID              gid.GID         `db:"id"`
		OrganizationID  gid.GID         `db:"organization_id"`
		ThirdPartyID    gid.GID         `db:"third_party_id"`
		ExpiresAt       time.Time       `db:"expires_at"`
		DataSensitivity DataSensitivity `db:"data_sensitivity"`
		BusinessImpact  BusinessImpact  `db:"business_impact"`
		Notes           *string         `db:"notes"`
		SnapshotID      *gid.GID        `db:"snapshot_id"`
		SourceID        *gid.GID        `db:"source_id"`
		CreatedAt       time.Time       `db:"created_at"`
		UpdatedAt       time.Time       `db:"updated_at"`
	}

	ThirdPartyRiskAssessments []*ThirdPartyRiskAssessment
)

func (v ThirdPartyRiskAssessment) CursorKey(orderBy ThirdPartyRiskAssessmentOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyRiskAssessmentOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	case ThirdPartyRiskAssessmentOrderFieldExpiresAt:
		return page.NewCursorKey(v.ID, v.ExpiresAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (v *ThirdPartyRiskAssessment) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM third_party_risk_assessments WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query thirdParty risk assessment authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

// Insert adds a new risk assessment to the database
func (r ThirdPartyRiskAssessment) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    third_party_risk_assessments (
        tenant_id,
        id,
        organization_id,
        third_party_id,
        expires_at,
        data_sensitivity,
        business_impact,
        notes,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @id,
    @organization_id,
    @third_party_id,
    @expires_at,
    @data_sensitivity,
    @business_impact,
    @notes,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"id":               r.ID,
		"organization_id":  r.OrganizationID,
		"third_party_id":   r.ThirdPartyID,
		"expires_at":       r.ExpiresAt,
		"data_sensitivity": r.DataSensitivity,
		"business_impact":  r.BusinessImpact,
		"notes":            r.Notes,
		"created_at":       r.CreatedAt,
		"updated_at":       r.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	return err
}

// LoadByID loads a risk assessment by its ID
func (r *ThirdPartyRiskAssessment) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    third_party_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    third_party_risk_assessments
WHERE
    %s
    AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment: %w", err)
	}
	defer rows.Close()

	assessment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment: %w", err)
	}

	*r = assessment

	return nil
}

// LoadLatestByThirdPartyID loads the most recent risk assessment for a thirdParty
func (r *ThirdPartyRiskAssessment) LoadLatestByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    third_party_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    third_party_risk_assessments
WHERE
    %s
    AND third_party_id = @third_party_id
ORDER BY
    created_at DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment: %w", err)
	}
	defer rows.Close()

	assessment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment: %w", err)
	}

	*r = assessment

	return nil
}

// LoadByThirdPartyID loads all risk assessments for a thirdParty, ordered by assessment date
func (r *ThirdPartyRiskAssessments) LoadByThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyRiskAssessmentOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    third_party_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    third_party_risk_assessments
WHERE
    %s
    AND third_party_id = @third_party_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessments: %w", err)
	}

	assessments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessments: %w", err)
	}

	*r = assessments

	return nil
}

func (v ThirdPartyRiskAssessments) InsertThirdPartySnapshots(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	snapshot_third_parties AS (
		SELECT id, source_id
		FROM third_parties
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	)
INSERT INTO third_party_risk_assessments (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
	third_party_id,
	expires_at,
	data_sensitivity,
	business_impact,
	notes,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @third_party_risk_assessment_entity_type),
	@snapshot_id,
	vra.id,
	vra.organization_id,
	sv.id,
	vra.expires_at,
	vra.data_sensitivity,
	vra.business_impact,
	vra.notes,
	vra.created_at,
	vra.updated_at
FROM third_party_risk_assessments vra
INNER JOIN snapshot_third_parties sv ON sv.source_id = vra.third_party_id
WHERE %s AND vra.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
		"third_party_risk_assessment_entity_type": ThirdPartyRiskAssessmentEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty risk assessment snapshots: %w", err)
	}

	return nil
}
