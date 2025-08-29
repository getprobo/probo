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
	ProcessingActivityRegistry struct {
		ID                             gid.GID                                                  `db:"id"`
		SnapshotID                     *gid.GID                                                 `db:"snapshot_id"`
		SourceID                       *gid.GID                                                 `db:"source_id"`
		OrganizationID                 gid.GID                                                  `db:"organization_id"`
		Name                           string                                                   `db:"name"`
		Purpose                        *string                                                  `db:"purpose"`
		DataSubjectCategory            *string                                                  `db:"data_subject_category"`
		PersonalDataCategory           *string                                                  `db:"personal_data_category"`
		SpecialOrCriminalData          ProcessingActivityRegistrySpecialOrCriminalData          `db:"special_or_criminal_data"`
		ConsentEvidenceLink            *string                                                  `db:"consent_evidence_link"`
		LawfulBasis                    ProcessingActivityRegistryLawfulBasis                    `db:"lawful_basis"`
		Recipients                     *string                                                  `db:"recipients"`
		Location                       *string                                                  `db:"location"`
		InternationalTransfers         bool                                                     `db:"international_transfers"`
		TransferSafeguards             *ProcessingActivityRegistryTransferSafeguards            `db:"transfer_safeguards"`
		RetentionPeriod                *string                                                  `db:"retention_period"`
		SecurityMeasures               *string                                                  `db:"security_measures"`
		DataProtectionImpactAssessment ProcessingActivityRegistryDataProtectionImpactAssessment `db:"data_protection_impact_assessment"`
		TransferImpactAssessment       ProcessingActivityRegistryTransferImpactAssessment       `db:"transfer_impact_assessment"`
		CreatedAt                      time.Time                                                `db:"created_at"`
		UpdatedAt                      time.Time                                                `db:"updated_at"`
	}

	ProcessingActivityRegistries []*ProcessingActivityRegistry
)

func (p *ProcessingActivityRegistry) CursorKey(field ProcessingActivityRegistryOrderField) page.CursorKey {
	switch field {
	case ProcessingActivityRegistryOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case ProcessingActivityRegistryOrderFieldName:
		return page.NewCursorKey(p.ID, p.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (p *ProcessingActivityRegistry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	processingActivityRegistryID gid.GID,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment,
	transfer_impact_assessment,
	created_at,
	updated_at
FROM
	processing_activity_registries
WHERE
	%s
	AND id = @processing_activity_registry_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"processing_activity_registry_id": processingActivityRegistryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activity registry: %w", err)
	}

	processingActivityRegistry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivityRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect processing activity registry: %w", err)
	}

	*p = processingActivityRegistry

	return nil
}

func (p *ProcessingActivityRegistries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ProcessingActivityRegistryFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activity_registries
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count processing activity registries: %w", err)
	}

	return count, nil
}

func (p *ProcessingActivityRegistries) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ProcessingActivityRegistryOrderField],
	filter *ProcessingActivityRegistryFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment,
	transfer_impact_assessment,
	created_at,
	updated_at
FROM
	processing_activity_registries
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
		return fmt.Errorf("cannot query processing activity registries: %w", err)
	}

	processingActivityRegistries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ProcessingActivityRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect processing activity registries: %w", err)
	}

	*p = processingActivityRegistries

	return nil
}

func (p *ProcessingActivityRegistry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO processing_activity_registries (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment,
	transfer_impact_assessment,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@snapshot_id,
	@source_id,
	@organization_id,
	@name,
	@purpose,
	@data_subject_category,
	@personal_data_category,
	@special_or_criminal_data,
	@consent_evidence_link,
	@lawful_basis,
	@recipients,
	@location,
	@international_transfers,
	@transfer_safeguards,
	@retention_period,
	@security_measures,
	@data_protection_impact_assessment,
	@transfer_impact_assessment,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                                p.ID,
		"tenant_id":                         scope.GetTenantID(),
		"snapshot_id":                       p.SnapshotID,
		"source_id":                         p.SourceID,
		"organization_id":                   p.OrganizationID,
		"name":                              p.Name,
		"purpose":                           p.Purpose,
		"data_subject_category":             p.DataSubjectCategory,
		"personal_data_category":            p.PersonalDataCategory,
		"special_or_criminal_data":          p.SpecialOrCriminalData,
		"consent_evidence_link":             p.ConsentEvidenceLink,
		"lawful_basis":                      p.LawfulBasis,
		"recipients":                        p.Recipients,
		"location":                          p.Location,
		"international_transfers":           p.InternationalTransfers,
		"transfer_safeguards":               p.TransferSafeguards,
		"retention_period":                  p.RetentionPeriod,
		"security_measures":                 p.SecurityMeasures,
		"data_protection_impact_assessment": p.DataProtectionImpactAssessment,
		"transfer_impact_assessment":        p.TransferImpactAssessment,
		"created_at":                        p.CreatedAt,
		"updated_at":                        p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity registry: %w", err)
	}

	return nil
}

func (p *ProcessingActivityRegistry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE processing_activity_registries
SET
	name = @name,
	purpose = @purpose,
	data_subject_category = @data_subject_category,
	personal_data_category = @personal_data_category,
	special_or_criminal_data = @special_or_criminal_data,
	consent_evidence_link = @consent_evidence_link,
	lawful_basis = @lawful_basis,
	recipients = @recipients,
	location = @location,
	international_transfers = @international_transfers,
	transfer_safeguards = @transfer_safeguards,
	retention_period = @retention_period,
	security_measures = @security_measures,
	data_protection_impact_assessment = @data_protection_impact_assessment,
	transfer_impact_assessment = @transfer_impact_assessment,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                                p.ID,
		"name":                              p.Name,
		"purpose":                           p.Purpose,
		"data_subject_category":             p.DataSubjectCategory,
		"personal_data_category":            p.PersonalDataCategory,
		"special_or_criminal_data":          p.SpecialOrCriminalData,
		"consent_evidence_link":             p.ConsentEvidenceLink,
		"lawful_basis":                      p.LawfulBasis,
		"recipients":                        p.Recipients,
		"location":                          p.Location,
		"international_transfers":           p.InternationalTransfers,
		"transfer_safeguards":               p.TransferSafeguards,
		"retention_period":                  p.RetentionPeriod,
		"security_measures":                 p.SecurityMeasures,
		"data_protection_impact_assessment": p.DataProtectionImpactAssessment,
		"transfer_impact_assessment":        p.TransferImpactAssessment,
		"updated_at":                        p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update processing activity registry: %w", err)
	}

	return nil
}

func (p *ProcessingActivityRegistry) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM processing_activity_registries
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete processing activity registry: %w", err)
	}

	return nil
}

func (pars ProcessingActivityRegistries) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO processing_activity_registries (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment,
	transfer_impact_assessment,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @processing_activity_registry_entity_type),
	@tenant_id,
	@snapshot_id,
	par.id,
	par.organization_id,
	par.name,
	par.purpose,
	par.data_subject_category,
	par.personal_data_category,
	par.special_or_criminal_data,
	par.consent_evidence_link,
	par.lawful_basis,
	par.recipients,
	par.location,
	par.international_transfers,
	par.transfer_safeguards,
	par.retention_period,
	par.security_measures,
	par.data_protection_impact_assessment,
	par.transfer_impact_assessment,
	par.created_at,
	par.updated_at
FROM processing_activity_registries par
WHERE %s AND par.organization_id = @organization_id AND par.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
		"processing_activity_registry_entity_type": ProcessingActivityRegistryEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity registry snapshots: %w", err)
	}

	return nil
}
