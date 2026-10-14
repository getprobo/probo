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

package coredata_test

import (
	"context"
	"fmt"
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
)

type oneLocationBlockSource struct {
	done bool
}

func (s *oneLocationBlockSource) Next() bool {
	if s.done {
		return false
	}

	s.done = true

	return true
}

func (s *oneLocationBlockSource) LocationBlock() (coredata.IPLocationBlock, error) {
	return coredata.IPLocationBlock{
		AddressFamily: 4,
		IPStart:       netip.MustParseAddr("1.0.0.0"),
		IPEnd:         netip.MustParseAddr("1.0.0.255"),
		CountryCode:   coredata.CountryCodeUS,
	}, nil
}

func (s *oneLocationBlockSource) Err() error {
	return nil
}

func TestSwapIPLocationBlocksStagingPreservesOwnerAndGrants(t *testing.T) {
	client := test.PGClient(t)
	ctx := t.Context()
	suffix := time.Now().UnixNano()
	schema := fmt.Sprintf("ip_location_swap_%d", suffix)
	appRole := fmt.Sprintf("ip_loc_app_%d", suffix)
	readonly := fmt.Sprintf("ip_loc_ro_%d", suffix)
	readwrite := fmt.Sprintf("ip_loc_rw_%d", suffix)
	admin := fmt.Sprintf("ip_loc_admin_%d", suffix)
	service := fmt.Sprintf("ip_loc_svc_%d", suffix)

	t.Cleanup(func() {
		_ = client.WithConn(
			context.Background(),
			func(ctx context.Context, conn pg.Querier) error {
				_, _ = conn.Exec(ctx, "DROP SCHEMA IF EXISTS "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
				for _, role := range []string{appRole, readonly, readwrite, admin, service} {
					_, _ = conn.Exec(ctx, "DROP ROLE IF EXISTS "+pgx.Identifier{role}.Sanitize())
				}

				return nil
			},
		)
	})

	err := client.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			for _, role := range []string{appRole, readonly, readwrite, admin, service} {
				if _, err := tx.Exec(ctx, "CREATE ROLE "+pgx.Identifier{role}.Sanitize()+" NOLOGIN"); err != nil {
					return err
				}
			}

			if _, err := tx.Exec(ctx, "GRANT "+pgx.Identifier{appRole}.Sanitize()+" TO CURRENT_USER"); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, "SELECT set_config('search_path', $1, true)", schema); err != nil {
				return err
			}

			if _, err := tx.Exec(
				ctx,
				`
CREATE TABLE common_ip_location_blocks (
	address_family   SMALLINT    NOT NULL,
	ip_start         INET        NOT NULL,
	ip_end           INET        NOT NULL,
	country_code     CHAR(2)     NOT NULL,
	subdivision_code VARCHAR(15)
);
INSERT INTO common_ip_location_blocks
	(address_family, ip_start, ip_end, country_code)
VALUES
	(4, '9.9.9.0', '9.9.9.255', 'US');
`,
			); err != nil {
				return err
			}

			quotedTable := pgx.Identifier{"common_ip_location_blocks"}.Sanitize()
			if _, err := tx.Exec(
				ctx,
				"ALTER TABLE "+quotedTable+" OWNER TO "+pgx.Identifier{appRole}.Sanitize(),
			); err != nil {
				return err
			}

			grants := []string{
				"GRANT SELECT ON " + quotedTable + " TO " + pgx.Identifier{readonly}.Sanitize() + " WITH GRANT OPTION",
				"GRANT INSERT, UPDATE, DELETE ON " + quotedTable + " TO " + pgx.Identifier{readwrite}.Sanitize(),
				"GRANT TRUNCATE, REFERENCES, TRIGGER ON " + quotedTable + " TO " + pgx.Identifier{admin}.Sanitize(),
				"GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON " + quotedTable + " TO " + pgx.Identifier{service}.Sanitize(),
				"GRANT SELECT ON " + quotedTable + " TO PUBLIC",
				"GRANT SELECT ON " + quotedTable + " TO CURRENT_USER",
			}
			for _, grant := range grants {
				if _, err := tx.Exec(ctx, grant); err != nil {
					return err
				}
			}

			if err := coredata.CreateIPLocationBlocksStaging(ctx, tx); err != nil {
				return err
			}

			if _, err := coredata.CopyIPLocationBlocksStaging(ctx, tx, &oneLocationBlockSource{}); err != nil {
				return err
			}

			if err := coredata.FinalizeIPLocationBlocksStaging(ctx, tx); err != nil {
				return err
			}

			if err := coredata.SwapIPLocationBlocksStaging(ctx, tx); err != nil {
				return err
			}

			var owner string
			if err := tx.QueryRow(
				ctx,
				`
SELECT pg_get_userbyid(c.relowner)
FROM pg_class c
JOIN pg_namespace n ON n.oid = c.relnamespace
WHERE n.nspname = current_schema()
  AND c.relname = 'common_ip_location_blocks'
`,
			).Scan(&owner); err != nil {
				return err
			}

			assert.Equal(t, appRole, owner)

			privileges := []struct {
				role      string
				privilege string
				want      bool
			}{
				{readonly, "SELECT", true},
				{readonly, "INSERT", false},
				{readwrite, "INSERT", true},
				{readwrite, "UPDATE", true},
				{readwrite, "DELETE", true},
				{readwrite, "SELECT", true},
				{admin, "TRUNCATE", true},
				{admin, "REFERENCES", true},
				{admin, "TRIGGER", true},
				{service, "SELECT", true},
				{service, "INSERT", true},
				{service, "UPDATE", true},
				{service, "DELETE", true},
				{service, "TRUNCATE", true},
				{service, "REFERENCES", true},
				{service, "TRIGGER", true},
			}
			for _, check := range privileges {
				var allowed bool
				if err := tx.QueryRow(
					ctx,
					`SELECT has_table_privilege($1, 'common_ip_location_blocks', $2)`,
					check.role,
					check.privilege,
				).Scan(&allowed); err != nil {
					return err
				}

				assert.Equal(t, check.want, allowed, "%s %s", check.role, check.privilege)
			}

			var currentUserSelect bool
			if err := tx.QueryRow(
				ctx,
				`SELECT has_table_privilege(CURRENT_USER, 'common_ip_location_blocks', 'SELECT')`,
			).Scan(&currentUserSelect); err != nil {
				return err
			}

			assert.True(t, currentUserSelect)

			var publicSelect, currentUserGrant, readonlyGrantable bool
			if err := tx.QueryRow(
				ctx,
				`
SELECT
	EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		CROSS JOIN LATERAL aclexplode(COALESCE(c.relacl, '{}'::aclitem[])) AS a
		WHERE n.nspname = current_schema()
		  AND c.relname = 'common_ip_location_blocks'
		  AND a.grantee = 0
		  AND a.privilege_type = 'SELECT'
	),
	EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		CROSS JOIN LATERAL aclexplode(COALESCE(c.relacl, '{}'::aclitem[])) AS a
		JOIN pg_roles r ON r.oid = a.grantee
		WHERE n.nspname = current_schema()
		  AND c.relname = 'common_ip_location_blocks'
		  AND r.rolname = CURRENT_USER
		  AND a.privilege_type = 'SELECT'
	),
	EXISTS (
		SELECT 1
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		CROSS JOIN LATERAL aclexplode(COALESCE(c.relacl, '{}'::aclitem[])) AS a
		JOIN pg_roles r ON r.oid = a.grantee
		WHERE n.nspname = current_schema()
		  AND c.relname = 'common_ip_location_blocks'
		  AND r.rolname = $1
		  AND a.privilege_type = 'SELECT'
		  AND a.is_grantable
	)
`,
				readonly,
			).Scan(&publicSelect, &currentUserGrant, &readonlyGrantable); err != nil {
				return err
			}

			assert.True(t, publicSelect)
			assert.True(t, currentUserGrant)
			assert.True(t, readonlyGrantable)

			return nil
		},
	)
	require.NoError(t, err)
}
