// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const invitingOrganizationsQuery = `
	query {
		viewer {
			invitingOrganizations {
				id
				name
			}
		}
	}
`

type invitingOrganizationsResult struct {
	Viewer struct {
		InvitingOrganizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"invitingOrganizations"`
	} `json:"viewer"`
}

func TestInvitingOrganizations_List(t *testing.T) {
	t.Parallel()

	t.Run("includes orgs with a live pending invitation", func(t *testing.T) {
		t.Parallel()

		inviter := testutil.NewClient(t, testutil.RoleOwner)
		invitee := testutil.NewClient(t, testutil.RoleOwner)

		profileID := factory.CreateUser(inviter, factory.Attrs{
			"emailAddress": invitee.GetEmail(),
		})
		factory.InviteUser(inviter, profileID)

		var result invitingOrganizationsResult

		err := invitee.ExecuteConnect(invitingOrganizationsQuery, nil, &result)
		require.NoError(t, err)

		require.Len(t, result.Viewer.InvitingOrganizations, 1)
		assert.Equal(t, inviter.GetOrganizationID().String(), result.Viewer.InvitingOrganizations[0].ID)
	})

	t.Run("is empty when no invitation was sent", func(t *testing.T) {
		t.Parallel()

		inviter := testutil.NewClient(t, testutil.RoleOwner)
		invitee := testutil.NewClient(t, testutil.RoleOwner)

		// Profile created but no invitation sent — invitation is the trigger.
		factory.CreateUser(inviter, factory.Attrs{
			"emailAddress": invitee.GetEmail(),
		})

		var result invitingOrganizationsResult

		err := invitee.ExecuteConnect(invitingOrganizationsQuery, nil, &result)
		require.NoError(t, err)

		assert.Empty(t, result.Viewer.InvitingOrganizations)
	})

	t.Run("is empty for an identity with no invitations", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)

		var result invitingOrganizationsResult

		err := owner.ExecuteConnect(invitingOrganizationsQuery, nil, &result)
		require.NoError(t, err)

		assert.Empty(t, result.Viewer.InvitingOrganizations)
	})

	t.Run("excludes orgs once the invitation is accepted", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)

		// NewClientInOrg goes through the full invite + activation flow, so the
		// invitation is in ACCEPTED state by the time it returns.
		member := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		var result invitingOrganizationsResult

		err := member.ExecuteConnect(invitingOrganizationsQuery, nil, &result)
		require.NoError(t, err)

		assert.Empty(t, result.Viewer.InvitingOrganizations)
	})

	t.Run("lists multiple orgs when invited by several", func(t *testing.T) {
		t.Parallel()

		inviter1 := testutil.NewClient(t, testutil.RoleOwner)
		inviter2 := testutil.NewClient(t, testutil.RoleOwner)
		invitee := testutil.NewClient(t, testutil.RoleOwner)

		profile1 := factory.CreateUser(inviter1, factory.Attrs{"emailAddress": invitee.GetEmail()})
		factory.InviteUser(inviter1, profile1)

		profile2 := factory.CreateUser(inviter2, factory.Attrs{"emailAddress": invitee.GetEmail()})
		factory.InviteUser(inviter2, profile2)

		var result invitingOrganizationsResult

		err := invitee.ExecuteConnect(invitingOrganizationsQuery, nil, &result)
		require.NoError(t, err)

		ids := make(map[string]struct{}, len(result.Viewer.InvitingOrganizations))
		for _, org := range result.Viewer.InvitingOrganizations {
			ids[org.ID] = struct{}{}
		}

		assert.Contains(t, ids, inviter1.GetOrganizationID().String())
		assert.Contains(t, ids, inviter2.GetOrganizationID().String())
	})
}
