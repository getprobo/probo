// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyContact_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty first
	thirdPartyID := factory.NewThirdParty(owner).WithName("Contact Test ThirdParty").Create()

	query := `
		mutation CreateThirdPartyContact($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge {
					node {
						id
						fullName
						email
						phone
						role
					}
				}
			}
		}
	`

	var result struct {
		CreateThirdPartyContact struct {
			ThirdPartyContactEdge struct {
				Node struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
					Email    string `json:"email"`
					Phone    string `json:"phone"`
					Role     string `json:"role"`
				} `json:"node"`
			} `json:"thirdPartyContactEdge"`
		} `json:"createThirdPartyContact"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"fullName":     "John Doe",
			"email":        fmt.Sprintf("john.doe.%d@thirdParty.com", time.Now().UnixNano()),
			"phone":        "+1-555-123-4567",
			"role":         "Account Manager",
		},
	}, &result)
	require.NoError(t, err)

	contact := result.CreateThirdPartyContact.ThirdPartyContactEdge.Node
	assert.NotEmpty(t, contact.ID)
	assert.Equal(t, "John Doe", contact.FullName)
	assert.Equal(t, "+1-555-123-4567", contact.Phone)
	assert.Equal(t, "Account Manager", contact.Role)
}

func TestThirdPartyContact_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a thirdParty and contact
	thirdPartyID := factory.NewThirdParty(owner).WithName("Update Contact ThirdParty").Create()

	createQuery := `
		mutation CreateThirdPartyContact($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateThirdPartyContact struct {
			ThirdPartyContactEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyContactEdge"`
		} `json:"createThirdPartyContact"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"fullName":     "Initial Name",
			"email":        fmt.Sprintf("initial.%d@thirdParty.com", time.Now().UnixNano()),
		},
	}, &createResult)
	require.NoError(t, err)

	contactID := createResult.CreateThirdPartyContact.ThirdPartyContactEdge.Node.ID

	query := `
		mutation UpdateThirdPartyContact($input: UpdateThirdPartyContactInput!) {
			updateThirdPartyContact(input: $input) {
				thirdPartyContact {
					id
					fullName
					phone
					role
				}
			}
		}
	`

	var result struct {
		UpdateThirdPartyContact struct {
			ThirdPartyContact struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Phone    string `json:"phone"`
				Role     string `json:"role"`
			} `json:"thirdPartyContact"`
		} `json:"updateThirdPartyContact"`
	}

	err = owner.Execute(query, map[string]any{
		"input": map[string]any{
			"id":       contactID,
			"fullName": "Updated Name",
			"phone":    "+1-555-999-8888",
			"role":     "Senior Account Manager",
		},
	}, &result)
	require.NoError(t, err)

	assert.Equal(t, contactID, result.UpdateThirdPartyContact.ThirdPartyContact.ID)
	assert.Equal(t, "Updated Name", result.UpdateThirdPartyContact.ThirdPartyContact.FullName)
	assert.Equal(t, "+1-555-999-8888", result.UpdateThirdPartyContact.ThirdPartyContact.Phone)
	assert.Equal(t, "Senior Account Manager", result.UpdateThirdPartyContact.ThirdPartyContact.Role)
}

func TestThirdPartyContact_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner).WithName("Delete Contact ThirdParty").Create()

	// Create a contact to delete
	createQuery := `
		mutation CreateThirdPartyContact($input: CreateThirdPartyContactInput!) {
			createThirdPartyContact(input: $input) {
				thirdPartyContactEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateThirdPartyContact struct {
			ThirdPartyContactEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"thirdPartyContactEdge"`
		} `json:"createThirdPartyContact"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"thirdPartyId": thirdPartyID,
			"fullName":     fmt.Sprintf("Contact to Delete %d", time.Now().UnixNano()),
			"email":        fmt.Sprintf("delete.%d@thirdParty.com", time.Now().UnixNano()),
		},
	}, &createResult)
	require.NoError(t, err)

	contactID := createResult.CreateThirdPartyContact.ThirdPartyContactEdge.Node.ID

	deleteQuery := `
		mutation DeleteThirdPartyContact($input: DeleteThirdPartyContactInput!) {
			deleteThirdPartyContact(input: $input) {
				deletedThirdPartyContactId
			}
		}
	`

	var result struct {
		DeleteThirdPartyContact struct {
			DeletedThirdPartyContactID string `json:"deletedThirdPartyContactId"`
		} `json:"deleteThirdPartyContact"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"thirdPartyContactId": contactID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, contactID, result.DeleteThirdPartyContact.DeletedThirdPartyContactID)
}

func TestThirdPartyContact_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	thirdPartyID := factory.NewThirdParty(owner).WithName("List Contacts ThirdParty").Create()

	// Create multiple contacts
	for i := range 3 {
		query := `
			mutation CreateThirdPartyContact($input: CreateThirdPartyContactInput!) {
				createThirdPartyContact(input: $input) {
					thirdPartyContactEdge {
						node {
							id
						}
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"thirdPartyId": thirdPartyID,
				"fullName":     fmt.Sprintf("Contact %d", i),
				"email":        fmt.Sprintf("contact.%d.%d@thirdParty.com", i, time.Now().UnixNano()),
			},
		})
		require.NoError(t, err)
	}

	query := `
		query GetThirdPartyContacts($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					contacts(first: 10) {
						edges {
							node {
								id
								fullName
								email
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Contacts struct {
				Edges []struct {
					Node struct {
						ID       string `json:"id"`
						FullName string `json:"fullName"`
						Email    string `json:"email"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"contacts"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"id": thirdPartyID,
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Node.Contacts.Edges), 3)
}
