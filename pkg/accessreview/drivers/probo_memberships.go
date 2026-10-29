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

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// ProboMembershipsDriver is a built-in identity source that queries
// the organization's memberships and their profiles. No external
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
	var (
		memberships           []*coredata.Membership
		membershipsByIdentity = map[gid.GID]*coredata.Membership{}
		records               []AccountRecord
	)

	err := d.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			err := page.WalkAll(
				ctx,
				page.OrderBy[coredata.MembershipOrderField]{
					Field:     coredata.MembershipOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.MembershipOrderField]) (coredata.Memberships, error) {
					var batch coredata.Memberships
					if err := batch.LoadByOrganizationID(ctx, conn, d.scope, d.organizationID, cursor); err != nil {
						return nil, fmt.Errorf("cannot load memberships: %w", err)
					}

					return batch, nil
				},
				func(batch coredata.Memberships) error {
					memberships = append(memberships, batch...)
					for _, membership := range batch {
						membershipsByIdentity[membership.IdentityID] = membership
					}

					return nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot walk memberships: %w", err)
			}

			filter := coredata.NewMembershipProfileFilter(nil).WithMembership()

			err = page.WalkAll(
				ctx,
				page.OrderBy[coredata.MembershipProfileOrderField]{
					Field:     coredata.MembershipProfileOrderFieldEmailAddress,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.MembershipProfileOrderField]) (coredata.MembershipProfiles, error) {
					var batch coredata.MembershipProfiles
					if err := batch.LoadByOrganizationID(ctx, conn, d.scope, d.organizationID, cursor, filter); err != nil {
						return nil, fmt.Errorf("cannot load membership profiles: %w", err)
					}

					return batch, nil
				},
				func(profiles coredata.MembershipProfiles) error {
					for _, profile := range profiles {
						membership, ok := membershipsByIdentity[profile.IdentityID]
						if !ok {
							continue
						}

						delete(membershipsByIdentity, profile.IdentityID)

						records = append(records, proboMembershipRecord(membership, profile))
					}

					return nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot walk membership profiles: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list probo membership accounts: %w", err)
	}

	// A membership without a profile is still a grant on record, so it
	// is reviewed as inactive (the authorizer denies it) rather than hidden.
	for _, membership := range memberships {
		if _, ok := membershipsByIdentity[membership.IdentityID]; ok {
			records = append(records, proboMembershipRecord(membership, nil))
		}
	}

	return records, nil
}

func proboMembershipRecord(membership *coredata.Membership, profile *coredata.MembershipProfile) AccountRecord {
	isAdmin := membership.Role == coredata.MembershipRoleOwner || membership.Role == coredata.MembershipRoleAdmin
	createdAt := membership.CreatedAt

	record := AccountRecord{
		Roles:       []string{string(membership.Role)},
		IsAdmin:     new(isAdmin),
		ExternalID:  membership.ID.String(),
		CreatedAt:   &createdAt,
		MFAStatus:   coredata.MFAStatusUnknown,
		AuthMethod:  coredata.AccessReviewEntryAuthMethodUnknown,
		AccountType: coredata.AccessReviewEntryAccountTypeUser,
		Active:      new(false),
	}

	if profile != nil {
		record.Email = profile.EmailAddress.String()
		record.FullName = profile.FullName
		record.Active = new(profile.State == coredata.ProfileStateActive)
	}

	return record
}
