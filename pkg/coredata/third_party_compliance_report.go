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
	ThirdPartyComplianceReport struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		ThirdPartyID   gid.GID    `db:"third_party_id"`
		ReportDate     time.Time  `db:"report_date"`
		ValidUntil     *time.Time `db:"valid_until"`
		ReportName     string     `db:"report_name"`
		ReportFileId   *gid.GID   `db:"report_file_id"`
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

func (v *ThirdPartyComplianceReport) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM third_party_compliance_reports WHERE id = ANY(@resource_ids::text[])`

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
		err := rows.Scan(&id, &organizationID)
		if err != nil {
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

func (vcs *ThirdPartyComplianceReports) LoadByThirdPartyIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	thirdPartyIDs []gid.GID,
) error {
	if len(thirdPartyIDs) == 0 {
		*vcs = ThirdPartyComplianceReports{}
		return nil
	}

	q := `
SELECT
	id,
	organization_id,
	third_party_id,
	report_date,
	valid_until,
	report_name,
	report_file_id,
	created_at,
	updated_at
FROM
	third_party_compliance_reports
WHERE
	%s
	AND third_party_id = ANY(@third_party_ids)
ORDER BY
	third_party_id, report_date DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	ids := make([]string, len(thirdPartyIDs))
	for i, id := range thirdPartyIDs {
		ids[i] = id.String()
	}

	args := pgx.NamedArgs{"third_party_ids": ids}
	maps.Copy(args, scope.SQLArguments())

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
