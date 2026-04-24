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
	ThirdPartyComplianceReport struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		ThirdPartyID   gid.GID    `db:"third_party_id"`
		ReportDate     time.Time  `db:"report_date"`
		ValidUntil     *time.Time `db:"valid_until"`
		ReportName     string     `db:"report_name"`
		ReportFileId   *gid.GID   `db:"report_file_id"`
		SnapshotID     *gid.GID   `db:"snapshot_id"`
		SourceID       *gid.GID   `db:"source_id"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	ThirdPartyComplianceReports []*ThirdPartyComplianceReport
)

func (c ThirdPartyComplianceReport) CursorKey(orderBy ThirdPartyComplianceReportOrderField) page.CursorKey {
	switch orderBy {
	case ThirdPartyComplianceReportOrderFieldReportDate:
		return page.NewCursorKey(c.ID, c.ReportDate)
	case ThirdPartyComplianceReportOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (v *ThirdPartyComplianceReport) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM third_party_compliance_reports WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, v.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query thirdParty compliance report authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (vcs *ThirdPartyComplianceReports) LoadForThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyID gid.GID,
	cursor *page.Cursor[ThirdPartyComplianceReportOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	report_date,
	valid_until,
	report_name,
	report_file_id,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	third_party_compliance_reports
WHERE
	%s
	AND third_party_id = @third_party_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"third_party_id": thirdPartyID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty compliance reports: %w", err)
	}

	thirdPartyComplianceReports, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ThirdPartyComplianceReport])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty compliance reports: %w", err)
	}

	*vcs = thirdPartyComplianceReports

	return nil
}

func (vcr *ThirdPartyComplianceReport) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyComplianceReportID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	report_date,
	valid_until,
	report_name,
	report_file_id,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	third_party_compliance_reports
WHERE
	%s
	AND id = @id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.NamedArgs{"id": thirdPartyComplianceReportID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query thirdParty compliance report: %w", err)
	}

	thirdPartyComplianceReport, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ThirdPartyComplianceReport])
	if err != nil {
		return fmt.Errorf("cannot collect thirdParty compliance report: %w", err)
	}

	*vcr = thirdPartyComplianceReport

	return nil
}

func (vcr *ThirdPartyComplianceReport) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
	third_party_compliance_reports (
		id,
		organization_id,
		tenant_id,
		third_party_id,
		report_date,
		valid_until,
		report_name,
		report_file_id,
		created_at,
		updated_at
	)
VALUES (
	@id,
	@organization_id,
	@tenant_id,
	@third_party_id,
	@report_date,
	@valid_until,
	@report_name,
	@report_file_id,
	@created_at,
	@updated_at
)
`
	args := pgx.NamedArgs{
		"id":              vcr.ID,
		"organization_id": vcr.OrganizationID,
		"tenant_id":       scope.GetTenantID(),
		"third_party_id":  vcr.ThirdPartyID,
		"report_date":     vcr.ReportDate,
		"valid_until":     vcr.ValidUntil,
		"report_name":     vcr.ReportName,
		"report_file_id":  vcr.ReportFileId,
		"created_at":      vcr.CreatedAt,
		"updated_at":      vcr.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (vcr *ThirdPartyComplianceReport) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE
FROM
	third_party_compliance_reports
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
RETURNING report_file_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": vcr.ID}
	maps.Copy(args, scope.SQLArguments())

	var vcrFileId *gid.GID
	err := conn.QueryRow(ctx, q, args).Scan(&vcrFileId)

	if err != nil {
		return fmt.Errorf("cannot delete thirdParty compliance report: %w", err)
	}

	if vcrFileId != nil {
		file := &File{ID: *vcrFileId}
		if err = file.SoftDelete(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot soft delete thirdParty compliance file: %w", err)
		}
	}
	return nil
}

func (vcrs ThirdPartyComplianceReports) InsertThirdPartySnapshots(
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
INSERT INTO third_party_compliance_reports (
	tenant_id,
	id,
	organization_id,
	snapshot_id,
	source_id,
	third_party_id,
	report_date,
	valid_until,
	report_name,
	report_file_id,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @third_party_compliance_report_entity_type),
	@organization_id,
	@snapshot_id,
	vcr.id,
	sv.id,
	vcr.report_date,
	vcr.valid_until,
	vcr.report_name,
	vcr.report_file_id,
	vcr.created_at,
	vcr.updated_at
FROM third_party_compliance_reports vcr
INNER JOIN snapshot_third_parties sv ON sv.source_id = vcr.third_party_id
WHERE %s AND vcr.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
		"third_party_compliance_report_entity_type": ThirdPartyComplianceReportEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert thirdParty compliance report snapshots: %w", err)
	}

	return nil
}
