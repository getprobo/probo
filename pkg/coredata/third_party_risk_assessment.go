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
	"go.probo.inc/probo/pkg/iam/policy"
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

func (v *ThirdPartyRiskAssessment) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_party_risk_assessments WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
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

func (r *ThirdPartyRiskAssessments) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*r = ThirdPartyRiskAssessments{}
		return nil
	}

	q := `
SELECT
    id,
    organization_id,
    third_party_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    created_at,
    updated_at
FROM
    third_party_risk_assessments
WHERE
    %s
    AND third_party_id = ANY(@third_party_ids)
ORDER BY
    third_party_id, created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]string, len(thirdPartyIDs))
	for i, id := range thirdPartyIDs {
		ids[i] = id.String()
	}

	args := pgx.StrictNamedArgs{"third_party_ids": ids}
	maps.Copy(args, scope.SQLArguments())

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
