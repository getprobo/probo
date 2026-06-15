// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/iam/policy"
)

type (
	// AccessReviewCampaignSource is the per-campaign snapshot of an access
	// source. It captures the source identity (name, connector) at
	// the time the source was scoped into the campaign so that the review's
	// data survives even if the live access source is later deleted. Access
	// entries and fetch attempts reference this snapshot, not the live source.
	AccessReviewCampaignSource struct {
		ID                     gid.GID      `db:"id"`
		TenantID               gid.TenantID `db:"tenant_id"`
		AccessReviewCampaignID gid.GID      `db:"access_review_campaign_id"`
		AccessReviewSourceID   *gid.GID     `db:"access_review_source_id"`
		Name                   string       `db:"name"`
		ConnectorID            *gid.GID     `db:"connector_id"`
		CreatedAt              time.Time    `db:"created_at"`
		UpdatedAt              time.Time    `db:"updated_at"`
	}

	AccessReviewCampaignSources []*AccessReviewCampaignSource
)

func (s *AccessReviewCampaignSource) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
	cs.id,
	c.organization_id
FROM
	access_review_campaign_sources cs
JOIN
	access_review_campaigns c ON c.id = cs.access_review_campaign_id
WHERE
	cs.id = ANY(@resource_ids::text[])
`

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

// Upsert inserts the snapshot or refreshes its denormalized identity from the
// live source. The generated ID is preserved across upserts because it is not
// part of the conflict target, so entries that already reference the snapshot
// keep pointing at the same row.
func (s *AccessReviewCampaignSource) Upsert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO access_review_campaign_sources (
	id,
	tenant_id,
	access_review_campaign_id,
	access_review_source_id,
	name,
	connector_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@access_review_campaign_id,
	@access_review_source_id,
	@name,
	@connector_id,
	@created_at,
	@updated_at
)
ON CONFLICT (access_review_campaign_id, access_review_source_id) DO UPDATE SET
	name         = EXCLUDED.name,
	connector_id = EXCLUDED.connector_id,
	updated_at   = EXCLUDED.updated_at
RETURNING id
`
	args := pgx.StrictNamedArgs{
		"id":                        s.ID,
		"tenant_id":                 scope.GetTenantID(),
		"access_review_campaign_id": s.AccessReviewCampaignID,
		"access_review_source_id":   s.AccessReviewSourceID,
		"name":                      s.Name,
		"connector_id":              s.ConnectorID,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}

	if err := conn.QueryRow(ctx, q, args).Scan(&s.ID); err != nil {
		return fmt.Errorf("cannot upsert campaign source: %w", err)
	}

	return nil
}

func (s *AccessReviewCampaignSource) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_id,
	access_review_source_id,
	name,
	connector_id,
	created_at,
	updated_at
FROM access_review_campaign_sources
WHERE
	%s
	AND id = @id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query campaign source: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaignSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect campaign source: %w", err)
	}

	*s = result

	return nil
}

func (s *AccessReviewCampaignSource) DeleteByCampaignIDAndAccessReviewSourceID(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	campaignID gid.GID,
	accessSourceID gid.GID,
) error {
	q := `
DELETE FROM access_review_campaign_sources
WHERE
	%s
	AND access_review_campaign_id = @access_review_campaign_id
	AND access_review_source_id = @access_review_source_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": campaignID,
		"access_review_source_id":   accessSourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete campaign source: %w", err)
	}

	return nil
}

func (sources *AccessReviewCampaignSources) LoadByCampaignID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_id,
	access_review_source_id,
	name,
	connector_id,
	created_at,
	updated_at
FROM access_review_campaign_sources
WHERE
	%s
	AND access_review_campaign_id = @access_review_campaign_id
ORDER BY name ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query campaign sources: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSource])
	if err != nil {
		return fmt.Errorf("cannot collect campaign sources: %w", err)
	}

	*sources = result

	return nil
}
