// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestUser_UpdateMembership(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create an admin to update
	_ = testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	// Get the user ID of the admin
	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 10) {
						edges {
							node {
								membership {
									id
									role
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						Membership struct {
							ID   string `json:"id"`
							Role string `json:"role"`
						} `json:"membership"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	// Find the admin
	var adminMembershipID string
	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "ADMIN" {
			adminMembershipID = edge.Node.Membership.ID
			break
		}
	}
	require.NotEmpty(t, adminMembershipID, "Should find admin member")

	// Update the member role to VIEWER
	mutation := `
		mutation($input: UpdateMembershipInput!) {
			updateMembership(input: $input) {
				membership {
					id
					role
				}
			}
		}
	`

	var mutationResult struct {
		UpdateMembership struct {
			Membership struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"membership"`
		} `json:"updateMembership"`
	}

	err = owner.ExecuteConnect(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"membershipId":   adminMembershipID,
			"role":           "VIEWER",
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, "VIEWER", mutationResult.UpdateMembership.Membership.Role)
}

func TestUser_RemoveUser(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a user to remove
	userToRemove := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	_ = userToRemove

	// Get the user ID
	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 50) {
						edges {
							node {
								id
								membership {
									role
								}
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						Membership struct {
							Role string `json:"role"`
						} `json:"membership"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"profiles"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	// Find a viewer user to remove
	var userID string
	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "VIEWER" {
			userID = edge.Node.ID
			break
		}
	}
	assert.NotEmpty(t, userID, "Should find viewer member")

	// Remove the member
	mutation := `
		mutation($input: RemoveUserInput!) {
			removeUser(input: $input) {
				deletedProfileId
			}
		}
	`

	var mutationResult struct {
		RemoveUser struct {
			DeletedProfileID string `json:"deletedProfileId"`
		} `json:"removeUser"`
	}

	err = owner.ExecuteConnect(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      userID,
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, userID, mutationResult.RemoveUser.DeletedProfileID)
}

func TestUser_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create additional members
	testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 10) {
						edges {
							node {
								id
								membership {
									role
								}
							}
						}
						totalCount
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Profiles struct {
				Edges []struct {
					Node struct {
						ID         string `json:"id"`
						Membership struct {
							Role string `json:"role"`
						} `json:"membership"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"profiles"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Profiles.TotalCount, 3, "Should have at least 3 members")
}
