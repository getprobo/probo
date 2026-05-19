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
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	CommonThirdPartyDomain struct {
		ID                 gid.GID   `db:"id"`
		CommonThirdPartyID gid.GID   `db:"common_third_party_id"`
		Domain             string    `db:"domain"`
		CreatedAt          time.Time `db:"created_at"`
		UpdatedAt          time.Time `db:"updated_at"`
	}

	CommonThirdPartyDomains []*CommonThirdPartyDomain
)

func (d *CommonThirdPartyDomain) LoadByDomain(
	ctx context.Context,
	conn pg.Querier,
	domain string,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    domain,
    created_at,
    updated_at
FROM
    common_third_party_domains
WHERE
    domain = @domain
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"domain": domain}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third party domain: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdPartyDomain])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect common third party domain: %w", err)
	}

	*d = row

	return nil
}

func (d CommonThirdPartyDomain) Insert(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
INSERT INTO common_third_party_domains (
    id,
    common_third_party_id,
    domain,
    created_at,
    updated_at
) VALUES (
    @id,
    @common_third_party_id,
    @domain,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                    d.ID,
		"common_third_party_id": d.CommonThirdPartyID,
		"domain":                d.Domain,
		"created_at":            d.CreatedAt,
		"updated_at":            d.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "common_third_party_domains_party_domain_key" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert common third party domain: %w", err)
	}

	return nil
}

func (d CommonThirdPartyDomain) Upsert(
	ctx context.Context,
	conn pg.Tx,
) (inserted bool, err error) {
	q := `
INSERT INTO common_third_party_domains (
    id,
    common_third_party_id,
    domain,
    created_at,
    updated_at
) VALUES (
    @id,
    @common_third_party_id,
    @domain,
    @created_at,
    @updated_at
)
ON CONFLICT (common_third_party_id, domain) DO UPDATE
SET
    updated_at = EXCLUDED.updated_at
RETURNING (xmax = 0) AS inserted
`

	args := pgx.StrictNamedArgs{
		"id":                    d.ID,
		"common_third_party_id": d.CommonThirdPartyID,
		"domain":                d.Domain,
		"created_at":            d.CreatedAt,
		"updated_at":            d.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert common third party domain: %w", err)
	}
	defer rows.Close()

	inserted, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	return inserted, nil
}

func (d CommonThirdPartyDomain) Delete(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `DELETE FROM common_third_party_domains WHERE id = @id`

	args := pgx.StrictNamedArgs{"id": d.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete common third party domain: %w", err)
	}

	return nil
}

func (ds *CommonThirdPartyDomains) LoadByCommonThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
	commonThirdPartyID gid.GID,
) error {
	q := `
SELECT
    id,
    common_third_party_id,
    domain,
    created_at,
    updated_at
FROM
    common_third_party_domains
WHERE
    common_third_party_id = @common_third_party_id
ORDER BY domain ASC;
`

	args := pgx.StrictNamedArgs{"common_third_party_id": commonThirdPartyID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common third party domains: %w", err)
	}

	domains, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonThirdPartyDomain])
	if err != nil {
		return fmt.Errorf("cannot collect common third party domains: %w", err)
	}

	*ds = domains

	return nil
}

func (ds *CommonThirdPartyDomains) DeleteByCommonThirdPartyID(
	ctx context.Context,
	conn pg.Tx,
	commonThirdPartyID gid.GID,
) error {
	q := `DELETE FROM common_third_party_domains WHERE common_third_party_id = @common_third_party_id`

	args := pgx.StrictNamedArgs{"common_third_party_id": commonThirdPartyID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete common third party domains: %w", err)
	}

	return nil
}
