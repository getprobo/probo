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
	"bytes"
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

// UpsertSearchResource writes the palette index row for a live resource.
// tag_vector is left untouched when the row already exists.
func UpsertSearchResource(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
	organizationID gid.GID,
	title string,
	body string,
) error {
	q := `
INSERT INTO search_resources (
	resource_id,
	tenant_id,
	organization_id,
	entity_type,
	title,
	body_vector,
	tag_vector
) VALUES (
	@resource_id,
	@tenant_id,
	@organization_id,
	extract_entity_type(parse_gid(@resource_id)),
	COALESCE(NULLIF(BTRIM(@title), ''), @resource_id),
	to_tsvector('simple', translate(COALESCE(@body, '') || ' ' || @resource_id, '-+', '  ')),
	''::tsvector
)
ON CONFLICT (resource_id) DO UPDATE SET
	tenant_id = EXCLUDED.tenant_id,
	organization_id = EXCLUDED.organization_id,
	entity_type = EXCLUDED.entity_type,
	title = EXCLUDED.title,
	body_vector = EXCLUDED.body_vector
`

	args := pgx.StrictNamedArgs{
		"resource_id":     id,
		"tenant_id":       id.TenantID(),
		"organization_id": organizationID,
		"title":           title,
		"body":            body,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert search resource: %w", err)
	}

	return nil
}

// LockSearchResource serializes tag attachment with index cleanup for one resource.
func LockSearchResource(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `SELECT pg_advisory_xact_lock(hashtext('search-resource:' || @resource_id::text))`

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_id": id}); err != nil {
		return fmt.Errorf("cannot lock search resource: %w", err)
	}

	return nil
}

// SearchResourceExists reports whether the live index still has this resource.
func SearchResourceExists(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) (bool, error) {
	q := `SELECT 1 FROM search_resources WHERE resource_id = @resource_id`

	var one int

	err := conn.QueryRow(ctx, q, pgx.StrictNamedArgs{"resource_id": id}).Scan(&one)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("cannot check search resource: %w", err)
	}

	return true, nil
}

// DeleteSearchResource removes the index row and the resource's tag assignments.
func DeleteSearchResource(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `
WITH deleted_index AS (
	DELETE FROM search_resources
	WHERE resource_id = @resource_id
)
DELETE FROM tag_assignments
WHERE resource_id = @resource_id
`

	if err := LockSearchResource(ctx, conn, id); err != nil {
		return err
	}

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_id": id}); err != nil {
		return fmt.Errorf("cannot delete search resource: %w", err)
	}

	return nil
}

// DeleteSearchResources removes index rows and tag assignments for many resources.
func DeleteSearchResources(
	ctx context.Context,
	conn pg.Querier,
	ids []gid.GID,
) error {
	if len(ids) == 0 {
		return nil
	}

	// Concurrent bulk deletes must take advisory locks in the same order.
	sorted := slices.Clone(ids)
	slices.SortFunc(sorted, func(a, b gid.GID) int {
		return bytes.Compare(a[:], b[:])
	})

	for _, id := range sorted {
		if err := LockSearchResource(ctx, conn, id); err != nil {
			return err
		}
	}

	q := `
WITH deleted_index AS (
	DELETE FROM search_resources
	WHERE resource_id = ANY(@resource_ids::text[])
)
DELETE FROM tag_assignments
WHERE resource_id = ANY(@resource_ids::text[])
`

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_ids": ids}); err != nil {
		return fmt.Errorf("cannot delete search resources: %w", err)
	}

	return nil
}

// RefreshSearchTagVector recomputes tag_vector from current assignments.
// It does nothing when the resource is not indexed.
func RefreshSearchTagVector(
	ctx context.Context,
	conn pg.Querier,
	resourceID gid.GID,
) error {
	q := `
UPDATE search_resources
SET tag_vector = COALESCE(
	(
		SELECT to_tsvector(
			'simple',
			translate(string_agg(t.name, ' '), '-+', '  ')
		)
		FROM tag_assignments ta
		INNER JOIN tags t ON t.id = ta.tag_id
		WHERE ta.resource_id = @resource_id
	),
	''::tsvector
)
WHERE resource_id = @resource_id
`

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_id": resourceID}); err != nil {
		return fmt.Errorf("cannot refresh search tag vector: %w", err)
	}

	return nil
}

// RefreshSearchTagVectors recomputes tag_vector for each indexed resource.
func RefreshSearchTagVectors(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) error {
	if len(resourceIDs) == 0 {
		return nil
	}

	q := `
UPDATE search_resources AS sr
SET tag_vector = COALESCE(
	(
		SELECT to_tsvector(
			'simple',
			translate(string_agg(t.name, ' '), '-+', '  ')
		)
		FROM tag_assignments ta
		INNER JOIN tags t ON t.id = ta.tag_id
		WHERE ta.resource_id = sr.resource_id
	),
	''::tsvector
)
WHERE sr.resource_id = ANY(@resource_ids::text[])
`

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs}); err != nil {
		return fmt.Errorf("cannot refresh search tag vectors: %w", err)
	}

	return nil
}

// ReindexTreatmentPlansForRisk rewrites plan search rows after the risk is renamed.
func ReindexTreatmentPlansForRisk(
	ctx context.Context,
	conn pg.Querier,
	riskID gid.GID,
	riskName string,
) error {
	q := `
UPDATE search_resources AS sr
SET
	title = COALESCE(NULLIF(BTRIM(@risk_name), ''), sr.resource_id),
	body_vector = to_tsvector(
		'simple',
		translate(COALESCE(@risk_name, '') || ' ' || tp.treatment::text || ' ' || sr.resource_id, '-+', '  ')
	)
FROM treatment_plans AS tp
WHERE
	tp.risk_id = @risk_id
	AND sr.resource_id = tp.id
`

	args := pgx.StrictNamedArgs{
		"risk_id":   riskID,
		"risk_name": riskName,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot reindex treatment plans: %w", err)
	}

	return nil
}

// LockID locks one scoped row until the transaction ends.
func LockID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	table string,
	id gid.GID,
) error {
	switch table {
	case "risks", "third_parties", "evidences", "business_functions", "tags",
		"risk_analyses", "risk_analysis_diagrams", "risk_analysis_processes":
	default:
		return fmt.Errorf("cannot lock table %s", table)
	}

	q := fmt.Sprintf(
		`SELECT id FROM %s WHERE %s AND id = @id FOR UPDATE`,
		table,
		scope.SQLFragment(),
	)

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	var locked gid.GID
	if err := conn.QueryRow(ctx, q, args).Scan(&locked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot lock %s: %w", table, err)
	}

	return nil
}

func joinSearch(parts ...string) string {
	return strings.Join(parts, " ")
}

func derefSearch(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
