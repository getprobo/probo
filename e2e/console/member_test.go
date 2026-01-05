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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMember_UpdateMembership(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create an admin to update
	_ = testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	// Get the member ID of the admin
	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					members(first: 10) {
						edges {
							node {
								id
								role
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Members struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Role string `json:"role"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"members"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	// Find the admin member
	var adminMemberID string
	for _, edge := range result.Node.Members.Edges {
		if edge.Node.Role == "ADMIN" {
			adminMemberID = edge.Node.ID
			break
		}
	}
	require.NotEmpty(t, adminMemberID, "Should find admin member")

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
			"membershipId":   adminMemberID,
			"role":           "VIEWER",
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, "VIEWER", mutationResult.UpdateMembership.Membership.Role)
}

func TestMember_RemoveMember(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a member to remove
	memberToRemove := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	_ = memberToRemove

	// Get the member ID
	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					members(first: 50) {
						edges {
							node {
								id
								role
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Members struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Role string `json:"role"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"members"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	// Find a viewer member to remove
	var memberID string
	for _, edge := range result.Node.Members.Edges {
		if edge.Node.Role == "VIEWER" {
			memberID = edge.Node.ID
			break
		}
	}
	assert.NotEmpty(t, memberID, "Should find viewer member")

	// Remove the member
	mutation := `
		mutation($input: RemoveMemberInput!) {
			removeMember(input: $input) {
				deletedMembershipId
			}
		}
	`

	var mutationResult struct {
		RemoveMember struct {
			DeletedMembershipID string `json:"deletedMembershipId"`
		} `json:"removeMember"`
	}

	err = owner.ExecuteConnect(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"membershipId":   memberID,
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, memberID, mutationResult.RemoveMember.DeletedMembershipID)
}

func TestInvitation_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create an invitation
	inviteMutation := `
		mutation($input: InviteMemberInput!) {
			inviteMember(input: $input) {
				invitationEdge {
					node {
						id
						email
						status
					}
				}
			}
		}
	`

	var inviteResult struct {
		InviteMember struct {
			InvitationEdge struct {
				Node struct {
					ID     string `json:"id"`
					Email  string `json:"email"`
					Status string `json:"status"`
				} `json:"node"`
			} `json:"invitationEdge"`
		} `json:"inviteMember"`
	}

	err := owner.ExecuteConnect(inviteMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"email":          fmt.Sprintf("invite.delete.%d@example.com", time.Now().UnixNano()),
			"fullName":       "Test User",
			"role":           "VIEWER",
		},
	}, &inviteResult)
	require.NoError(t, err)

	invitationID := inviteResult.InviteMember.InvitationEdge.Node.ID
	assert.NotEmpty(t, invitationID)

	// Delete the invitation
	deleteMutation := `
		mutation($input: DeleteInvitationInput!) {
			deleteInvitation(input: $input) {
				deletedInvitationId
			}
		}
	`

	var deleteResult struct {
		DeleteInvitation struct {
			DeletedInvitationID string `json:"deletedInvitationId"`
		} `json:"deleteInvitation"`
	}

	err = owner.ExecuteConnect(deleteMutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"invitationId":   invitationID,
		},
	}, &deleteResult)
	require.NoError(t, err)

	assert.Equal(t, invitationID, deleteResult.DeleteInvitation.DeletedInvitationID)
}

func TestMember_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create additional members
	testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					members(first: 10) {
						edges {
							node {
								id
								role
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
			Members struct {
				Edges []struct {
					Node struct {
						ID   string `json:"id"`
						Role string `json:"role"`
					} `json:"node"`
				} `json:"edges"`
				TotalCount int `json:"totalCount"`
			} `json:"members"`
		} `json:"node"`
	}

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, result.Node.Members.TotalCount, 3, "Should have at least 3 members")
}
