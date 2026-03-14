// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package accesssource

import (
	"context"
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// ProboMembershipsDriver is a built-in identity source that queries
// iam_memberships + identities for the organization. No external
// connector is needed.
type ProboMembershipsDriver struct {
	pg             *pg.Client
	scope          coredata.Scoper
	organizationID gid.GID
}

func NewProboMembershipsDriver(
	pgClient *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
) *ProboMembershipsDriver {
	return &ProboMembershipsDriver{
		pg:             pgClient,
		scope:          scope,
		organizationID: organizationID,
	}
}

func (d *ProboMembershipsDriver) ListAccounts(ctx context.Context) ([]AccountRecord, error) {
	var records []AccountRecord

	err := d.pg.WithConn(ctx, func(conn pg.Conn) error {
		q := `
SELECT
    i.email_address,
    i.full_name,
    m.state
FROM
    iam_memberships m
JOIN
    identities i ON i.id = m.identity_id
WHERE
    m.%s
    AND m.organization_id = @organization_id
ORDER BY
    i.email_address ASC
`
		q = fmt.Sprintf(q, d.scope.SQLFragment())

		args := pgx.StrictNamedArgs{
			"organization_id": d.organizationID,
		}
		maps.Copy(args, d.scope.SQLArguments())

		rows, err := conn.Query(ctx, q, args)
		if err != nil {
			return fmt.Errorf("cannot query memberships: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var email, fullName, state string

			if err := rows.Scan(&email, &fullName, &state); err != nil {
				return fmt.Errorf("cannot scan membership row: %w", err)
			}

			records = append(records, AccountRecord{
				Email:    email,
				FullName: fullName,
				Active:   state == "ACTIVE",
				MFAStatus: coredata.MFAStatusUnknown,
				AuthMethod: coredata.AccessEntryAuthMethodUnknown,
			})
		}

		return rows.Err()
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list probo membership accounts: %w", err)
	}

	return records, nil
}
