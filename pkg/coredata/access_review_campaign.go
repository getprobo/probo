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
	AccessReviewCampaign struct {
		ID             gid.GID                    `db:"id"`
		OrganizationID gid.GID                    `db:"organization_id"`
		Name           string                     `db:"name"`
		Description    string                     `db:"description"`
		Status         AccessReviewCampaignStatus `db:"status"`
		StartedAt      *time.Time                 `db:"started_at"`
		CompletedAt    *time.Time                 `db:"completed_at"`
		CreatedAt      time.Time                  `db:"created_at"`
		UpdatedAt      time.Time                  `db:"updated_at"`
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

func (c *AccessReviewCampaign) LockForUpdate(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
SELECT id
FROM access_review_campaigns
WHERE %s
  AND id = @id
FOR UPDATE
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	var id gid.GID
	if err := conn.QueryRow(ctx, q, args).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot lock campaign: %w", err)
	}

	return nil
}

func (c *AccessReviewCampaign) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM access_review_campaigns WHERE id = ANY(@resource_ids::text[])`

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
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": c.OrganizationID,
		"name":            c.Name,
		"description":     c.Description,
		"status":          c.Status,
		"started_at":      c.StartedAt,
		"completed_at":    c.CompletedAt,
		"created_at":      c.CreatedAt,
		"updated_at":      c.UpdatedAt,
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
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":           c.ID,
		"name":         c.Name,
		"description":  c.Description,
		"status":       c.Status,
		"started_at":   c.StartedAt,
		"completed_at": c.CompletedAt,
		"updated_at":   c.UpdatedAt,
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

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete access_review_campaign: %w", err)
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
