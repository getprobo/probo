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
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMeasure_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createInternalControlMutation = `
		mutation CreateInternalControl($input: CreateInternalControlInput!) {
			createInternalControl(input: $input) {
				internalControlEdge { node { id } }
			}
		}
	`

	const updateInternalControlMutation = `
		mutation UpdateInternalControl($input: UpdateInternalControlInput!) {
			updateInternalControl(input: $input) {
				internalControl { id }
			}
		}
	`

	const deleteInternalControlMutation = `
		mutation DeleteInternalControl($input: DeleteInternalControlInput!) {
			deleteInternalControl(input: $input) {
				deletedInternalControlId
			}
		}
	`

	const readInternalControlQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on InternalControl { id name }
			}
		}
	`

	t.Run(
		"create",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can create",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to create internal control",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create internal control",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create internal control",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createInternalControlMutation,
							map[string]any{
								"input": map[string]any{
									"organizationId": client.GetOrganizationID().String(),
									"name":           factory.SafeName("RBAC create " + string(tt.role)),
									"category":       "POLICY",
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				updateName   string
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:       "owner can update",
					role:       testutil.RoleOwner,
					updateName: "Updated by Owner",
					allowMsg:   "owner should be able to update internal control",
				},
				{
					name:       "admin can update",
					role:       testutil.RoleAdmin,
					updateName: "Updated by Admin",
					allowMsg:   "admin should be able to update internal control",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					updateName:   "Updated by Viewer",
					forbiddenMsg: "viewer should not be able to update internal control",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						internalControlID := factory.NewInternalControl(ownerClient).
							WithName(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateInternalControlMutation,
							map[string]any{
								"input": map[string]any{
									"id":   internalControlID,
									"name": tt.updateName,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name         string
				role         testutil.TestRole
				forbidden    bool
				allowMsg     string
				forbiddenMsg string
			}{
				{
					name:     "owner can delete",
					role:     testutil.RoleOwner,
					allowMsg: "owner should be able to delete internal control",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete internal control",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete internal control",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						internalControlID := factory.NewInternalControl(ownerClient).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteInternalControlMutation,
							map[string]any{
								"input": map[string]any{
									"internalControlId": internalControlID,
								},
							},
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(t, err, tt.forbiddenMsg)
						} else {
							require.NoError(t, err, tt.allowMsg)
						}
					},
				)
			}
		},
	)

	t.Run(
		"read",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name        string
				role        testutil.TestRole
				allowMsg    string
				nodePresent string
			}{
				{
					name:        "owner can read",
					role:        testutil.RoleOwner,
					allowMsg:    "owner should be able to read internal control",
					nodePresent: "owner should receive internal control data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read internal control",
					nodePresent: "admin should receive internal control data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read internal control",
					nodePresent: "viewer should receive internal control data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						internalControlID := factory.NewInternalControl(ownerClient).
							WithName(factory.SafeName("RBAC read " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						var result struct {
							Node *struct {
								ID   string `json:"id"`
								Name string `json:"name"`
							} `json:"node"`
						}

						err := client.Execute(
							readInternalControlQuery,
							map[string]any{
								"id": internalControlID,
							},
							&result,
						)
						require.NoError(t, err, tt.allowMsg)
						require.NotNil(t, result.Node, tt.nodePresent)
					},
				)
			}
		},
	)
}
