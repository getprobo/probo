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
