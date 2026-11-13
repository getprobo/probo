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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	// CMPDomain groups an organization's cookie banners that share an
	// eTLD+1. The row stays when its last banner is deleted so the
	// decisions recorded on its surfaces survive.
	CMPDomain struct {
		ID                         gid.GID    `db:"id"`
		OrganizationID             gid.GID    `db:"organization_id"`
		Domain                     string     `db:"domain"`
		SourcePromotionRequestedAt *time.Time `db:"source_promotion_requested_at"`
		SourcePromotionAttempts    int        `db:"source_promotion_attempts"`
		SourcePromotionAttemptedAt *time.Time `db:"source_promotion_attempted_at"`
		CreatedAt                  time.Time  `db:"created_at"`
		UpdatedAt                  time.Time  `db:"updated_at"`
	}

	CMPDomains []*CMPDomain
)

func (d *CMPDomain) CursorKey(field CMPDomainOrderField) page.CursorKey {
	switch field {
	case CMPDomainOrderFieldCreatedAt:
		return page.NewCursorKey(d.ID, d.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (d *CMPDomain) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cmp_domains WHERE id = ANY(@resource_ids::text[])`

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

func (d *CMPDomain) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	domainID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
FROM cmp_domains
WHERE
	%s
	AND id = @domain_id
LIMIT 1
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"domain_id": domainID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cmp domain: %w", err)
	}

	domain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CMPDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cmp domain: %w", err)
	}

	*d = domain

	return nil
}

func (d *CMPDomains) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[CMPDomainOrderField],
	filter *CMPDomainFilter,
) error {
	if filter == nil {
		filter = NewCMPDomainFilter(nil)
	}

	q := `
SELECT
	id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
FROM cmp_domains
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
	cursorArgs, err := cursor.SQLArguments()
	if err != nil {
		return fmt.Errorf("cannot validate cursor: %w", err)
	}
	maps.Copy(args, cursorArgs)

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cmp domains: %w", err)
	}

	domains, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CMPDomain])
	if err != nil {
		return fmt.Errorf("cannot collect cmp domains: %w", err)
	}

	*d = domains

	return nil
}

func (d *CMPDomain) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cmp_domains (
	id,
	tenant_id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@domain,
	@source_promotion_requested_at,
	@source_promotion_attempts,
	@source_promotion_attempted_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                            d.ID,
		"tenant_id":                     scope.GetTenantID(),
		"organization_id":               d.OrganizationID,
		"domain":                        d.Domain,
		"source_promotion_requested_at": d.SourcePromotionRequestedAt,
		"source_promotion_attempts":     d.SourcePromotionAttempts,
		"source_promotion_attempted_at": d.SourcePromotionAttemptedAt,
		"created_at":                    d.CreatedAt,
		"updated_at":                    d.UpdatedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "cmp_domains_organization_id_domain_key" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert cmp domain: %w", err)
	}

	return nil
}

// UpsertByOrganizationAndDomain inserts the domain or returns the row
// that already occupies (organization, domain). An existing row is not
// modified. The caller sets ID; this method does not allocate one.
func (d *CMPDomain) UpsertByOrganizationAndDomain(
	ctx context.Context,
	tx pg.Tx,
	tenantID gid.TenantID,
	organizationID gid.GID,
	domain string,
	now time.Time,
) error {
	if d.ID == gid.Nil {
		return fmt.Errorf("cmp domain id is required")
	}

	d.OrganizationID = organizationID
	d.Domain = domain
	d.CreatedAt = now
	d.UpdatedAt = now

	q := `
INSERT INTO cmp_domains (
	id,
	tenant_id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@domain,
	@source_promotion_requested_at,
	@source_promotion_attempts,
	@source_promotion_attempted_at,
	@created_at,
	@updated_at
)
ON CONFLICT ON CONSTRAINT cmp_domains_organization_id_domain_key
DO UPDATE SET updated_at = cmp_domains.updated_at
RETURNING
	id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
`

	args := pgx.StrictNamedArgs{
		"id":                            d.ID,
		"tenant_id":                     tenantID,
		"organization_id":               organizationID,
		"domain":                        domain,
		"source_promotion_requested_at": nil,
		"source_promotion_attempts":     0,
		"source_promotion_attempted_at": nil,
		"created_at":                    now,
		"updated_at":                    now,
	}

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert cmp domain: %w", err)
	}

	upserted, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CMPDomain])
	if err != nil {
		return fmt.Errorf("cannot collect upserted cmp domain: %w", err)
	}

	*d = upserted

	return nil
}

func (d *CMPDomain) Update(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_requested_at = @source_promotion_requested_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                            d.ID,
		"source_promotion_requested_at": d.SourcePromotionRequestedAt,
		"updated_at":                    d.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update cmp domain: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (d *CMPDomain) Delete(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
DELETE FROM cmp_domains
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": d.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete cmp domain: %w", err)
	}

	return nil
}

// SetSourcePromotionRequested asks the CMP domain worker to compare
// cookie sources across the domain's banners. A request that is already
// waiting keeps its original time and its attempt budget, so a busy
// domain is not pushed back. A new request, including one after the
// budget was exhausted, starts the budget over.
func (d *CMPDomain) SetSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_attempts = CASE
		WHEN source_promotion_requested_at IS NULL THEN 0
		ELSE source_promotion_attempts
	END,
	source_promotion_attempted_at = CASE
		WHEN source_promotion_requested_at IS NULL THEN NULL
		ELSE source_promotion_attempted_at
	END,
	source_promotion_requested_at = COALESCE(source_promotion_requested_at, NOW()),
	updated_at = NOW()
WHERE id = @id
`

	args := pgx.StrictNamedArgs{"id": d.ID}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot request cmp domain source promotion: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// ClearSourcePromotionRequested removes the domain from the queue and
// counts the attempt. The attempt is recorded before the work runs, so a
// crash between claim and completion still spends budget.
func (d *CMPDomain) ClearSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_requested_at = NULL,
	source_promotion_attempts = source_promotion_attempts + 1,
	source_promotion_attempted_at = NOW(),
	updated_at = NOW()
WHERE id = @id
RETURNING source_promotion_attempts, source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{"id": d.ID}

	var (
		attempts    int
		attemptedAt *time.Time
	)

	err := tx.QueryRow(ctx, q, args).Scan(&attempts, &attemptedAt)
	if err != nil {
		return fmt.Errorf("cannot clear cmp domain source promotion: %w", err)
	}

	d.SourcePromotionRequestedAt = nil
	d.SourcePromotionAttempts = attempts
	d.SourcePromotionAttemptedAt = attemptedAt

	return nil
}

// RequeueSourcePromotionRequested puts a failed attempt back on the queue
// without resetting the budget. The update matches the claim that failed,
// so an older task cannot queue the domain after a newer claim has taken
// it. A superseded claim is a no-op.
func (d *CMPDomain) RequeueSourcePromotionRequested(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_requested_at = COALESCE(source_promotion_requested_at, NOW()),
	updated_at = NOW()
WHERE
	id = @id
	AND source_promotion_attempts = @source_promotion_attempts
	AND source_promotion_attempted_at IS NOT DISTINCT FROM @source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{
		"id":                            d.ID,
		"source_promotion_attempts":     d.SourcePromotionAttempts,
		"source_promotion_attempted_at": d.SourcePromotionAttemptedAt,
	}

	_, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot requeue cmp domain source promotion: %w", err)
	}

	return nil
}

// ResetSourcePromotionAttempts clears the budget after a promotion finishes.
// It matches the claim that finished, so a late success cannot wipe the
// budget of a newer claim. A superseded claim is a no-op.
func (d *CMPDomain) ResetSourcePromotionAttempts(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_attempts = 0,
	source_promotion_attempted_at = NULL,
	updated_at = NOW()
WHERE
	id = @id
	AND source_promotion_attempts = @source_promotion_attempts
	AND source_promotion_attempted_at IS NOT DISTINCT FROM @source_promotion_attempted_at
`

	args := pgx.StrictNamedArgs{
		"id":                            d.ID,
		"source_promotion_attempts":     d.SourcePromotionAttempts,
		"source_promotion_attempted_at": d.SourcePromotionAttemptedAt,
	}

	result, err := tx.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset cmp domain source promotion attempts: %w", err)
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	d.SourcePromotionAttempts = 0
	d.SourcePromotionAttemptedAt = nil

	return nil
}

// ResetStaleSourcePromotions requeues a claim that was taken but never
// finished and has been idle longer than staleAfter. A finished promotion
// has a zero attempt count, and a claim at the cap stays clear.
func ResetStaleSourcePromotions(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
	maxAttempts int,
) error {
	q := `
UPDATE cmp_domains
SET
	source_promotion_requested_at = NOW(),
	updated_at = NOW()
WHERE
	source_promotion_requested_at IS NULL
	AND source_promotion_attempts > 0
	AND source_promotion_attempts < @max_attempts
	AND source_promotion_attempted_at < @stale_before
`

	args := pgx.StrictNamedArgs{
		"max_attempts": maxAttempts,
		"stale_before": time.Now().Add(-staleAfter),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot reset stale cmp domain source promotions: %w", err)
	}

	return nil
}

func (d *CMPDomain) LoadNextForSourcePromotionForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	domain,
	source_promotion_requested_at,
	source_promotion_attempts,
	source_promotion_attempted_at,
	created_at,
	updated_at
FROM cmp_domains
WHERE source_promotion_requested_at IS NOT NULL
ORDER BY source_promotion_requested_at ASC, id ASC
FOR UPDATE SKIP LOCKED
LIMIT 1
`

	rows, err := tx.Query(ctx, q)
	if err != nil {
		return fmt.Errorf("cannot query cmp domains for source promotion: %w", err)
	}

	domain, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CMPDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cmp domain for source promotion: %w", err)
	}

	*d = domain

	return nil
}
