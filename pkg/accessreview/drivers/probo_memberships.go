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

package drivers

import (
	"context"
	"fmt"
	"strings"

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

	err := d.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			accounts, err := coredata.LoadMembershipAccountsByOrganizationID(
				ctx,
				conn,
				d.scope,
				d.organizationID,
			)
			if err != nil {
				return fmt.Errorf("cannot load membership accounts: %w", err)
			}

			for _, account := range accounts {
				role := strings.TrimSpace(account.Role)

				roles := []string{}
				if role != "" {
					roles = []string{role}
				}

				isAdmin := role == string(coredata.MembershipRoleOwner) || role == string(coredata.MembershipRoleAdmin)
				createdAt := account.CreatedAt

				records = append(
					records,
					AccountRecord{
						Email:       account.Email,
						FullName:    account.FullName,
						Roles:       roles,
						Active:      new(account.State == string(coredata.ProfileStateActive)),
						IsAdmin:     isAdmin,
						ExternalID:  account.ID.String(),
						CreatedAt:   &createdAt,
						MFAStatus:   coredata.MFAStatusUnknown,
						AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
						AccountType: coredata.AccessReviewEntryAccountTypeUser,
					},
				)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list probo membership accounts: %w", err)
	}

	return records, nil
}
