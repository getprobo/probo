// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type listedOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func organizationsByID(organizations []listedOrganization) map[string]string {
	namesByID := make(map[string]string, len(organizations))

	for _, organization := range organizations {
		namesByID[organization.ID] = organization.Name
	}

	return namesByID
}

func organizationNames(namesByID map[string]string) []string {
	names := make([]string, 0, len(namesByID))

	for _, name := range namesByID {
		names = append(names, name)
	}

	return names
}

func TestMCP_ListOrganizations(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	var result struct {
		Organizations []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"organizations"`
	}
	mc.CallToolInto("listOrganizations", map[string]any{}, &result)

	assert.NotEmpty(t, result.Organizations)
}

func TestMCP_ListOrganizations_LegalName(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	legalName := "Listed Org Legal Name Inc."
	err := owner.ExecuteConnect(`
		mutation UpdateOrganization($input: UpdateOrganizationInput!) {
			updateOrganization(input: $input) {
				organization { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"legalName":      legalName,
		},
	}, nil)
	require.NoError(t, err)

	var result struct {
		Organizations []struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			LegalName *string `json:"legal_name"`
		} `json:"organizations"`
	}
	mc.CallToolInto("listOrganizations", map[string]any{}, &result)

	var (
		found        bool
		gotName      string
		gotLegalName *string
	)

	for _, organization := range result.Organizations {
		if organization.ID != owner.GetOrganizationID().String() {
			continue
		}

		found = true
		gotName = organization.Name
		gotLegalName = organization.LegalName

		break
	}

	require.True(t, found)
	assert.NotEmpty(t, gotName)
	require.NotNil(t, gotLegalName)
	assert.Equal(t, legalName, *gotLegalName)
}

func TestMCP_ListOrganizations_ExcludesDeactivatedMembership(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	member := testutil.NewClient(t, testutil.RoleOwner)
	ownerOrgID := owner.GetOrganizationID().String()
	memberOrgID := member.GetOrganizationID().String()

	profileID := factory.CreateUser(owner, factory.Attrs{
		"emailAddress": member.GetEmail(),
	})
	factory.InviteUser(owner, profileID)

	var accepted struct {
		AcceptInvitation struct {
			Profile struct {
				ID    string `json:"id"`
				State string `json:"state"`
			} `json:"profile"`
		} `json:"acceptInvitation"`
	}

	err := member.ExecuteConnect(
		`
			mutation($input: AcceptInvitationInput!) {
				acceptInvitation(input: $input) {
					profile { id state }
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"organizationId": ownerOrgID,
			},
		},
		&accepted,
	)
	require.NoError(t, err)
	require.Equal(t, profileID, accepted.AcceptInvitation.Profile.ID)
	require.Equal(t, "ACTIVE", accepted.AcceptInvitation.Profile.State)

	var before struct {
		Organizations []listedOrganization `json:"organizations"`
	}
	testutil.NewMCPClient(t, member).CallToolInto("listOrganizations", map[string]any{}, &before)

	beforeByID := organizationsByID(before.Organizations)
	require.Contains(t, beforeByID, ownerOrgID)
	require.Contains(t, beforeByID, memberOrgID)

	deactivatedName := beforeByID[ownerOrgID]
	require.NotEmpty(t, deactivatedName)

	var deactivated struct {
		DeactivateUser struct {
			Success bool `json:"success"`
		} `json:"deactivateUser"`
	}

	err = owner.ExecuteConnect(
		`
			mutation($input: DeactivateUserInput!) {
				deactivateUser(input: $input) {
					success
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"organizationId": ownerOrgID,
				"profileId":      profileID,
			},
		},
		&deactivated,
	)
	require.NoError(t, err)
	require.True(t, deactivated.DeactivateUser.Success)

	token := member.CreateOAuth2AccessToken(
		"e2e-mcp-deactivated-membership",
		[]string{"v1:iam"},
	)

	var after struct {
		Organizations []listedOrganization `json:"organizations"`
	}
	testutil.NewMCPClientWithAccessToken(t, member, token).CallToolInto(
		"listOrganizations",
		map[string]any{},
		&after,
	)

	afterByID := organizationsByID(after.Organizations)
	assert.NotContains(t, afterByID, ownerOrgID)
	assert.NotContains(t, organizationNames(afterByID), deactivatedName)
	assert.Contains(t, afterByID, memberOrgID)
}
