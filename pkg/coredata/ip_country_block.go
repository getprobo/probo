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

func TruncateIPCountryBlocks(ctx context.Context, conn pg.Querier) error {
	_, err := conn.Exec(ctx, "TRUNCATE common_ip_country_blocks")
	if err != nil {
		return fmt.Errorf("cannot truncate common_ip_country_blocks: %w", err)
	}
	return nil
}

func CopyIPCountryBlocks(ctx context.Context, conn pg.Querier, blocks []IPCountryBlock) error {
	rows := make([][]any, len(blocks))
	for i, b := range blocks {
		rows[i] = []any{b.CIDR, b.CountryCode.String()}
	}

	_, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"common_ip_country_blocks"},
		[]string{"cidr", "country_code"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("cannot copy rows into common_ip_country_blocks: %w", err)
	}

	return nil
}
