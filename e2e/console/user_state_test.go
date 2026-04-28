// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const deactivateUserMutation = `
	mutation($input: DeactivateUserInput!) {
		deactivateUser(input: $input) {
			profile {
				id
				state
			}
		}
	}
`

const activateUserMutation = `
	mutation($input: ActivateUserInput!) {
		activateUser(input: $input) {
			profile {
				id
				state
			}
		}
	}
`

type userStateMutationResult struct {
	Profile struct {
		ID    string `json:"id"`
		State string `json:"state"`
	} `json:"profile"`
}

func TestUser_DeactivateAndActivate(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	target := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	// Deactivate the viewer.
	var deactivateResp struct {
		DeactivateUser userStateMutationResult `json:"deactivateUser"`
	}
	err := owner.ExecuteConnect(deactivateUserMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      target.GetProfileID().String(),
		},
	}, &deactivateResp)
	require.NoError(t, err)
	assert.Equal(t, target.GetProfileID().String(), deactivateResp.DeactivateUser.Profile.ID)
	assert.Equal(t, "INACTIVE", deactivateResp.DeactivateUser.Profile.State)

	// Reactivate the viewer.
	var activateResp struct {
		ActivateUser userStateMutationResult `json:"activateUser"`
	}
	err = owner.ExecuteConnect(activateUserMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      target.GetProfileID().String(),
		},
	}, &activateResp)
	require.NoError(t, err)
	assert.Equal(t, target.GetProfileID().String(), activateResp.ActivateUser.Profile.ID)
	assert.Equal(t, "ACTIVE", activateResp.ActivateUser.Profile.State)
}

func TestUser_DeactivateNoop(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	target := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	for i := 0; i < 2; i++ {
		var resp struct {
			DeactivateUser userStateMutationResult `json:"deactivateUser"`
		}
		err := owner.ExecuteConnect(deactivateUserMutation, map[string]any{
			"input": map[string]any{
				"organizationId": owner.GetOrganizationID().String(),
				"profileId":      target.GetProfileID().String(),
			},
		}, &resp)
		require.NoError(t, err)
		assert.Equal(t, "INACTIVE", resp.DeactivateUser.Profile.State)
	}
}

func TestUser_DeactivateLastActiveOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Find the owner's profile id.
	const profilesQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 50) {
						edges {
							node {
								id
								identity { id }
								membership { role }
							}
						}
					}
				}
			}
		}
	`

	var profilesResp struct {
		Node struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						Identity struct {
							ID string `json:"id"`
						} `json:"identity"`
						Membership struct {
							Role string `json:"role"`
						} `json:"membership"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(profilesQuery, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &profilesResp)
	require.NoError(t, err)

	var ownerProfileID string
	for _, edge := range profilesResp.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "OWNER" && edge.Node.Identity.ID == owner.GetUserID().String() {
			ownerProfileID = edge.Node.ID
			break
		}
	}
	require.NotEmpty(t, ownerProfileID)

	// Attempting to deactivate the last active owner must fail with a conflict.
	err = owner.ExecuteConnect(deactivateUserMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      ownerProfileID,
		},
	}, nil)

	require.Error(t, err)
	var gqlErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &gqlErrors)
	require.NotEmpty(t, gqlErrors)
	assert.Equal(t, "CONFLICT", gqlErrors[0].Code(), "got code=%q message=%q", gqlErrors[0].Code(), gqlErrors[0].Message)
	assert.True(t,
		strings.Contains(gqlErrors[0].Message, "last active owner"),
		"expected last active owner message, got: %q", gqlErrors[0].Message,
	)
}

func TestUser_DeactivateForbiddenForViewer(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	target := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	err := viewer.ExecuteConnect(deactivateUserMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      target.GetProfileID().String(),
		},
	}, nil)

	require.Error(t, err)
	var gqlErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &gqlErrors)
	require.NotEmpty(t, gqlErrors)
	code := gqlErrors[0].Code()
	msg := gqlErrors[0].Message
	isForbidden := code == "FORBIDDEN" ||
		(code == "" && (strings.Contains(msg, "does not have sufficient permissions") ||
			strings.Contains(msg, "insufficient permissions")))
	assert.True(t, isForbidden, "expected FORBIDDEN error, got code=%q message=%q", code, msg)
}

func TestUser_ActivateForbiddenForViewer(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	target := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)

	err := viewer.ExecuteConnect(activateUserMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      target.GetProfileID().String(),
		},
	}, nil)

	require.Error(t, err)
	var gqlErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &gqlErrors)
	require.NotEmpty(t, gqlErrors)
	code := gqlErrors[0].Code()
	msg := gqlErrors[0].Message
	isForbidden := code == "FORBIDDEN" ||
		(code == "" && (strings.Contains(msg, "does not have sufficient permissions") ||
			strings.Contains(msg, "insufficient permissions")))
	assert.True(t, isForbidden, "expected FORBIDDEN error, got code=%q message=%q", code, msg)
}
