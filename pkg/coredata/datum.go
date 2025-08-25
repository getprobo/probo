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
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Datum struct {
		ID                 gid.GID            `db:"id"`
		Name               string             `db:"name"`
		OrganizationID     gid.GID            `db:"organization_id"`
		OwnerID            gid.GID            `db:"owner_id"`
		DataClassification DataClassification `db:"data_classification"`
		SnapshotID         *gid.GID           `db:"snapshot_id"`
		OriginalID         *gid.GID           `db:"original_id"`
		CreatedAt          time.Time          `db:"created_at"`
		UpdatedAt          time.Time          `db:"updated_at"`
	}

	Data []*Datum

	idMap map[gid.GID]gid.GID
)

func (d *Datum) CursorKey(field DatumOrderField) page.CursorKey {
	switch field {
	case DatumOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	case DatumOrderFieldName:
		return page.NewCursorKey(d.ID, d.Name)
	case DatumOrderFieldDataClassification:
		return page.NewCursorKey(d.ID, d.DataClassification)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (d *Datum) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	dataID gid.GID,
) error {
	q := `
SELECT
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	original_id,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND id = @data_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"data_id": dataID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	datum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = datum

	return nil
}

func (d *Datum) LoadByOwnerID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	original_id,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND owner_id = @owner_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"owner_id": d.OwnerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d *Data) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count data: %w", err)
	}

	return count, nil
}

func (d *Data) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DatumOrderField],
	filter *DatumFilter,
) error {
	q := `
SELECT
	id,
	name,
	organization_id,
	owner_id,
	data_classification,
	snapshot_id,
	original_id,
	created_at,
	updated_at
FROM
	data
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data: %w", err)
	}

	data, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect data: %w", err)
	}

	*d = data

	return nil
}

func (d Data) BulkInsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	columnNames := []string{
		"id",
		"tenant_id",
		"name",
		"owner_id",
		"organization_id",
		"data_classification",
		"snapshot_id",
		"original_id",
		"created_at",
		"updated_at",
	}

	copyFromSource := &datumCopy{
		data:     d,
		scope:    scope,
		position: 0,
	}

	_, err := conn.CopyFrom(ctx, pgx.Identifier{"data"}, columnNames, copyFromSource)
	if err != nil {
		return fmt.Errorf("cannot bulk insert data: %w", err)
	}

	return nil
}

func (d *Datum) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO data (
	id,
	tenant_id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	original_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@name,
	@owner_id,
	@organization_id,
	@data_classification,
	@snapshot_id,
	@original_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"tenant_id":           scope.GetTenantID(),
		"name":                d.Name,
		"owner_id":            d.OwnerID,
		"organization_id":     d.OrganizationID,
		"data_classification": d.DataClassification,
		"snapshot_id":         d.SnapshotID,
		"original_id":         d.OriginalID,
		"created_at":          d.CreatedAt,
		"updated_at":          d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert data: %w", err)
	}

	return nil
}

func (d *Datum) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE data
SET
	name = @name,
	owner_id = @owner_id,
	data_classification = @data_classification,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
RETURNING
	id,
	name,
	owner_id,
	organization_id,
	data_classification,
	snapshot_id,
	original_id,
	created_at,
	updated_at
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  d.ID,
		"name":                d.Name,
		"owner_id":            d.OwnerID,
		"data_classification": d.DataClassification,
		"updated_at":          d.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update data: %w", err)
	}

	datum, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Datum])
	if err != nil {
		return fmt.Errorf("cannot collect updated data: %w", err)
	}

	*d = datum

	return nil
}

func (d *Datum) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM data
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": d.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete data: %w", err)
	}

	return nil
}

func (d Data) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	currentDatumIDs, dataIDMap, err := d.snapshotData(
		ctx,
		conn,
		scope,
		organizationID,
		snapshotID,
	)
	if err != nil {
		return fmt.Errorf("cannot snapshot data: %w", err)
	}

	vendorIDMap, err := d.snapshotVendors(ctx, conn, scope, currentDatumIDs, snapshotID)
	if err != nil {
		return fmt.Errorf("cannot snapshot vendors: %w", err)
	}

	if err := d.snapshotDatumVendors(
		ctx,
		conn,
		scope,
		currentDatumIDs,
		snapshotID,
		dataIDMap,
		vendorIDMap,
	); err != nil {
		return fmt.Errorf("cannot snapshot datum vendors: %w", err)
	}

	return nil
}

func (d Data) snapshotData(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) ([]gid.GID, idMap, error) {
	var nilSnapshotID *gid.GID = nil
	filter := NewDatumFilterBySnapshotID(&nilSnapshotID)
	maxRows := 1000 // Use batch in the future

	cursor := page.NewCursor(maxRows, nil, page.Head, page.OrderBy[DatumOrderField]{
		Field:     DatumOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	})

	currentData := Data{}
	if err := currentData.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor, filter); err != nil {
		return nil, nil, fmt.Errorf("cannot load data without snapshot_id: %w", err)
	}

	var currentDatumIDs []gid.GID
	var snapshotData Data
	dataIDMap := make(idMap)

	for _, currentDatum := range currentData {
		currentDatumIDs = append(currentDatumIDs, currentDatum.ID)
		snapshotDatumID := gid.New(scope.GetTenantID(), DatumEntityType)
		dataIDMap[currentDatum.ID] = snapshotDatumID

		snapshotDatum := &Datum{
			ID:                 snapshotDatumID,
			SnapshotID:         &snapshotID,
			OriginalID:         &currentDatum.ID,
			Name:               currentDatum.Name,
			OrganizationID:     currentDatum.OrganizationID,
			OwnerID:            currentDatum.OwnerID,
			DataClassification: currentDatum.DataClassification,
			CreatedAt:          currentDatum.CreatedAt,
			UpdatedAt:          currentDatum.UpdatedAt,
		}

		snapshotData = append(snapshotData, snapshotDatum)
	}

	if err := snapshotData.BulkInsert(ctx, conn, scope); err != nil {
		return nil, nil, fmt.Errorf("cannot bulk insert data copies: %w", err)
	}

	return currentDatumIDs, dataIDMap, nil
}

func (d Data) snapshotVendors(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	currentDatumIDs []gid.GID,
	snapshotID gid.GID,
) (idMap, error) {
	var currentVendors Vendors
	if err := currentVendors.LoadByDatumIDs(ctx, conn, scope, currentDatumIDs); err != nil {
		return nil, fmt.Errorf("cannot load vendors by datum IDs: %w", err)
	}

	vendorIDMap := make(idMap)
	var snapshotVendors Vendors

	for _, currentVendor := range currentVendors {
		snapshotVendorID := gid.New(scope.GetTenantID(), VendorEntityType)
		vendorIDMap[currentVendor.ID] = snapshotVendorID

		snapshotVendor := &Vendor{
			ID:                            snapshotVendorID,
			TenantID:                      currentVendor.TenantID,
			OriginalID:                    &currentVendor.ID,
			OrganizationID:                currentVendor.OrganizationID,
			Name:                          currentVendor.Name,
			Description:                   currentVendor.Description,
			Category:                      currentVendor.Category,
			HeadquarterAddress:            currentVendor.HeadquarterAddress,
			LegalName:                     currentVendor.LegalName,
			WebsiteURL:                    currentVendor.WebsiteURL,
			PrivacyPolicyURL:              currentVendor.PrivacyPolicyURL,
			ServiceLevelAgreementURL:      currentVendor.ServiceLevelAgreementURL,
			DataProcessingAgreementURL:    currentVendor.DataProcessingAgreementURL,
			BusinessAssociateAgreementURL: currentVendor.BusinessAssociateAgreementURL,
			SubprocessorsListURL:          currentVendor.SubprocessorsListURL,
			Certifications:                currentVendor.Certifications,
			BusinessOwnerID:               currentVendor.BusinessOwnerID,
			SecurityOwnerID:               currentVendor.SecurityOwnerID,
			StatusPageURL:                 currentVendor.StatusPageURL,
			TermsOfServiceURL:             currentVendor.TermsOfServiceURL,
			SecurityPageURL:               currentVendor.SecurityPageURL,
			TrustPageURL:                  currentVendor.TrustPageURL,
			ShowOnTrustCenter:             currentVendor.ShowOnTrustCenter,
			SnapshotID:                    &snapshotID,
			CreatedAt:                     currentVendor.CreatedAt,
			UpdatedAt:                     currentVendor.UpdatedAt,
		}

		snapshotVendors = append(snapshotVendors, snapshotVendor)
	}

	if err := snapshotVendors.BulkInsert(ctx, conn, scope); err != nil {
		return nil, fmt.Errorf("cannot bulk insert vendor snapshots: %w", err)
	}

	return vendorIDMap, nil
}

func (d Data) snapshotDatumVendors(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	currentDatumIDs []gid.GID,
	snapshotID gid.GID,
	dataIDMap idMap,
	vendorIDMap idMap,
) error {
	var datumVendors DatumVendors
	if err := datumVendors.LoadByDatumIDs(ctx, conn, scope, currentDatumIDs); err != nil {
		return fmt.Errorf("cannot load datum vendors: %w", err)
	}

	for _, datumVendor := range datumVendors {
		datumVendor.DatumID = dataIDMap[datumVendor.DatumID]
		datumVendor.VendorID = vendorIDMap[datumVendor.VendorID]
		datumVendor.SnapshotID = &snapshotID
	}

	if err := datumVendors.BulkInsert(ctx, conn, scope); err != nil {
		return fmt.Errorf("cannot bulk insert datum vendors: %w", err)
	}

	return nil
}
