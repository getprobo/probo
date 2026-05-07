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
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type IPCountryBlock struct {
	CIDR        string      `db:"cidr"`
	CountryCode CountryCode `db:"country_code"`
}

func LookupCountryByIP(ctx context.Context, conn pg.Querier, ip string) (CountryCode, error) {
	q := `
SELECT country_code
FROM common_ip_country_blocks
WHERE cidr >>= @ip::inet
ORDER BY masklen(cidr) DESC
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"ip": ip}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return "", fmt.Errorf("cannot query ip country blocks: %w", err)
	}

	cc, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[CountryCode])
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("cannot collect ip country block row: %w", err)
	}

	return cc, nil
}

func IsIPCountryBlocksPopulated(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `SELECT EXISTS (SELECT 1 FROM common_ip_country_blocks);`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return false, fmt.Errorf("cannot check if ip country blocks is populated: %w", err)
	}

	populated, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect populated check: %w", err)
	}

	return populated, nil
}

const ipCountryBlocksStagingTable = "common_ip_country_blocks_staging"

func CreateIPCountryBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
DROP TABLE IF EXISTS common_ip_country_blocks_staging;
CREATE TABLE common_ip_country_blocks_staging (LIKE common_ip_country_blocks INCLUDING DEFAULTS);
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot create staging table: %w", err)
	}

	return nil
}

func CopyIPCountryBlocksStaging(ctx context.Context, conn pg.Querier, blocks []IPCountryBlock) error {
	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{ipCountryBlocksStagingTable},
		[]string{"cidr", "country_code"},
		pgx.CopyFromSlice(len(blocks), func(i int) ([]any, error) {
			return []any{blocks[i].CIDR, blocks[i].CountryCode.String()}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("cannot copy rows into staging table: %w", err)
	}

	return nil
}

func FinalizeIPCountryBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
CREATE INDEX idx_common_ip_country_blocks_staging_cidr
    ON common_ip_country_blocks_staging USING gist (cidr inet_ops);
ANALYZE common_ip_country_blocks_staging;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot finalize staging table: %w", err)
	}

	return nil
}

func SwapIPCountryBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
DROP TABLE common_ip_country_blocks;
ALTER TABLE common_ip_country_blocks_staging RENAME TO common_ip_country_blocks;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot swap staging table: %w", err)
	}

	return nil
}

func DropIPCountryBlocksStaging(ctx context.Context, conn pg.Querier) error {
	if _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS common_ip_country_blocks_staging"); err != nil {
		return fmt.Errorf("cannot drop staging table: %w", err)
	}

	return nil
}
