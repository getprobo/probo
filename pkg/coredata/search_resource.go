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
	"fmt"
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
	to_tsvector('simple', translate(COALESCE(@body, ''), '-+', '  ')),
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

	if _, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"resource_id": id}); err != nil {
		return fmt.Errorf("cannot delete search resource: %w", err)
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

func joinSearch(parts ...string) string {
	return strings.Join(parts, " ")
}

func derefSearch(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
