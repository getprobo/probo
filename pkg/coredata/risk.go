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
	Risk struct {
		ID                 gid.GID       `db:"id"`
		OrganizationID     gid.GID       `db:"organization_id"`
		Name               string        `db:"name"`
		Description        *string       `db:"description"`
		Category           string        `db:"category"`
		Treatment          RiskTreatment `db:"treatment"`
		Note               string        `db:"note"`
		OwnerID            *gid.GID      `db:"owner_profile_id"`
		InherentLikelihood int           `db:"inherent_likelihood"`
		InherentImpact     int           `db:"inherent_impact"`
		InherentRiskScore  int           `db:"inherent_risk_score"`
		ResidualLikelihood int           `db:"residual_likelihood"`
		ResidualImpact     int           `db:"residual_impact"`
		ResidualRiskScore  int           `db:"residual_risk_score"`
		SnapshotID         *gid.GID      `db:"snapshot_id"`
		SourceID           *gid.GID      `db:"source_id"`
		CreatedAt          time.Time     `db:"created_at"`
		UpdatedAt          time.Time     `db:"updated_at"`

		// Ordering only
		OwnerFullName *string `db:"owner_full_name"`
	}

	Risks []*Risk

	RiskSnapshotter interface {
		InsertRiskSnapshots(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error
	}
)

func (r *Risk) CursorKey(orderBy RiskOrderField) page.CursorKey {
	switch orderBy {
	case RiskOrderFieldCreatedAt:
		return page.CursorKey{ID: r.ID, Value: r.CreatedAt}
	case RiskOrderFieldName:
		return page.CursorKey{ID: r.ID, Value: r.Name}
	case RiskOrderFieldCategory:
		return page.CursorKey{ID: r.ID, Value: r.Category}
	case RiskOrderFieldTreatment:
		return page.CursorKey{ID: r.ID, Value: r.Treatment}
	case RiskOrderFieldInherentRiskScore:
		return page.CursorKey{ID: r.ID, Value: r.InherentRiskScore}
	case RiskOrderFieldResidualRiskScore:
		return page.CursorKey{ID: r.ID, Value: r.ResidualRiskScore}
	case RiskOrderFieldOwnerFullName:
		return page.CursorKey{ID: r.ID, Value: r.OwnerFullName}
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (r *Risk) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM risks WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, r.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query risk authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (r *Risks) CountByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	filter *RiskFilter,
) (int, error) {
	q := `
WITH rsks AS (
	SELECT
		r.id,
		r.tenant_id,
		r.search_vector,
		r.snapshot_id
	FROM
		risks r
	INNER JOIN
		risks_measures rm ON r.id = rm.risk_id
	WHERE
		rm.measure_id = @measure_id
)
SELECT
	COUNT(id)
FROM
	rsks
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (r *Risks) LoadByMeasureID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	measureID gid.GID,
	cursor *page.Cursor[RiskOrderField],
	filter *RiskFilter,
) error {
	q := `
WITH rsks AS (
	SELECT
		r.id,
		r.tenant_id,
		r.organization_id,
		r.name,
		r.description,
		r.category,
		r.owner_profile_id,
		p.full_name as owner_full_name,
		r.treatment,
		r.note,
		r.inherent_likelihood,
		r.inherent_impact,
		r.inherent_risk_score,
		r.residual_likelihood,
		r.residual_impact,
		r.residual_risk_score,
		r.snapshot_id,
		r.source_id,
		r.search_vector,
		r.created_at,
		r.updated_at
	FROM
		risks r
	INNER JOIN
		risks_measures rm ON r.id = rm.risk_id
	LEFT JOIN
		iam_membership_profiles p ON r.owner_profile_id = p.id
	WHERE
		rm.measure_id = @measure_id
)
SELECT
	id,
	organization_id,
	name,
	description,
	category,
	owner_profile_id,
	owner_full_name,
	treatment,
	note,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	rsks
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"measure_id": measureID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risks: %w", err)
	}

	risks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Risk])
	if err != nil {
		return fmt.Errorf("cannot collect risks: %w", err)
	}

	*r = risks

	return nil
}

func (r *Risks) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *RiskFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM risks
WHERE %s
	AND organization_id = @organization_id
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (r *Risks) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[RiskOrderField],
	filter *RiskFilter,
) error {
	q := `
WITH rsks AS (
	SELECT
		r.id,
		r.tenant_id,
		r.organization_id,
		r.name,
		r.description,
		r.owner_profile_id,
		p.full_name as owner_full_name,
		r.treatment,
		r.note,
		r.inherent_likelihood,
		r.inherent_impact,
		r.inherent_risk_score,
		r.residual_likelihood,
		r.residual_impact,
		r.residual_risk_score,
		r.category,
		r.snapshot_id,
		r.source_id,
		r.search_vector,
		r.created_at,
		r.updated_at
	FROM
		risks r
	LEFT JOIN
		iam_membership_profiles p ON r.owner_profile_id = p.id
	WHERE
		r.organization_id = @organization_id
)
SELECT
	id,
	organization_id,
	name,
	description,
	owner_profile_id,
	owner_full_name,
	treatment,
	note,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	category,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM
	rsks
WHERE %s
	AND %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risks: %w", err)
	}

	risks, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Risk])
	if err != nil {
		return fmt.Errorf("cannot collect risks: %w", err)
	}

	*r = risks

	return nil
}

func (r *Risk) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	description,
	category,
	owner_profile_id,
	NULL as owner_full_name,
	treatment,
	note,
	inherent_likelihood,
	inherent_impact,
	inherent_risk_score,
	residual_likelihood,
	residual_impact,
	residual_risk_score,
	snapshot_id,
	source_id,
	created_at,
	updated_at
FROM risks
WHERE %s
	AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": riskID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query risk: %w", err)
	}

	risk, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Risk])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect risk: %w", err)
	}

	*r = risk

	return nil
}

func (r *Risk) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO risks (id, tenant_id, organization_id, name, description, category, owner_profile_id, treatment, note, inherent_likelihood, inherent_impact, residual_likelihood, residual_impact, created_at, updated_at)
VALUES (@id, @tenant_id, @organization_id, @name, @description, @category, @owner_profile_id, @treatment, @note, @inherent_likelihood, @inherent_impact, @residual_likelihood, @residual_impact, @created_at, @updated_at)
`

	args := pgx.StrictNamedArgs{
		"id":                  r.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     r.OrganizationID,
		"name":                r.Name,
		"description":         r.Description,
		"category":            r.Category,
		"owner_profile_id":    r.OwnerID,
		"treatment":           r.Treatment,
		"note":                r.Note,
		"inherent_likelihood": r.InherentLikelihood,
		"inherent_impact":     r.InherentImpact,
		"residual_likelihood": r.ResidualLikelihood,
		"residual_impact":     r.ResidualImpact,
		"created_at":          r.CreatedAt,
		"updated_at":          r.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (r *Risk) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE risks
SET
	name = @name,
	description = @description,
	owner_profile_id = @owner_profile_id,
	treatment = @treatment,
	inherent_likelihood = @inherent_likelihood,
	inherent_impact = @inherent_impact,
	residual_likelihood = @residual_likelihood,
	residual_impact = @residual_impact,
	category = @category,
	note = @note,
	updated_at = @updated_at
WHERE %s
	AND id = @risk_id
	AND snapshot_id IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"risk_id":             r.ID,
		"name":                r.Name,
		"description":         r.Description,
		"category":            r.Category,
		"owner_profile_id":    r.OwnerID,
		"treatment":           r.Treatment,
		"note":                r.Note,
		"inherent_likelihood": r.InherentLikelihood,
		"inherent_impact":     r.InherentImpact,
		"residual_likelihood": r.ResidualLikelihood,
		"residual_impact":     r.ResidualImpact,
		"updated_at":          r.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (r *Risk) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	riskID gid.GID,
) error {
	q := `
DELETE FROM risks WHERE %s AND id = @id AND snapshot_id IS NULL
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": riskID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (r *Risks) CountByDocumentID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	documentID gid.GID,
	filter *RiskFilter,
) (int, error) {
	q := `
WITH rsks AS (
	SELECT
		r.id,
		r.tenant_id,
		r.search_vector,
		r.snapshot_id
	FROM
		risks r
	INNER JOIN
		risks_documents rd ON r.id = rd.risk_id
	WHERE
		rd.document_id = @document_id
)
SELECT
	COUNT(id)
FROM
	rsks
WHERE %s
	AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.NamedArgs{"document_id": documentID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (r Risks) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	if err := r.InsertRiskSnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
		return fmt.Errorf("cannot create risk snapshots: %w", err)
	}

	return nil
}

func (r Risks) InsertRiskSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO risks (
	tenant_id,
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	description,
	category,
	treatment,
	note,
	owner_profile_id,
	inherent_likelihood,
	inherent_impact,
	residual_likelihood,
	residual_impact,
	created_at,
	updated_at
)
SELECT
	@tenant_id,
	generate_gid(decode_base64_unpadded(@tenant_id), @risk_entity_type),
	@snapshot_id,
	r.id,
	r.organization_id,
	r.name,
	r.description,
	r.category,
	r.treatment,
	r.note,
	r.owner_profile_id,
	r.inherent_likelihood,
	r.inherent_impact,
	r.residual_likelihood,
	r.residual_impact,
	r.created_at,
	r.updated_at
FROM risks r
WHERE %s AND organization_id = @organization_id AND snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"snapshot_id":      snapshotID,
		"organization_id":  organizationID,
		"risk_entity_type": RiskEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert risk snapshots: %w", err)
	}

	return nil
}
