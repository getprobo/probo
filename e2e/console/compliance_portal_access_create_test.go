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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const createCompliancePortalAccessMutation = `
	mutation($input: CreateCompliancePortalAccessInput!) {
		createCompliancePortalAccess(input: $input) {
			compliancePortalAccessEdge {
				node {
					id
					state
					authenticatedAt
					profile {
						id
						fullName
						emailAddress
					}
				}
			}
		}
	}
`

const memberCandidatesQuery = `
	query($id: ID!, $query: String!) {
		node(id: $id) {
			... on CompliancePortal {
				memberCandidates(query: $query) {
					id
					fullName
					emailAddress
				}
			}
		}
	}
`

type compliancePortalAccessNode struct {
	ID              string  `json:"id"`
	State           string  `json:"state"`
	AuthenticatedAt *string `json:"authenticatedAt"`
	Profile         struct {
		ID           string `json:"id"`
		FullName     string `json:"fullName"`
		EmailAddress string `json:"emailAddress"`
	} `json:"profile"`
}

type createCompliancePortalAccessResult struct {
	CreateCompliancePortalAccess struct {
		CompliancePortalAccessEdge struct {
			Node compliancePortalAccessNode `json:"node"`
		} `json:"compliancePortalAccessEdge"`
	} `json:"createCompliancePortalAccess"`
}

func createCompliancePortalAccess(
	t *testing.T,
	client *testutil.Client,
	input map[string]any,
) compliancePortalAccessNode {
	t.Helper()

	var result createCompliancePortalAccessResult

	err := client.Execute(createCompliancePortalAccessMutation, map[string]any{"input": input}, &result)
	require.NoError(t, err)

	node := result.CreateCompliancePortalAccess.CompliancePortalAccessEdge.Node
	require.NotEmpty(t, node.ID)

	return node
}

func listMemberCandidateIDs(
	t *testing.T,
	client *testutil.Client,
	compliancePortalID string,
	query string,
) []string {
	t.Helper()

	var result struct {
		Node struct {
			MemberCandidates []struct {
				ID           string `json:"id"`
				FullName     string `json:"fullName"`
				EmailAddress string `json:"emailAddress"`
			} `json:"memberCandidates"`
		} `json:"node"`
	}

	err := client.Execute(memberCandidatesQuery, map[string]any{
		"id":    compliancePortalID,
		"query": query,
	}, &result)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Node.MemberCandidates))
	for _, candidate := range result.Node.MemberCandidates {
		ids = append(ids, candidate.ID)
	}

	return ids
}

func requireAccessMailpitMessageEventually(
	t *testing.T,
	client *testutil.Client,
	recipientEmail string,
) *testutil.MailpitMessageDetail {
	t.Helper()

	expectedSubject := fmt.Sprintf(
		"Compliance Portal Access Invitation - %s",
		queryOrganizationName(t, client),
	)
	searchQuery := fmt.Sprintf("to:%s", recipientEmail)

	var (
		detail  *testutil.MailpitMessageDetail
		lastErr error
	)

	ok := testutil.Poll(
		t,
		90*time.Second,
		500*time.Millisecond,
		func() bool {
			detail, lastErr = client.FindMailpitMessage(
				searchQuery,
				func(message *testutil.MailpitMessageDetail) bool {
					return message.Subject == expectedSubject
				},
			)

			return lastErr == nil && detail != nil
		},
	)
	if !ok {
		if lastErr != nil {
			t.Logf("last mailpit access search did not find a matching message: %v", lastErr)
		}

		require.FailNow(t, "mailpit compliance portal access email not found")
	}

	assert.Contains(t, detail.Text, "probopage.localhost")

	return detail
}

func TestCompliancePortalAccess_CreateByProfileID(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	fullName := factory.SafeName("PortalVisitor")
	profileID := factory.CreateUser(owner, factory.Attrs{
		"fullName": fullName,
	})

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"profileId":          profileID,
	})

	assert.Equal(t, "ACTIVE", node.State)
	assert.Nil(t, node.AuthenticatedAt)
	assert.Equal(t, profileID, node.Profile.ID)
	assert.Equal(t, fullName, node.Profile.FullName)
}

func TestCompliancePortalAccess_CreateByNewEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	assert.Equal(t, "ACTIVE", node.State)
	assert.Nil(t, node.AuthenticatedAt)
	assert.Equal(t, email, node.Profile.EmailAddress)
	assert.Empty(t, node.Profile.FullName)
}

func TestCompliancePortalAccess_CreateUpsertsExistingProfile(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()
	fullName := factory.SafeName("ExistingVisitor")
	profileID := factory.CreateUser(owner, factory.Attrs{
		"emailAddress": email,
		"fullName":     fullName,
	})

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	assert.Equal(t, profileID, node.Profile.ID)
	assert.Equal(t, fullName, node.Profile.FullName)
	assert.Equal(t, email, node.Profile.EmailAddress)
}

func TestCompliancePortalAccess_CreateConflict(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()

	createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	err := owner.ExecuteShouldFail(createCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"email":              email,
		},
	})
	require.Error(t, err)
}

func TestCompliancePortalAccess_CreateDoesNotQueueAccessEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()
	searchQuery := fmt.Sprintf("to:%s", email)

	createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	var lastErr error

	foundMail := testutil.Poll(
		t,
		10*time.Second,
		500*time.Millisecond,
		func() bool {
			mails, err := owner.SearchMails(searchQuery)
			lastErr = err

			return err == nil && len(mails.Messages) > 0
		},
	)
	if lastErr != nil {
		t.Logf("last mailpit search failed: %v", lastErr)
	}

	require.NoError(t, lastErr)
	assert.False(t, foundMail)
}

func TestCompliancePortalAccess_GrantQueuesAccessEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()
	documentID := factory.NewDocument(owner).WithTitle("Visitor grant email").Create()
	publishDocumentMinor(t, owner, documentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, documentID)

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
	})

	grantDocumentAccess(t, owner, node.ID, documentID)
	requireAccessMailpitMessageEventually(t, owner, email)
}

func TestCompliancePortalAccess_CreateWithDocumentQueuesAccessEmail(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	email := factory.SafeEmail()
	documentID := factory.NewDocument(owner).WithTitle("Visitor create grant email").Create()
	publishDocumentMinor(t, owner, documentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, documentID)

	createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"email":              email,
		"documents":          []string{documentID},
	})

	requireAccessMailpitMessageEventually(t, owner, email)
}

func TestCompliancePortalAccess_CreateDeactivatedEmployee(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	compliancePortalID := compliancePortalID(t, owner)

	require.NoError(t, deactivateUser(t, owner, employee))

	profileID := employee.GetProfileID().String()
	ids := listMemberCandidateIDs(t, owner, compliancePortalID, employee.GetEmail())
	assert.Contains(t, ids, profileID)

	node := createCompliancePortalAccess(t, owner, map[string]any{
		"compliancePortalId": compliancePortalID,
		"profileId":          profileID,
	})

	assert.Equal(t, profileID, node.Profile.ID)
	assert.Equal(t, "ACTIVE", node.State)
	assert.Nil(t, node.AuthenticatedAt)
}

func TestCompliancePortalAccess_MemberCandidatesAccessManager(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	manager := testutil.NewClientInOrg(t, testutil.RoleCompliancePortalAccessManager, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	compliancePortalID := compliancePortalID(t, owner)

	fullName := factory.SafeName("TypeaheadCand")
	profileID := factory.CreateUser(owner, factory.Attrs{
		"fullName": fullName,
	})

	ids := listMemberCandidateIDs(t, manager, compliancePortalID, fullName)
	assert.Contains(t, ids, profileID)

	err := viewer.ExecuteShouldFail(memberCandidatesQuery, map[string]any{
		"id":    compliancePortalID,
		"query": fullName,
	})
	require.Error(t, err)

	createCompliancePortalAccess(t, manager, map[string]any{
		"compliancePortalId": compliancePortalID,
		"profileId":          profileID,
	})

	ids = listMemberCandidateIDs(t, manager, compliancePortalID, fullName)
	assert.NotContains(t, ids, profileID)
}

func TestCompliancePortalAccess_CreateTenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	org1CompliancePortalID := compliancePortalID(t, org1Owner)

	err := org2Owner.ExecuteShouldFail(createCompliancePortalAccessMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": org1CompliancePortalID,
			"email":              factory.SafeEmail(),
		},
	})
	require.Error(t, err)
}
