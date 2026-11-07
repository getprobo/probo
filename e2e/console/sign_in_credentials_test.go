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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestSignIn_PasswordlessAccountsMatchUnknownEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	caller := testutil.NewUnauthenticatedClient(t)
	password := "TestPassword123!"

	pendingEmail := factory.SafeEmail()
	createPendingUser(t, owner, pendingEmail)

	sc := newSCIMClient(t, owner)
	deactivatedEmail := factory.SafeEmail()
	body, status := sc.createUser(deactivatedEmail, "Deactivated User", "ext-signin-"+factory.SafeName(""), false)
	require.Equal(t, http.StatusCreated, status, body)

	unknownEmail := factory.SafeEmail()

	pendingErr := caller.SignIn(pendingEmail, password)
	deactivatedErr := caller.SignIn(deactivatedEmail, password)
	unknownErr := caller.SignIn(unknownEmail, password)

	assertSameInvalidCredentials(t, unknownErr, pendingErr)
	assertSameInvalidCredentials(t, unknownErr, deactivatedErr)
}

func createPendingUser(t *testing.T, owner *testutil.Client, email string) {
	t.Helper()

	const query = `
		mutation($input: CreateUserInput!) {
			createUser(input: $input) {
				profileEdge { node { id } }
			}
		}
	`

	var result struct {
		CreateUser struct {
			ProfileEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"profileEdge"`
		} `json:"createUser"`
	}

	err := owner.ExecuteConnect(
		query,
		map[string]any{
			"input": map[string]any{
				"organizationId":           owner.GetOrganizationID().String(),
				"emailAddress":             email,
				"fullName":                 factory.SafeName("Pending User"),
				"role":                     "VIEWER",
				"kind":                     "EMPLOYEE",
				"additionalEmailAddresses": []string{},
			},
		},
		&result,
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.CreateUser.ProfileEdge.Node.ID)
}

func assertSameInvalidCredentials(t *testing.T, unknown error, got error) {
	t.Helper()

	testutil.RequireErrorCode(t, unknown, "INVALID_CREDENTIALS")
	testutil.RequireErrorCode(t, got, "INVALID_CREDENTIALS")

	var unknownErrors testutil.GraphQLErrors
	var gotErrors testutil.GraphQLErrors
	require.ErrorAs(t, unknown, &unknownErrors)
	require.ErrorAs(t, got, &gotErrors)
	require.NotEmpty(t, unknownErrors)
	require.NotEmpty(t, gotErrors)
	assert.Equal(t, unknownErrors[0].Message, gotErrors[0].Message)
	assert.NotEqual(t, "INTERNAL", gotErrors[0].Code())
}
