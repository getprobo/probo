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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestVendorContact_Create(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor first
	vendorID := factory.NewVendor(owner).WithName("Contact Test Vendor").Create()

	query := `
		mutation CreateVendorContact($input: CreateVendorContactInput!) {
			createVendorContact(input: $input) {
				vendorContactEdge {
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
		CreateVendorContact struct {
			VendorContactEdge struct {
				Node struct {
					ID       string `json:"id"`
					FullName string `json:"fullName"`
					Email    string `json:"email"`
					Phone    string `json:"phone"`
					Role     string `json:"role"`
				} `json:"node"`
			} `json:"vendorContactEdge"`
		} `json:"createVendorContact"`
	}

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"vendorId": vendorID,
			"fullName": "John Doe",
			"email":    fmt.Sprintf("john.doe.%d@vendor.com", time.Now().UnixNano()),
			"phone":    "+1-555-123-4567",
			"role":     "Account Manager",
		},
	}, &result)
	require.NoError(t, err)

	contact := result.CreateVendorContact.VendorContactEdge.Node
	assert.NotEmpty(t, contact.ID)
	assert.Equal(t, "John Doe", contact.FullName)
	assert.Equal(t, "+1-555-123-4567", contact.Phone)
	assert.Equal(t, "Account Manager", contact.Role)
}

func TestVendorContact_Update(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	// Create a vendor and contact
	vendorID := factory.NewVendor(owner).WithName("Update Contact Vendor").Create()

	createQuery := `
		mutation CreateVendorContact($input: CreateVendorContactInput!) {
			createVendorContact(input: $input) {
				vendorContactEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateVendorContact struct {
			VendorContactEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorContactEdge"`
		} `json:"createVendorContact"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"vendorId": vendorID,
			"fullName": "Initial Name",
			"email":    fmt.Sprintf("initial.%d@vendor.com", time.Now().UnixNano()),
		},
	}, &createResult)
	require.NoError(t, err)

	contactID := createResult.CreateVendorContact.VendorContactEdge.Node.ID

	query := `
		mutation UpdateVendorContact($input: UpdateVendorContactInput!) {
			updateVendorContact(input: $input) {
				vendorContact {
					id
					fullName
					phone
					role
				}
			}
		}
	`

	var result struct {
		UpdateVendorContact struct {
			VendorContact struct {
				ID       string `json:"id"`
				FullName string `json:"fullName"`
				Phone    string `json:"phone"`
				Role     string `json:"role"`
			} `json:"vendorContact"`
		} `json:"updateVendorContact"`
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

	assert.Equal(t, contactID, result.UpdateVendorContact.VendorContact.ID)
	assert.Equal(t, "Updated Name", result.UpdateVendorContact.VendorContact.FullName)
	assert.Equal(t, "+1-555-999-8888", result.UpdateVendorContact.VendorContact.Phone)
	assert.Equal(t, "Senior Account Manager", result.UpdateVendorContact.VendorContact.Role)
}

func TestVendorContact_Delete(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(owner).WithName("Delete Contact Vendor").Create()

	// Create a contact to delete
	createQuery := `
		mutation CreateVendorContact($input: CreateVendorContactInput!) {
			createVendorContact(input: $input) {
				vendorContactEdge {
					node {
						id
					}
				}
			}
		}
	`

	var createResult struct {
		CreateVendorContact struct {
			VendorContactEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"vendorContactEdge"`
		} `json:"createVendorContact"`
	}

	err := owner.Execute(createQuery, map[string]any{
		"input": map[string]any{
			"vendorId": vendorID,
			"fullName": fmt.Sprintf("Contact to Delete %d", time.Now().UnixNano()),
			"email":    fmt.Sprintf("delete.%d@vendor.com", time.Now().UnixNano()),
		},
	}, &createResult)
	require.NoError(t, err)

	contactID := createResult.CreateVendorContact.VendorContactEdge.Node.ID

	deleteQuery := `
		mutation DeleteVendorContact($input: DeleteVendorContactInput!) {
			deleteVendorContact(input: $input) {
				deletedVendorContactId
			}
		}
	`

	var result struct {
		DeleteVendorContact struct {
			DeletedVendorContactID string `json:"deletedVendorContactId"`
		} `json:"deleteVendorContact"`
	}

	err = owner.Execute(deleteQuery, map[string]any{
		"input": map[string]any{
			"vendorContactId": contactID,
		},
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, contactID, result.DeleteVendorContact.DeletedVendorContactID)
}

func TestVendorContact_List(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)

	vendorID := factory.NewVendor(owner).WithName("List Contacts Vendor").Create()

	// Create multiple contacts
	for i := 0; i < 3; i++ {
		query := `
			mutation CreateVendorContact($input: CreateVendorContactInput!) {
				createVendorContact(input: $input) {
					vendorContactEdge {
						node {
							id
						}
					}
				}
			}
		`

		_, err := owner.Do(query, map[string]any{
			"input": map[string]any{
				"vendorId": vendorID,
				"fullName": fmt.Sprintf("Contact %d", i),
				"email":    fmt.Sprintf("contact.%d.%d@vendor.com", i, time.Now().UnixNano()),
			},
		})
		require.NoError(t, err)
	}

	query := `
		query GetVendorContacts($id: ID!) {
			node(id: $id) {
				... on Vendor {
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
		"id": vendorID,
	}, &result)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Node.Contacts.Edges), 3)
}
