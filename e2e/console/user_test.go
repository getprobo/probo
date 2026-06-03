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
	"testing"
	"time"

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

func TestUser_UpdateMembershipRejectsLastOwnerDemotion(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	queryProfiles := `
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

	var profilesResult struct {
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

	err := owner.ExecuteConnect(queryProfiles, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &profilesResult)
	require.NoError(t, err)

	var ownerMembershipID string

	for _, edge := range profilesResult.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "OWNER" {
			ownerMembershipID = edge.Node.Membership.ID

			break
		}
	}

	require.NotEmpty(t, ownerMembershipID, "Should find owner member")

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

	err = owner.ExecuteConnect(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"membershipId":   ownerMembershipID,
			"role":           "VIEWER",
		},
	}, nil)
	testutil.RequireErrorCode(t, err, "CONFLICT")

	var gqlErrors testutil.GraphQLErrors
	require.ErrorAs(t, err, &gqlErrors)
	assert.Equal(t, "cannot demote last active owner", gqlErrors[0].Message)

	queryMembership := `
		query($id: ID!) {
			node(id: $id) {
				... on Membership {
					id
					role
				}
			}
		}
	`

	var membershipResult struct {
		Node struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		} `json:"node"`
	}

	err = owner.ExecuteConnect(queryMembership, map[string]any{
		"id": ownerMembershipID,
	}, &membershipResult)
	require.NoError(t, err)
	assert.Equal(t, "OWNER", membershipResult.Node.Role)
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
								state
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
						State      string `json:"state"`
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

	// Removed user should no longer be returned.
	err = owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	var removedUserFound bool

	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.ID == userID {
			removedUserFound = true
			break
		}
	}

	assert.False(t, removedUserFound, "Should not find removed user")
}

func TestUser_ArchiveUser(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a user to archive.
	userToArchive := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	_ = userToArchive

	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 50) {
						edges {
							node {
								id
								state
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
						State      string `json:"state"`
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

	var userID string

	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "VIEWER" {
			userID = edge.Node.ID
			break
		}
	}

	require.NotEmpty(t, userID, "Should find viewer member")

	mutation := `
		mutation($input: ArchiveUserInput!) {
			archiveUser(input: $input) {
				archivedProfileId
			}
		}
	`

	var mutationResult struct {
		ArchiveUser struct {
			ArchivedProfileID string `json:"archivedProfileId"`
		} `json:"archiveUser"`
	}

	err = owner.ExecuteConnect(mutation, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      userID,
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, userID, mutationResult.ArchiveUser.ArchivedProfileID)

	err = owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	var archivedUserState string

	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.ID == userID {
			archivedUserState = edge.Node.State
			break
		}
	}

	require.NotEmpty(t, archivedUserState, "Should still find archived user")
	assert.Equal(t, "DEACTIVATED", archivedUserState)
}

func TestUser_DeactivateUserCancelsSignatureRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	documentVersionID := versionResult.Node.Versions.Edges[0].Node.ID
	signerProfileID := signer.GetProfileID().String()

	_, err = owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": documentVersionID,
			"signatoryId":       signerProfileID,
		},
	})
	require.NoError(t, err)

	assertRequestedSignatureCount(t, owner, documentVersionID, 1)

	var deactivateResult struct {
		DeactivateUser struct {
			Success bool `json:"success"`
		} `json:"deactivateUser"`
	}

	err = owner.ExecuteConnect(`
		mutation($input: DeactivateUserInput!) {
			deactivateUser(input: $input) {
				success
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"organizationId": owner.GetOrganizationID().String(),
			"profileId":      signerProfileID,
		},
	}, &deactivateResult)
	require.NoError(t, err)
	require.True(t, deactivateResult.DeactivateUser.Success)

	assertRequestedSignatureCount(t, owner, documentVersionID, 0)
}

func TestUser_EndedContractCancelsSignatureRequests(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	signer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	docID, _ := createTestDocument(t, owner)
	approveTestDocument(t, owner, docID)

	var versionResult struct {
		Node struct {
			Versions struct {
				Edges []struct {
					Node struct {
						ID string `json:"id"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"versions"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on Document {
					versions(first: 1, orderBy: { field: CREATED_AT, direction: DESC }) {
						edges { node { id } }
					}
				}
			}
		}
	`, map[string]any{"id": docID}, &versionResult)
	require.NoError(t, err)
	require.NotEmpty(t, versionResult.Node.Versions.Edges)

	documentVersionID := versionResult.Node.Versions.Edges[0].Node.ID
	signerProfileID := signer.GetProfileID().String()

	_, err = owner.Do(`
		mutation($input: RequestSignatureInput!) {
			requestSignature(input: $input) {
				documentVersionSignatureEdge { node { id } }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"documentVersionId": documentVersionID,
			"signatoryId":       signerProfileID,
		},
	})
	require.NoError(t, err)

	assertRequestedSignatureCount(t, owner, documentVersionID, 1)

	contractEndDate := time.Now().Add(-24 * time.Hour).Format(time.RFC3339)
	_, err = owner.DoConnect(`
		mutation($input: UpdateUserInput!) {
			updateUser(input: $input) {
				profile { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id":                       signerProfileID,
			"fullName":                 "Ended Contract Signer",
			"additionalEmailAddresses": []string{},
			"contractEndDate":          contractEndDate,
		},
	})
	require.NoError(t, err)

	assertRequestedSignatureCount(t, owner, documentVersionID, 0)
}

func assertRequestedSignatureCount(
	t *testing.T,
	owner *testutil.Client,
	documentVersionID string,
	expected int,
) {
	t.Helper()

	var result struct {
		Node struct {
			Signatures struct {
				TotalCount int `json:"totalCount"`
			} `json:"signatures"`
		} `json:"node"`
	}

	err := owner.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on DocumentVersion {
					signatures(first: 10, filter: { states: [REQUESTED] }) {
						totalCount
					}
				}
			}
		}
	`, map[string]any{"id": documentVersionID}, &result)
	require.NoError(t, err)
	assert.Equal(t, expected, result.Node.Signatures.TotalCount)
}

func TestUser_RemoveOwner(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create another owner to remove
	_ = testutil.NewClientInOrg(t, testutil.RoleOwner, owner)

	// Get the profiles
	query := `
		query($id: ID!) {
			node(id: $id) {
				... on Organization {
					profiles(first: 50) {
						edges {
							node {
								id
								identity {
									id
								}
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

	err := owner.ExecuteConnect(query, map[string]any{
		"id": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)

	// Find the other owner (not the calling owner)
	var targetProfileID string

	for _, edge := range result.Node.Profiles.Edges {
		if edge.Node.Membership.Role == "OWNER" && edge.Node.Identity.ID != owner.GetUserID().String() {
			targetProfileID = edge.Node.ID
			break
		}
	}

	require.NotEmpty(t, targetProfileID, "Should find another owner to remove")

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
			"profileId":      targetProfileID,
		},
	}, &mutationResult)
	require.NoError(t, err)

	assert.Equal(t, targetProfileID, mutationResult.RemoveUser.DeletedProfileID)
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
