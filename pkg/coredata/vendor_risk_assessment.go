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
	// RiskAssessment represents a point-in-time risk assessment for a vendor
	VendorRiskAssessment struct {
		ID              gid.GID         `db:"id"`
		OrganizationID  gid.GID         `db:"organization_id"`
		VendorID        gid.GID         `db:"vendor_id"`
		ExpiresAt       time.Time       `db:"expires_at"`
		DataSensitivity DataSensitivity `db:"data_sensitivity"`
		BusinessImpact  BusinessImpact  `db:"business_impact"`
		Notes           *string         `db:"notes"`
		SnapshotID      *gid.GID        `db:"snapshot_id"`
		SourceID        *gid.GID        `db:"source_id"`
		CreatedAt       time.Time       `db:"created_at"`
		UpdatedAt       time.Time       `db:"updated_at"`
	}

	VendorRiskAssessments []*VendorRiskAssessment
)

func (v VendorRiskAssessment) CursorKey(orderBy VendorRiskAssessmentOrderField) page.CursorKey {
	switch orderBy {
	case VendorRiskAssessmentOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	case VendorRiskAssessmentOrderFieldExpiresAt:
		return page.NewCursorKey(v.ID, v.ExpiresAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (v *VendorRiskAssessment) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM vendor_risk_assessments WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query vendor risk assessment authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

// Insert adds a new risk assessment to the database
func (r VendorRiskAssessment) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    vendor_risk_assessments (
        tenant_id,
        id,
        organization_id,
        vendor_id,
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
    @vendor_id,
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
		"vendor_id":        r.VendorID,
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
func (r *VendorRiskAssessment) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    vendor_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    vendor_risk_assessments
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

	assessment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[VendorRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment: %w", err)
	}

	*r = assessment

	return nil
}

// LoadLatestByVendorID loads the most recent risk assessment for a vendor
func (r *VendorRiskAssessment) LoadLatestByVendorID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    vendor_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    vendor_risk_assessments
WHERE
    %s
    AND vendor_id = @vendor_id
ORDER BY
    created_at DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_id": vendorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessment: %w", err)
	}
	defer rows.Close()

	assessment, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[VendorRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessment: %w", err)
	}

	*r = assessment

	return nil
}

// LoadByVendorID loads all risk assessments for a vendor, ordered by assessment date
func (r *VendorRiskAssessments) LoadByVendorID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorID gid.GID,
	cursor *page.Cursor[VendorRiskAssessmentOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    vendor_id,
    expires_at,
    data_sensitivity,
    business_impact,
    notes,
    snapshot_id,
    source_id,
    created_at,
    updated_at
FROM
    vendor_risk_assessments
WHERE
    %s
    AND vendor_id = @vendor_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_id": vendorID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk assessments: %w", err)
	}

	assessments, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[VendorRiskAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect risk assessments: %w", err)
	}

	*r = assessments

	return nil
}

func (v VendorRiskAssessments) InsertVendorSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH
	snapshot_vendors AS (
		SELECT id, source_id
		FROM vendors
		WHERE organization_id = @organization_id AND snapshot_id = @snapshot_id
	)
INSERT INTO vendor_risk_assessments (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
	vendor_id,
	expires_at,
	data_sensitivity,
	business_impact,
	notes,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @vendor_risk_assessment_entity_type),
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
FROM vendor_risk_assessments vra
INNER JOIN snapshot_vendors sv ON sv.source_id = vra.vendor_id
WHERE %s AND vra.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                          scope.GetTenantID(),
		"snapshot_id":                        snapshotID,
		"organization_id":                    organizationID,
		"vendor_risk_assessment_entity_type": VendorRiskAssessmentEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert vendor risk assessment snapshots: %w", err)
	}

	return nil
}
