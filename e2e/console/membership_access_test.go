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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const createOrganizationMutation = `
	mutation($input: CreateOrganizationInput!) {
		createOrganization(input: $input) {
			organization { id }
		}
	}
`

const consoleTypenameQuery = `
	query {
		__typename
	}
`

// TestMembershipAccess_DisableSignup covers the private-instance gate: when
// PROBOD_AUTH_DISABLE_SIGNUP is set, an authenticated identity without an
// active organization membership must not create organizations or reach the
// console API. Magic-link identity creation itself stays allowed so the
// compliance portal keeps working.
func TestMembershipAccess_DisableSignup(t *testing.T) {
	t.Parallel()

	env := testutil.StartIsolatedEnv(t, testutil.IsolatedEnvOptions{
		DisableSignup: true,
	})

	t.Run("membership_less_identity_cannot_create_organization", func(t *testing.T) {
		t.Parallel()

		client := testutil.NewUnauthenticatedClientFor(t, env.BaseURL)
		email := fmt.Sprintf("no-membership-%d@e2e.probo.test", time.Now().UnixNano())
		client.SignInWithMagicLink(email)

		err := client.ExecuteConnectShouldFail(
			createOrganizationMutation,
			map[string]any{
				"input": map[string]any{
					"name": fmt.Sprintf("Blocked Org %d", time.Now().UnixNano()),
				},
			},
		)
		testutil.RequireMembershipRequiredError(
			t,
			err,
			"membership-less identity must not create an organization when signup is disabled",
		)
	})

	t.Run("membership_less_identity_cannot_use_console", func(t *testing.T) {
		t.Parallel()

		client := testutil.NewUnauthenticatedClientFor(t, env.BaseURL)
		email := fmt.Sprintf("no-console-%d@e2e.probo.test", time.Now().UnixNano())
		client.SignInWithMagicLink(email)

		err := client.ExecuteShouldFail(consoleTypenameQuery, nil)
		testutil.RequireMembershipRequiredError(
			t,
			err,
			"membership-less identity must not reach the console API when signup is disabled",
		)
	})

	t.Run("active_member_can_create_organization", func(t *testing.T) {
		t.Parallel()

		// Bootstrap the org on the shared signup-enabled server, then open a
		// session against the disable-signup instance. Production private
		// instances are provisioned the same way: orgs exist before the flag
		// is flipped.
		owner := testutil.NewClient(t, testutil.RoleOwner)

		client := testutil.NewUnauthenticatedClientFor(t, env.BaseURL)
		client.SignInWithMagicLink(owner.GetEmail())

		var result struct {
			CreateOrganization struct {
				Organization struct {
					ID string `json:"id"`
				} `json:"organization"`
			} `json:"createOrganization"`
		}

		err := client.ExecuteConnect(
			createOrganizationMutation,
			map[string]any{
				"input": map[string]any{
					"name": fmt.Sprintf("Second Org %d", time.Now().UnixNano()),
				},
			},
			&result,
		)
		require.NoError(t, err, "active member must still be able to create another organization")
		require.NotEmpty(t, result.CreateOrganization.Organization.ID)
	})

	t.Run("deactivated_member_cannot_create_organization", func(t *testing.T) {
		t.Parallel()

		owner := testutil.NewClient(t, testutil.RoleOwner)
		member := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

		const deactivateMutation = `
			mutation($input: DeactivateUserInput!) {
				deactivateUser(input: $input) {
					success
				}
			}
		`

		var deactivateResult struct {
			DeactivateUser struct {
				Success bool `json:"success"`
			} `json:"deactivateUser"`
		}

		err := owner.ExecuteConnect(
			deactivateMutation,
			map[string]any{
				"input": map[string]any{
					"organizationId": owner.GetOrganizationID().String(),
					"profileId":      member.GetProfileID().String(),
				},
			},
			&deactivateResult,
		)
		require.NoError(t, err)
		require.True(t, deactivateResult.DeactivateUser.Success)

		client := testutil.NewUnauthenticatedClientFor(t, env.BaseURL)
		client.SignInWithMagicLink(member.GetEmail())

		err = client.ExecuteConnectShouldFail(
			createOrganizationMutation,
			map[string]any{
				"input": map[string]any{
					"name": fmt.Sprintf("Deactivated Org %d", time.Now().UnixNano()),
				},
			},
		)
		testutil.RequireMembershipRequiredError(
			t,
			err,
			"deactivated membership must not count toward the console access gate",
		)

		err = client.ExecuteShouldFail(consoleTypenameQuery, nil)
		testutil.RequireMembershipRequiredError(
			t,
			err,
			"deactivated membership must not unlock the console API",
		)
	})
}
