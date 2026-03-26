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

package drivers

import (
	"context"
	"fmt"

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
		accounts, err := coredata.LoadMembershipAccountsByOrganizationID(
			ctx,
			conn,
			d.scope,
			d.organizationID,
		)
		if err != nil {
			return err
		}

		for _, account := range accounts {
			records = append(records, AccountRecord{
				Email:       account.Email,
				FullName:    account.FullName,
				Active:      account.State == "ACTIVE",
				MFAStatus:   coredata.MFAStatusUnknown,
				AuthMethod:  coredata.AccessEntryAuthMethodUnknown,
				AccountType: coredata.AccessEntryAccountTypeUser,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list probo membership accounts: %w", err)
	}

	return records, nil
}
