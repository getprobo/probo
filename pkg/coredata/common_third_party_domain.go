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

// LoadByDomain resolves a host to the catalog entry that owns it.
//
// A domain is unique per catalog entry, not globally: two entries for the same
// vendor each carry it until one is merged away. The ordering makes the choice
// deterministic in that window, so attribution cannot flip between two calls
// or after a vacuum. Collapsing the duplicate is `proboctl common-third-party
// merge`; this only keeps the answer stable until it runs.
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
ORDER BY
    created_at, id
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

func (d *CommonThirdPartyDomain) Upsert(
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
RETURNING
    id,
    common_third_party_id,
    domain,
    created_at,
    updated_at
`

	originalID := d.ID

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

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonThirdPartyDomain])
	if err != nil {
		return false, fmt.Errorf("cannot collect upsert result: %w", err)
	}

	*d = row

	return originalID == d.ID, nil
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

func (ds *CommonThirdPartyDomains) Load(
	ctx context.Context,
	conn pg.Querier,
	limit int,
	filter *CommonThirdPartyDomainFilter,
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
    %s
ORDER BY domain ASC
LIMIT @limit;
`

	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{"limit": limit}
	maps.Copy(args, filter.SQLArguments())

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

// LoadAllGroupedByCommonThirdPartyID returns every owned domain in the
// catalog, grouped by the catalog entry that owns it.
//
// Duplicate detection compares domain sets across the whole catalog, so it
// needs all of them at once; loading per entry would be one round trip per
// candidate. Only the domain strings are returned because that is all the
// comparison uses.
func (ds *CommonThirdPartyDomains) LoadAllGroupedByCommonThirdPartyID(
	ctx context.Context,
	conn pg.Querier,
) (map[gid.GID][]string, error) {
	q := `
SELECT
    common_third_party_id,
    domain
FROM
    common_third_party_domains
ORDER BY
    common_third_party_id,
    domain
`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("cannot query common third party domains: %w", err)
	}
	defer rows.Close()

	byParty := make(map[gid.GID][]string)

	for rows.Next() {
		var (
			partyID gid.GID
			domain  string
		)

		if err := rows.Scan(&partyID, &domain); err != nil {
			return nil, fmt.Errorf("cannot scan common third party domain: %w", err)
		}

		byParty[partyID] = append(byParty[partyID], domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate common third party domains: %w", err)
	}

	return byParty, nil
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
