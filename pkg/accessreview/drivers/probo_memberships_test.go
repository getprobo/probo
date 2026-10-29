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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

// proboMember describes one seeded identity. A nil State seeds no
// profile and an empty Role seeds no membership.
type proboMember struct {
	Name  string
	State *coredata.ProfileState
	Role  coredata.MembershipRole
}

func insertProboMembersOrganization(t *testing.T, ctx context.Context, client *pg.Client, tenantID gid.TenantID) (coredata.Scoper, gid.GID) {
	t.Helper()

	now := time.Now().UTC()
	organization := &coredata.Organization{
		ID:        gid.New(tenantID, coredata.OrganizationEntityType),
		TenantID:  tenantID,
		Name:      "Probo Memberships Driver Test Org",
		CreatedAt: now,
		UpdatedAt: now,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return organization.Insert(ctx, tx)
	}))

	return coredata.NewScope(tenantID, gid.Nil), organization.ID
}

// insertProboMembers seeds the members of one organization and returns
// their membership IDs by name. identities reuses an identity across
// organizations when the name is already in it.
func insertProboMembers(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	scope coredata.Scoper,
	organizationID gid.GID,
	identities map[string]gid.GID,
	members []proboMember,
) map[string]gid.GID {
	t.Helper()

	now := time.Now().UTC()
	membershipIDs := map[string]gid.GID{}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		for i, member := range members {
			identityID, ok := identities[member.Name]
			if !ok {
				email, err := mail.ParseAddr(fmt.Sprintf("%s-%s@example.com", member.Name, organizationID))
				if err != nil {
					return err
				}

				identity := &coredata.Identity{
					ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
					EmailAddress:         email,
					FullName:             "identity " + member.Name,
					EmailAddressVerified: true,
					CreatedAt:            now,
					UpdatedAt:            now,
				}
				if err := identity.Insert(ctx, tx); err != nil {
					return err
				}

				identityID = identity.ID
				identities[member.Name] = identityID
			}

			if member.State != nil {
				profile := &coredata.MembershipProfile{
					ID:             gid.New(scope.GetTenantID(), coredata.MembershipProfileEntityType),
					IdentityID:     identityID,
					OrganizationID: organizationID,
					Source:         coredata.ProfileSourceManual,
					State:          *member.State,
					FullName:       "profile " + member.Name,
					CreatedAt:      now,
					UpdatedAt:      now,
				}
				if err := profile.Insert(ctx, tx); err != nil {
					return err
				}
			}

			if member.Role != "" {
				membership := &coredata.Membership{
					ID:             gid.New(scope.GetTenantID(), coredata.MembershipEntityType),
					IdentityID:     identityID,
					OrganizationID: organizationID,
					Role:           member.Role,
					CreatedAt:      now.Add(-time.Hour + time.Duration(i)*time.Millisecond),
					UpdatedAt:      now,
				}
				if err := membership.Insert(ctx, tx, scope); err != nil {
					return err
				}

				membershipIDs[member.Name] = membership.ID
			}
		}

		return nil
	}))

	return membershipIDs
}

// TestProboMembershipsDriverListAccounts pins the built-in Probo members
// source: name and state come from the organization's membership
// profile, role and creation date from the membership. Pending and
// deactivated members are kept with Active=false, and so is a
// membership without a profile.
func TestProboMembershipsDriverListAccounts(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	tenantID := gid.NewTenantID()
	scope, organizationID := insertProboMembersOrganization(t, ctx, client, tenantID)
	// Same tenant, so only the organization filter keeps it out.
	otherScope, otherOrganizationID := insertProboMembersOrganization(t, ctx, client, tenantID)
	identities := map[string]gid.GID{}

	membershipIDs := insertProboMembers(
		t,
		ctx,
		client,
		scope,
		organizationID,
		identities,
		[]proboMember{
			{Name: "alice", State: new(coredata.ProfileStateActive), Role: coredata.MembershipRoleOwner},
			{Name: "bob", State: new(coredata.ProfileStateDeactivated), Role: coredata.MembershipRoleViewer},
			{Name: "carol", State: new(coredata.ProfileStatePending), Role: coredata.MembershipRoleViewer},
			{Name: "dave", State: new(coredata.ProfileStateActive)},
			{Name: "erin", Role: coredata.MembershipRoleAdmin},
		},
	)
	insertProboMembers(
		t,
		ctx,
		client,
		otherScope,
		otherOrganizationID,
		identities,
		[]proboMember{
			{Name: "alice", State: new(coredata.ProfileStateDeactivated), Role: coredata.MembershipRoleViewer},
		},
	)

	records, err := NewProboMembershipsDriver(client, scope, organizationID).ListAccounts(ctx)
	require.NoError(t, err)
	require.Len(t, records, 4, "dave has no membership and must be left out")

	alice := records[0]
	require.Equal(t, membershipIDs["alice"].String(), alice.ExternalID)
	require.Equal(t, fmt.Sprintf("alice-%s@example.com", organizationID), alice.Email)
	require.Equal(t, "profile alice", alice.FullName)
	require.Equal(t, []string{string(coredata.MembershipRoleOwner)}, alice.Roles)
	require.Equal(t, new(true), alice.IsAdmin)
	require.Equal(t, new(true), alice.Active)
	require.NotNil(t, alice.CreatedAt)
	require.Less(t, *alice.CreatedAt, time.Now().Add(-30*time.Minute), "creation date comes from the membership")

	require.Equal(t, membershipIDs["bob"].String(), records[1].ExternalID)
	require.Equal(t, new(false), records[1].IsAdmin)
	require.Equal(t, new(false), records[1].Active)

	require.Equal(t, membershipIDs["carol"].String(), records[2].ExternalID)
	require.Equal(t, new(false), records[2].Active)

	erin := records[3]
	require.Equal(t, membershipIDs["erin"].String(), erin.ExternalID)
	require.Empty(t, erin.Email)
	require.Equal(t, new(false), erin.Active)
	require.Equal(t, new(true), erin.IsAdmin)
}

// TestProboMembershipsDriverListAccountsAcrossPages pins that both walks
// follow their cursor past the first page without dropping or repeating
// a row. Every other member has no profile, so a membership fetched twice
// would surface as an extra record.
func TestProboMembershipsDriverListAccountsAcrossPages(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	scope, organizationID := insertProboMembersOrganization(t, ctx, client, gid.NewTenantID())

	members := make([]proboMember, 2*page.MaxCursorSize+1)
	for i := range members {
		members[i] = proboMember{
			Name: fmt.Sprintf("member%04d", i),
			Role: coredata.MembershipRoleViewer,
		}
		if i%2 == 0 {
			members[i].State = new(coredata.ProfileStateActive)
		}
	}

	insertProboMembers(t, ctx, client, scope, organizationID, map[string]gid.GID{}, members)

	records, err := NewProboMembershipsDriver(client, scope, organizationID).ListAccounts(ctx)
	require.NoError(t, err)
	require.Len(t, records, len(members))

	var (
		seen        = map[string]bool{}
		withProfile int
	)

	for _, record := range records {
		if record.Email != "" {
			withProfile++
		}

		seen[record.ExternalID] = true
	}

	require.Len(t, seen, len(members))
	require.Equal(t, (len(members)+1)/2, withProfile, "every profile must match its membership")
}
