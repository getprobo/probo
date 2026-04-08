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
	"encoding/json"
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
	ConsentRecord struct {
		ID             gid.GID         `db:"id"`
		CookieBannerID gid.GID         `db:"cookie_banner_id"`
		VisitorID      string          `db:"visitor_id"`
		IPAddress      *string         `db:"ip_address"`
		UserAgent      *string         `db:"user_agent"`
		ConsentData    json.RawMessage `db:"consent_data"`
		Action         ConsentAction   `db:"action"`
		BannerVersion  int             `db:"banner_version"`
		CreatedAt      time.Time       `db:"created_at"`
	}

	ConsentRecords []*ConsentRecord
)

func (r *ConsentRecord) CursorKey(field ConsentRecordOrderField) page.CursorKey {
	switch field {
	case ConsentRecordOrderFieldCreatedAt:
		return page.NewCursorKey(r.ID, r.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (r *ConsentRecord) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `
SELECT cb.organization_id
FROM consent_records cr
JOIN cookie_banners cb ON cr.cookie_banner_id = cb.id
WHERE cr.id = $1
LIMIT 1;
`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, r.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query consent record authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (r *ConsentRecords) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[ConsentRecordOrderField],
	filter *ConsentRecordFilter,
) error {
	q := `
SELECT
	id,
	cookie_banner_id,
	visitor_id,
	ip_address,
	user_agent,
	consent_data,
	action,
	banner_version,
	created_at
FROM
	consent_records
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query consent records: %w", err)
	}

	records, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ConsentRecord])
	if err != nil {
		return fmt.Errorf("cannot collect consent records: %w", err)
	}

	*r = records

	return nil
}

func (r *ConsentRecords) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	filter *ConsentRecordFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	consent_records
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

type ConsentAnalytics struct {
	TotalRecords        int
	AcceptAllCount      int
	RejectAllCount      int
	CustomizeCount      int
	AcceptCategoryCount int
	GPCCount            int
}

func (a *ConsentAnalytics) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	COUNT(*) AS total_records,
	COUNT(*) FILTER (WHERE action = 'ACCEPT_ALL') AS accept_all_count,
	COUNT(*) FILTER (WHERE action = 'REJECT_ALL') AS reject_all_count,
	COUNT(*) FILTER (WHERE action = 'CUSTOMIZE') AS customize_count,
	COUNT(*) FILTER (WHERE action = 'ACCEPT_CATEGORY') AS accept_category_count,
	COUNT(*) FILTER (WHERE action = 'GPC') AS gpc_count
FROM
	consent_records
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	if err := row.Scan(
		&a.TotalRecords,
		&a.AcceptAllCount,
		&a.RejectAllCount,
		&a.CustomizeCount,
		&a.AcceptCategoryCount,
		&a.GPCCount,
	); err != nil {
		return fmt.Errorf("cannot scan consent analytics: %w", err)
	}

	return nil
}

func (r *ConsentRecord) Insert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO consent_records (
	id,
	tenant_id,
	cookie_banner_id,
	visitor_id,
	ip_address,
	user_agent,
	consent_data,
	action,
	banner_version,
	created_at
) VALUES (
	@id,
	@tenant_id,
	@cookie_banner_id,
	@visitor_id,
	@ip_address,
	@user_agent,
	@consent_data,
	@action,
	@banner_version,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               r.ID,
		"tenant_id":        scope.GetTenantID(),
		"cookie_banner_id": r.CookieBannerID,
		"visitor_id":       r.VisitorID,
		"ip_address":       r.IPAddress,
		"user_agent":       r.UserAgent,
		"consent_data":     r.ConsentData,
		"action":           r.Action,
		"banner_version":   r.BannerVersion,
		"created_at":       r.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert consent record: %w", err)
	}

	return nil
}
