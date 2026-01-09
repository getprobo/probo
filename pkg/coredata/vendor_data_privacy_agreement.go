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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	VendorDataPrivacyAgreement struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		VendorID       gid.GID    `db:"vendor_id"`
		ValidFrom      *time.Time `db:"valid_from"`
		ValidUntil     *time.Time `db:"valid_until"`
		FileID         gid.GID    `db:"file_id"`
		SnapshotID     *gid.GID   `db:"snapshot_id"`
		SourceID       *gid.GID   `db:"source_id"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	VendorDataPrivacyAgreements []*VendorDataPrivacyAgreement
)

func (v VendorDataPrivacyAgreement) CursorKey(orderBy VendorDataPrivacyAgreementOrderField) page.CursorKey {
	switch orderBy {
	case VendorDataPrivacyAgreementOrderFieldValidFrom:
		return page.NewCursorKey(v.ID, v.ValidFrom)
	case VendorDataPrivacyAgreementOrderFieldCreatedAt:
		return page.NewCursorKey(v.ID, v.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (vdpa *VendorDataPrivacyAgreement) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM vendor_data_privacy_agreements WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, vdpa.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query vendor data privacy agreement authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (vdpa *VendorDataPrivacyAgreement) LoadByVendorID(
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
	valid_from,
	valid_until,
	file_id,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	vendor_data_privacy_agreements
WHERE
	%s
	AND vendor_id = @vendor_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"vendor_id": vendorID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor data privacy agreement: %w", err)
	}

	vendorDataPrivacyAgreement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[VendorDataPrivacyAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect vendor data privacy agreement: %w", err)
	}

	*vdpa = vendorDataPrivacyAgreement

	return nil
}

func (vdpa *VendorDataPrivacyAgreement) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorDataPrivacyAgreementID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	vendor_id,
	valid_from,
	valid_until,
	file_id,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	vendor_data_privacy_agreements
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": vendorDataPrivacyAgreementID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query vendor data privacy agreement: %w", err)
	}

	vendorDataPrivacyAgreement, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[VendorDataPrivacyAgreement])
	if err != nil {
		return fmt.Errorf("cannot collect vendor data privacy agreement: %w", err)
	}

	*vdpa = vendorDataPrivacyAgreement

	return nil
}

func (vdpa *VendorDataPrivacyAgreement) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE
	vendor_data_privacy_agreements
SET
	valid_from = @valid_from,
	valid_until = @valid_until,
	file_id = @file_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":          vdpa.ID,
		"valid_from":  vdpa.ValidFrom,
		"valid_until": vdpa.ValidUntil,
		"file_id":     vdpa.FileID,
		"updated_at":  vdpa.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update vendor data privacy agreement: %w", err)
	}

	return nil
}

func (vdpa *VendorDataPrivacyAgreement) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
	vendor_data_privacy_agreements (
		id,
		tenant_id,
		organization_id,
		vendor_id,
		valid_from,
		valid_until,
		file_id,
		snapshot_id,
		source_id,
		created_at,
		updated_at
	)
VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@vendor_id,
	@valid_from,
	@valid_until,
	@file_id,
	@snapshot_id,
	@source_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id, vendor_id) DO UPDATE SET
	id = EXCLUDED.id,
	valid_from = EXCLUDED.valid_from,
	valid_until = EXCLUDED.valid_until,
	file_id = EXCLUDED.file_id,
	snapshot_id = EXCLUDED.snapshot_id,
	source_id = EXCLUDED.source_id,
	updated_at = EXCLUDED.updated_at
`
	args := pgx.StrictNamedArgs{
		"id":              vdpa.ID,
		"tenant_id":       scope.GetTenantID(),
		"vendor_id":       vdpa.VendorID,
		"organization_id": vdpa.OrganizationID,
		"valid_from":      vdpa.ValidFrom,
		"valid_until":     vdpa.ValidUntil,
		"file_id":         vdpa.FileID,
		"snapshot_id":     vdpa.SnapshotID,
		"source_id":       vdpa.SourceID,
		"created_at":      vdpa.CreatedAt,
		"updated_at":      vdpa.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "vendor_data_privacy_agreements_source_id_snapshot_id_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot upsert vendor data privacy agreement: %w", err)
	}
	return nil
}

func (vdpa *VendorDataPrivacyAgreement) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE
FROM
	vendor_data_privacy_agreements
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": vdpa.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (vdpa *VendorDataPrivacyAgreement) DeleteByVendorID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	vendorID gid.GID,
) error {
	q := `
DELETE
FROM
	vendor_data_privacy_agreements
WHERE
	%s
	AND vendor_id = @vendor_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"vendor_id": vendorID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (vdpa VendorDataPrivacyAgreements) InsertVendorSnapshots(
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
INSERT INTO vendor_data_privacy_agreements (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
	vendor_id,
	valid_from,
	valid_until,
	file_id,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @vendor_data_privacy_agreement_entity_type),
	@snapshot_id,
	vdpa.id,
	vdpa.organization_id,
	sv.id,
	vdpa.valid_from,
	vdpa.valid_until,
	vdpa.file_id,
	vdpa.created_at,
	vdpa.updated_at
FROM vendor_data_privacy_agreements vdpa
INNER JOIN snapshot_vendors sv ON sv.source_id = vdpa.vendor_id
WHERE %s AND vdpa.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
		"vendor_data_privacy_agreement_entity_type": VendorDataPrivacyAgreementEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert vendor data privacy agreement snapshots: %w", err)
	}

	return nil
}
