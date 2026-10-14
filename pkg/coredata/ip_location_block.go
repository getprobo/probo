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
	IPLocationBlock struct {
		AddressFamily   int16            `db:"address_family"`
		IPStart         netip.Addr       `db:"ip_start"`
		IPEnd           netip.Addr       `db:"ip_end"`
		CountryCode     CountryCode      `db:"country_code"`
		SubdivisionCode *SubdivisionCode `db:"subdivision_code"`
	}

	IPLocationBlockSource interface {
		Next() bool
		LocationBlock() (IPLocationBlock, error)
		Err() error
	}

	ipLocationCopySource struct {
		source IPLocationBlockSource
	}

	ipLocationTableGrant struct {
		Grantee   string
		Privilege string
	}
)

const (
	ipLocationBlocksStagingTable = "common_ip_location_blocks_staging"
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

func LookupLocationByIP(ctx context.Context, conn pg.Querier, ip string) (IPLocationBlock, error) {
	q := `
SELECT
	address_family,
	ip_start,
	ip_end,
	country_code,
	subdivision_code
FROM common_ip_location_blocks
WHERE
	address_family = family(@ip::inet)
	AND ip_start <= @ip::inet
	AND ip_end >= @ip::inet
ORDER BY ip_start DESC, ip_end ASC
LIMIT 1;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"ip": ip})
	if err != nil {
		return IPLocationBlock{}, fmt.Errorf("cannot query IP location blocks: %w", err)
	}

	block, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[IPLocationBlock])
	if err != nil {
		if err == pgx.ErrNoRows {
			return IPLocationBlock{}, nil
		}

		return IPLocationBlock{}, fmt.Errorf("cannot collect IP location block row: %w", err)
	}

	return block, nil
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
CREATE UNIQUE INDEX idx_common_ip_location_blocks_staging_range
	ON common_ip_location_blocks_staging (address_family, ip_start, ip_end);
CREATE INDEX idx_common_ip_location_blocks_staging_lookup
	ON common_ip_location_blocks_staging (address_family, ip_start DESC, ip_end ASC);
ANALYZE common_ip_location_blocks_staging;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot finalize IP location blocks staging table: %w", err)
	}

	return nil
}

func SwapIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	owner, grants, err := ipLocationBlocksPrivileges(ctx, conn)
	if err != nil {
		return err
	}

	if err := grantIPLocationTablePrivileges(
		ctx,
		conn,
		ipLocationBlocksStagingTable,
		grants,
	); err != nil {
		return err
	}

	q := `
DROP TABLE common_ip_location_blocks;
ALTER TABLE common_ip_location_blocks_staging
	RENAME TO common_ip_location_blocks;
ALTER INDEX idx_common_ip_location_blocks_staging_range
	RENAME TO idx_common_ip_location_blocks_range;
ALTER INDEX idx_common_ip_location_blocks_staging_lookup
	RENAME TO idx_common_ip_location_blocks_lookup;
`

	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot swap IP location blocks staging table: %w", err)
	}

	if err := setIPLocationBlocksOwner(ctx, conn, owner); err != nil {
		return err
	}

	return nil
}

func ipLocationBlocksPrivileges(
	ctx context.Context,
	conn pg.Querier,
) (string, []ipLocationTableGrant, error) {
	ownerRows, err := conn.Query(
		ctx,
		`
SELECT pg_get_userbyid(c.relowner)
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = current_schema()
  AND c.relname = 'common_ip_location_blocks'
`,
	)
	if err != nil {
		return "", nil, fmt.Errorf("cannot read IP location blocks owner: %w", err)
	}

	owner, err := pgx.CollectExactlyOneRow(ownerRows, pgx.RowTo[string])
	if err != nil {
		return "", nil, fmt.Errorf("cannot collect IP location blocks owner: %w", err)
	}

	grantRows, err := conn.Query(
		ctx,
		`
SELECT
	CASE
		WHEN a.grantee = 0 THEN 'PUBLIC'
		ELSE r.rolname
	END,
	a.privilege_type
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
CROSS JOIN LATERAL aclexplode(COALESCE(c.relacl, '{}'::aclitem[])) AS a
LEFT JOIN pg_roles r ON r.oid = a.grantee
WHERE n.nspname = current_schema()
  AND c.relname = 'common_ip_location_blocks'
  AND (
	a.grantee = 0
	OR r.rolname <> CURRENT_USER
  )
`,
	)
	if err != nil {
		return "", nil, fmt.Errorf("cannot read IP location blocks privileges: %w", err)
	}

	grants, err := pgx.CollectRows(
		grantRows,
		pgx.RowToStructByPos[ipLocationTableGrant],
	)
	if err != nil {
		return "", nil, fmt.Errorf("cannot collect IP location blocks privileges: %w", err)
	}

	return owner, grants, nil
}

func grantIPLocationTablePrivileges(
	ctx context.Context,
	conn pg.Querier,
	table string,
	grants []ipLocationTableGrant,
) error {
	tableName := pgx.Identifier{table}.Sanitize()

	for _, grant := range grants {
		if !validIPLocationPrivilege(grant.Privilege) {
			return fmt.Errorf("cannot grant unknown privilege %q on %s", grant.Privilege, table)
		}

		q := fmt.Sprintf(
			"GRANT %s ON TABLE %s TO %s",
			grant.Privilege,
			tableName,
			ipLocationGrantTarget(grant.Grantee),
		)
		if _, err := conn.Exec(ctx, q); err != nil {
			return fmt.Errorf(
				"cannot grant %s on %s to %s: %w",
				grant.Privilege,
				table,
				grant.Grantee,
				err,
			)
		}
	}

	return nil
}

func setIPLocationBlocksOwner(ctx context.Context, conn pg.Querier, owner string) error {
	if owner == "" {
		return fmt.Errorf("cannot set IP location blocks owner: empty role")
	}

	q := fmt.Sprintf(
		"ALTER TABLE common_ip_location_blocks OWNER TO %s",
		pgx.Identifier{owner}.Sanitize(),
	)
	if _, err := conn.Exec(ctx, q); err != nil {
		return fmt.Errorf("cannot set IP location blocks owner to %s: %w", owner, err)
	}

	return nil
}

func ipLocationGrantTarget(grantee string) string {
	if grantee == "PUBLIC" {
		return "PUBLIC"
	}

	return pgx.Identifier{grantee}.Sanitize()
}

func validIPLocationPrivilege(privilege string) bool {
	switch privilege {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "TRUNCATE", "REFERENCES", "TRIGGER", "MAINTAIN":
		return true
	default:
		return false
	}
}

func DropIPLocationBlocksStaging(ctx context.Context, conn pg.Querier) error {
	if _, err := conn.Exec(ctx, "DROP TABLE IF EXISTS common_ip_location_blocks_staging"); err != nil {
		return fmt.Errorf("cannot drop IP location blocks staging table: %w", err)
	}

	return nil
}
