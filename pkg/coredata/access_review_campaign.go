// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	AccessReviewCampaign struct {
		ID                gid.GID                    `db:"id"`
		OrganizationID    gid.GID                    `db:"organization_id"`
		Name              string                     `db:"name"`
		Description       string                     `db:"description"`
		Status            AccessReviewCampaignStatus `db:"status"`
		StartedAt         *time.Time                 `db:"started_at"`
		CompletedAt       *time.Time                 `db:"completed_at"`
		FrameworkControls []string                   `db:"framework_controls"`
		CreatedAt         time.Time                  `db:"created_at"`
		UpdatedAt         time.Time                  `db:"updated_at"`
	}

	AccessReviewCampaigns []*AccessReviewCampaign
)

func (c AccessReviewCampaign) CursorKey(orderBy AccessReviewCampaignOrderField) page.CursorKey {
	switch orderBy {
	case AccessReviewCampaignOrderFieldCreatedAt:
		return page.NewCursorKey(c.ID, c.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (c *AccessReviewCampaign) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM access_review_campaigns WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, c.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query access review campaign authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (c *AccessReviewCampaign) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    status,
    started_at,
    completed_at,
    framework_controls,
    created_at,
    updated_at
FROM
    access_review_campaigns
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_campaigns: %w", err)
	}

	campaign, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaign])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect access review campaign: %w", err)
	}

	*c = campaign

	return nil
}

func (c *AccessReviewCampaign) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO
    access_review_campaigns (
        id,
        tenant_id,
        organization_id,
        name,
        description,
        status,
        started_at,
        completed_at,
        framework_controls,
        created_at,
        updated_at
    )
VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @name,
    @description,
    @status,
    @started_at,
    @completed_at,
    @framework_controls,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"tenant_id":          scope.GetTenantID(),
		"organization_id":    c.OrganizationID,
		"name":               c.Name,
		"description":        c.Description,
		"status":             c.Status,
		"started_at":         c.StartedAt,
		"completed_at":       c.CompletedAt,
		"framework_controls": c.FrameworkControls,
		"created_at":         c.CreatedAt,
		"updated_at":         c.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert access_review_campaign: %w", err)
	}

	return nil
}

func (c *AccessReviewCampaign) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_review_campaigns
SET
    name = @name,
    description = @description,
    status = @status,
    started_at = @started_at,
    completed_at = @completed_at,
    framework_controls = @framework_controls,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                 c.ID,
		"name":               c.Name,
		"description":        c.Description,
		"status":             c.Status,
		"started_at":         c.StartedAt,
		"completed_at":       c.CompletedAt,
		"framework_controls": c.FrameworkControls,
		"updated_at":         c.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access_review_campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (c *AccessReviewCampaign) Delete(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM access_review_campaigns
WHERE %s AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access_review_campaign: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (campaigns *AccessReviewCampaigns) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[AccessReviewCampaignOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    status,
    started_at,
    completed_at,
    framework_controls,
    created_at,
    updated_at
FROM
    access_review_campaigns
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_campaigns: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaign])
	if err != nil {
		return fmt.Errorf("cannot collect access_review_campaigns: %w", err)
	}

	*campaigns = result

	return nil
}

func (campaigns *AccessReviewCampaigns) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM access_review_campaigns
WHERE
    %s
    AND organization_id = @organization_id;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count access_review_campaigns: %w", err)
	}

	return count, nil
}

func (c *AccessReviewCampaign) LoadLastCompletedByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    description,
    status,
    started_at,
    completed_at,
    framework_controls,
    created_at,
    updated_at
FROM
    access_review_campaigns
WHERE
    %s
    AND organization_id = @organization_id
    AND status = 'COMPLETED'
ORDER BY completed_at DESC
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query access_review_campaigns: %w", err)
	}

	campaign, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaign])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect access review campaign: %w", err)
	}

	*c = campaign

	return nil
}
