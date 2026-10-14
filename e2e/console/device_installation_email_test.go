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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
)

const sendDeviceInstallationEmailsMutation = `
mutation SendDeviceInstallationEmails($input: SendDeviceInstallationEmailsInput!) {
  sendDeviceInstallationEmails(input: $input) {
    recipientCount
  }
}`

func TestDeviceInstallationEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	uncovered := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	covered := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	enrollAndActivateDevice(t, covered, orgID)
	contractorEmail := createContractor(t, owner)

	var first struct {
		SendDeviceInstallationEmails struct {
			RecipientCount int `json:"recipientCount"`
		} `json:"sendDeviceInstallationEmails"`
	}
	owner.MustExecute(
		sendDeviceInstallationEmailsMutation,
		map[string]any{"input": map[string]any{"organizationId": orgID}},
		&first,
	)
	require.GreaterOrEqual(t, first.SendDeviceInstallationEmails.RecipientCount, 1)
	require.Equal(t, 1, countInstallationEmails(t, uncovered.GetEmail()))
	require.Equal(t, 0, countInstallationEmails(t, covered.GetEmail()))
	require.Equal(t, 0, countInstallationEmails(t, contractorEmail))

	owner.MustExecute(
		sendDeviceInstallationEmailsMutation,
		map[string]any{"input": map[string]any{"organizationId": orgID}},
		&first,
	)
	require.Equal(t, 2, countInstallationEmails(t, uncovered.GetEmail()))
}

func TestDeviceInstallationEmail_Forbidden(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	other := testutil.NewClient(t, testutil.RoleOwner)

	input := map[string]any{"input": map[string]any{"organizationId": orgID}}

	_, err := viewer.Do(sendDeviceInstallationEmailsMutation, input)
	testutil.RequireForbiddenError(t, err, "viewer cannot send installation emails")

	_, err = employee.Do(sendDeviceInstallationEmailsMutation, input)
	testutil.RequireForbiddenError(t, err, "employee cannot send installation emails")

	_, err = other.Do(sendDeviceInstallationEmailsMutation, input)
	testutil.RequireForbiddenError(t, err, "other organization cannot send installation emails")
}

func createContractor(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	email := fmt.Sprintf("contractor-%d@example.com", time.Now().UnixNano())

	const query = `
		mutation($input: CreateUserInput!) {
			createUser(input: $input) {
				profileEdge { node { id } }
			}
		}`

	var result struct {
		CreateUser struct {
			ProfileEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"profileEdge"`
		} `json:"createUser"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"input": map[string]any{
			"organizationId":           owner.GetOrganizationID().String(),
			"emailAddress":             email,
			"fullName":                 "Contract Person",
			"role":                     "VIEWER",
			"kind":                     "CONTRACTOR",
			"additionalEmailAddresses": []string{},
		},
	}, &result)
	require.NoError(t, err, "create contractor")
	require.NotEmpty(t, result.CreateUser.ProfileEdge.Node.ID)

	return email
}

func countInstallationEmails(t *testing.T, email string) int {
	t.Helper()

	var count int

	err := test.PGClient(t).WithConn(
		context.Background(),
		func(ctx context.Context, conn pg.Querier) error {
			return conn.QueryRow(
				ctx,
				`
SELECT COUNT(*)
FROM emails
WHERE recipient_email = $1
  AND subject LIKE 'Install the Probo device agent for %'
`,
				email,
			).Scan(&count)
		},
	)
	require.NoError(t, err, "count installation emails")

	return count
}
