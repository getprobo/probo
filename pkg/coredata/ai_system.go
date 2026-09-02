// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"errors"
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
	AiSystem struct {
		ID                      gid.GID                     `db:"id"`
		OrganizationID          gid.GID                     `db:"organization_id"`
		Name                    string                      `db:"name"`
		Version                 *string                     `db:"version"`
		CompanyRoles            []AiSystemCompanyRole       `db:"company_roles"`
		Status                  AiSystemStatus              `db:"status"`
		OwnerID                 *gid.GID                    `db:"owner_id"`
		Source                  *string                     `db:"source"`
		Purpose                 *string                     `db:"purpose"`
		IntendedUseCases        *string                     `db:"intended_use_cases"`
		AutonomyLevel           *string                     `db:"autonomy_level"`
		HumanOversightMechanism *string                     `db:"human_oversight_mechanism"`
		RiskClassification      *AiSystemRiskClassification `db:"risk_classification"`
		KeyStakeholders         *string                     `db:"key_stakeholders"`
		DataSourcesAndType      *string                     `db:"data_sources_and_type"`
		DeploymentDate          *time.Time                  `db:"deployment_date"`
		LastReviewDate          *time.Time                  `db:"last_review_date"`
		NextReviewDate          *time.Time                  `db:"next_review_date"`
		Notes                   *string                     `db:"notes"`
		CreatedAt               time.Time                   `db:"created_at"`
		UpdatedAt               time.Time                   `db:"updated_at"`
	}

	AiSystems []*AiSystem
)

func (s *AiSystem) CursorKey(field AiSystemOrderField) page.CursorKey {
	switch field {
	case AiSystemOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case AiSystemOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	case AiSystemOrderFieldStatus:
		return page.NewCursorKey(s.ID, s.Status)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (s *AiSystem) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM ai_systems WHERE id = ANY(@resource_ids::text[])`

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

func (s *AiSystem) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	aiSystemID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	version,
	company_roles,
	status,
	owner_id,
	source,
	purpose,
	intended_use_cases,
	autonomy_level,
	human_oversight_mechanism,
	risk_classification,
	key_stakeholders,
	data_sources_and_type,
	deployment_date,
	last_review_date,
	next_review_date,
	notes,
	created_at,
	updated_at
FROM
	ai_systems
WHERE
	%s
	AND id = @ai_system_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"ai_system_id": aiSystemID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query ai system: %w", err)
	}

	aiSystem, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AiSystem])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect ai system: %w", err)
	}

	*s = aiSystem

	return nil
}

func (ss *AiSystems) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	filter *AiSystemFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	ai_systems
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
		return 0, fmt.Errorf("cannot count ai systems: %w", err)
	}

	return count, nil
}

func (ss *AiSystems) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AiSystemOrderField],
	filter *AiSystemFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	version,
	company_roles,
	status,
	owner_id,
	source,
	purpose,
	intended_use_cases,
	autonomy_level,
	human_oversight_mechanism,
	risk_classification,
	key_stakeholders,
	data_sources_and_type,
	deployment_date,
	last_review_date,
	next_review_date,
	notes,
	created_at,
	updated_at
FROM
	ai_systems
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
		return fmt.Errorf("cannot query ai systems: %w", err)
	}

	aiSystems, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AiSystem])
	if err != nil {
		return fmt.Errorf("cannot collect ai systems: %w", err)
	}

	*ss = aiSystems

	return nil
}

func (s *AiSystem) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO ai_systems (
	id,
	tenant_id,
	organization_id,
	name,
	version,
	company_roles,
	status,
	owner_id,
	source,
	purpose,
	intended_use_cases,
	autonomy_level,
	human_oversight_mechanism,
	risk_classification,
	key_stakeholders,
	data_sources_and_type,
	deployment_date,
	last_review_date,
	next_review_date,
	notes,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@name,
	@version,
	COALESCE(@company_roles, '{}'::TEXT[]),
	@status,
	@owner_id,
	@source,
	@purpose,
	@intended_use_cases,
	@autonomy_level,
	@human_oversight_mechanism,
	@risk_classification,
	@key_stakeholders,
	@data_sources_and_type,
	@deployment_date,
	@last_review_date,
	@next_review_date,
	@notes,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"tenant_id":                 scope.GetTenantID(),
		"organization_id":           s.OrganizationID,
		"name":                      s.Name,
		"version":                   s.Version,
		"company_roles":             aiSystemCompanyRolesToStrings(s.CompanyRoles),
		"status":                    s.Status,
		"owner_id":                  s.OwnerID,
		"source":                    s.Source,
		"purpose":                   s.Purpose,
		"intended_use_cases":        s.IntendedUseCases,
		"autonomy_level":            s.AutonomyLevel,
		"human_oversight_mechanism": s.HumanOversightMechanism,
		"risk_classification":       s.RiskClassification,
		"key_stakeholders":          s.KeyStakeholders,
		"data_sources_and_type":     s.DataSourcesAndType,
		"deployment_date":           s.DeploymentDate,
		"last_review_date":          s.LastReviewDate,
		"next_review_date":          s.NextReviewDate,
		"notes":                     s.Notes,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert ai system: %w", err)
	}

	return nil
}

func (s *AiSystem) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE ai_systems
SET
	name = @name,
	version = @version,
	company_roles = COALESCE(@company_roles, '{}'::TEXT[]),
	status = @status,
	owner_id = @owner_id,
	source = @source,
	purpose = @purpose,
	intended_use_cases = @intended_use_cases,
	autonomy_level = @autonomy_level,
	human_oversight_mechanism = @human_oversight_mechanism,
	risk_classification = @risk_classification,
	key_stakeholders = @key_stakeholders,
	data_sources_and_type = @data_sources_and_type,
	deployment_date = @deployment_date,
	last_review_date = @last_review_date,
	next_review_date = @next_review_date,
	notes = @notes,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"name":                      s.Name,
		"version":                   s.Version,
		"company_roles":             aiSystemCompanyRolesToStrings(s.CompanyRoles),
		"status":                    s.Status,
		"owner_id":                  s.OwnerID,
		"source":                    s.Source,
		"purpose":                   s.Purpose,
		"intended_use_cases":        s.IntendedUseCases,
		"autonomy_level":            s.AutonomyLevel,
		"human_oversight_mechanism": s.HumanOversightMechanism,
		"risk_classification":       s.RiskClassification,
		"key_stakeholders":          s.KeyStakeholders,
		"data_sources_and_type":     s.DataSourcesAndType,
		"deployment_date":           s.DeploymentDate,
		"last_review_date":          s.LastReviewDate,
		"next_review_date":          s.NextReviewDate,
		"notes":                     s.Notes,
		"updated_at":                s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update ai system: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *AiSystem) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM ai_systems
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete ai system: %w", err)
	}

	return nil
}

func (s AiSystem) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	ai_systems_document_id
FROM
	generated_documents
WHERE
	organization_id = @organization_id
`,
		pgx.NamedArgs{"organization_id": organizationID},
	).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("cannot get ai system list document ID: %w", err)
	}

	return documentID, nil
}

func (s AiSystem) UpsertGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	documentID gid.GID,
) error {
	now := time.Now()

	_, err := conn.Exec(
		ctx,
		`
INSERT INTO generated_documents (
	organization_id,
	tenant_id,
	ai_systems_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@ai_systems_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	ai_systems_document_id = @ai_systems_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id":        organizationID,
			"tenant_id":              tenantID,
			"ai_systems_document_id": documentID,
			"created_at":             now,
			"updated_at":             now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert ai system list document ID: %w", err)
	}

	return nil
}

func (s AiSystem) ClearGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	documentIDs []gid.GID,
) error {
	ids := make([]string, len(documentIDs))
	for i, id := range documentIDs {
		ids[i] = id.String()
	}

	_, err := conn.Exec(
		ctx,
		`
UPDATE
	generated_documents
SET
	ai_systems_document_id = NULL,
	updated_at = @now
WHERE
	ai_systems_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear ai system list document references: %w", err)
	}

	return nil
}

func aiSystemCompanyRolesToStrings(roles []AiSystemCompanyRole) []string {
	if len(roles) == 0 {
		return nil
	}

	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}

	return result
}

func (c *AiSystems) DeleteByOrganizationID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
DELETE FROM ai_systems
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete ai systems: %w", err)
	}

	return nil
}
