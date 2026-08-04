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
	"net/netip"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	IPLocation struct {
		CountryCode     CountryCode      `db:"country_code"`
		SubdivisionCode *SubdivisionCode `db:"subdivision_code"`
	}

	IPLocationBlock struct {
		AddressFamily   int16
		IPStart         netip.Addr
		IPEnd           netip.Addr
		CountryCode     CountryCode
		SubdivisionCode *SubdivisionCode
	}

	IPLocationBlockSource interface {
		Next() bool
		LocationBlock() (IPLocationBlock, error)
		Err() error
	}

	ipLocationCopySource struct {
		source IPLocationBlockSource
	}
)

func (s *ipLocationCopySource) Next() bool {
	return s.source.Next()
}

func (s *ipLocationCopySource) Values() ([]any, error) {
	block, err := s.source.LocationBlock()
	if err != nil {
		return nil, err
	}

	return []any{
		block.AddressFamily,
		block.IPStart,
		block.IPEnd,
		block.CountryCode.String(),
		block.SubdivisionCode,
	}, nil
}

func (s *ipLocationCopySource) Err() error {
	return s.source.Err()
}

func LookupLocationByIP(ctx context.Context, conn pg.Querier, ip string) (IPLocation, error) {
	q := `
WITH candidate AS (
	SELECT
		ip_end,
		country_code,
		subdivision_code
	FROM common_ip_location_blocks
	WHERE
		address_family = family(@ip::inet)
		AND ip_start <= @ip::inet
	ORDER BY ip_start DESC
	LIMIT 1
)
SELECT
	country_code,
	subdivision_code
FROM candidate
WHERE ip_end >= @ip::inet;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"ip": ip})
	if err != nil {
		return IPLocation{}, fmt.Errorf("cannot query IP location blocks: %w", err)
	}

	location, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IPLocation])
	if err != nil {
		if err == pgx.ErrNoRows {
			return IPLocation{}, nil
		}

		return IPLocation{}, fmt.Errorf("cannot collect IP location block row: %w", err)
	}

	return location, nil
}

func IsIPLocationBlocksPopulated(ctx context.Context, conn pg.Querier) (bool, error) {
	q := `SELECT EXISTS (SELECT 1 FROM common_ip_location_blocks);`

	rows, err := conn.Query(ctx, q)
	if err != nil {
		return false, fmt.Errorf("cannot query IP location block population: %w", err)
	}

	populated, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[bool])
	if err != nil {
		return false, fmt.Errorf("cannot collect IP location block population: %w", err)
	}

	return populated, nil
}

const ipLocationBlocksStagingTable = "common_ip_location_blocks_staging"

func CreateIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
DROP TABLE IF EXISTS common_ip_location_blocks_staging;

CREATE TABLE common_ip_location_blocks_staging
	(LIKE common_ip_location_blocks INCLUDING DEFAULTS INCLUDING CONSTRAINTS);
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot create IP location blocks staging table: %w", err)
	}

	return nil
}

func CopyIPLocationBlocksStaging(
	ctx context.Context,
	conn pg.Querier,
	source IPLocationBlockSource,
) (int64, error) {
	count, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{ipLocationBlocksStagingTable},
		[]string{
			"address_family",
			"ip_start",
			"ip_end",
			"country_code",
			"subdivision_code",
		},
		&ipLocationCopySource{source: source},
	)
	if err != nil {
		return 0, fmt.Errorf("cannot copy IP location blocks to staging: %w", err)
	}

	return count, nil
}

func FinalizeIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
CREATE UNIQUE INDEX idx_common_ip_location_blocks_staging_start
	ON common_ip_location_blocks_staging (address_family, ip_start);
ANALYZE common_ip_location_blocks_staging;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot finalize IP location blocks staging table: %w", err)
	}

	return nil
}

func SwapIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	q := `
DROP TABLE common_ip_location_blocks;
ALTER TABLE common_ip_location_blocks_staging
	RENAME TO common_ip_location_blocks;
ALTER INDEX idx_common_ip_location_blocks_staging_start
	RENAME TO idx_common_ip_location_blocks_start;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot swap IP location blocks staging table: %w", err)
	}

	return nil
}

func DropIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	if _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS common_ip_location_blocks_staging"); err != nil {
		return fmt.Errorf("cannot drop IP location blocks staging table: %w", err)
	}

	return nil
}
