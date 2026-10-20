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

func TestActivateAccount_ReusedIdentityDoesNotSetPassword(t *testing.T) {
	t.Parallel()

	orgA := testutil.NewClient(t, testutil.RoleOwner)
	orgB := testutil.NewClient(t, testutil.RoleOwner)
	email := factory.SafeEmail()

	profileA := factory.CreateUser(orgA, factory.Attrs{
		"emailAddress": email,
		"fullName":     "Existing User",
	})
	profileB := factory.CreateUser(orgB, factory.Attrs{
		"emailAddress": email,
		"fullName":     "Existing User",
	})
	factory.InviteUser(orgB, profileB)

	attacker := testutil.NewUnauthenticatedClient(t)
	invitationToken := attacker.PollForLinkToken(`to:` + email + ` subject:"Invitation to join"`)

	const passwordTokenMutation = `
		mutation($input: ActivateAccountFromInvitationInput!) {
			activateAccountFromInvitation(input: $input) {
				createPasswordToken
			}
		}
	`

	err := attacker.ExecuteConnectShouldFail(passwordTokenMutation, map[string]any{
		"input": map[string]any{
			"token": invitationToken,
		},
	})
	assert.ErrorContains(t, err, "createPasswordToken")

	const activateAccountMutation = `
		mutation($input: ActivateAccountFromInvitationInput!) {
			activateAccountFromInvitation(input: $input) {
				profile { id state }
			}
		}
	`

	var activated struct {
		ActivateAccountFromInvitation struct {
			Profile struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"profile"`
		} `json:"activateAccountFromInvitation"`
	}

	err = attacker.ExecuteConnect(activateAccountMutation, map[string]any{
		"input": map[string]any{
			"token": invitationToken,
		},
	}, &activated)
	require.NoError(t, err, "activateAccount mutation failed")
	assert.Equal(t, profileB, activated.ActivateAccountFromInvitation.Profile.ID)
	assert.Equal(t, "ACTIVE", activated.ActivateAccountFromInvitation.Profile.State)
	assert.Equal(t, "PENDING", profileState(t, orgA, profileA))

	const chosenPassword = "AttackerPassword123!"

	require.Error(t, attacker.SignIn(email, chosenPassword))

	const forgotPasswordMutation = `
		mutation($input: ForgotPasswordInput!) {
			forgotPassword(input: $input) {
				success
			}
		}
	`

	err = attacker.ExecuteConnect(forgotPasswordMutation, map[string]any{
		"input": map[string]any{
			"email": email,
		},
	}, nil)
	require.NoError(t, err, "forgotPassword mutation failed")

	resetToken := attacker.PollForLinkToken(`to:` + email + ` subject:"Reset your password"`)

	const resetPasswordMutation = `
		mutation($input: ResetPasswordInput!) {
			resetPassword(input: $input) {
				success
			}
		}
	`

	err = attacker.ExecuteConnect(resetPasswordMutation, map[string]any{
		"input": map[string]any{
			"token":    resetToken,
			"password": chosenPassword,
		},
	}, nil)
	require.NoError(t, err, "resetPassword mutation failed")
	require.NoError(t, attacker.SignIn(email, chosenPassword))
}

func profileState(t *testing.T, client *testutil.Client, profileID string) string {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on Profile { state }
			}
		}
	`

	var result struct {
		Node struct {
			State string `json:"state"`
		} `json:"node"`
	}

	err := client.ExecuteConnect(query, map[string]any{"id": profileID}, &result)
	require.NoError(t, err, "cannot load profile")

	return result.Node.State
}
