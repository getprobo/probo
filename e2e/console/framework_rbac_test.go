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

func TestFramework_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	const createFrameworkMutation = `
		mutation CreateFramework($input: CreateFrameworkInput!) {
			createFramework(input: $input) {
				frameworkEdge { node { id } }
			}
		}
	`

	const updateFrameworkMutation = `
		mutation UpdateFramework($input: UpdateFrameworkInput!) {
			updateFramework(input: $input) {
				framework { id }
			}
		}
	`

	const deleteFrameworkMutation = `
		mutation DeleteFramework($input: DeleteFrameworkInput!) {
			deleteFramework(input: $input) {
				deletedFrameworkId
			}
		}
	`

	const readFrameworkQuery = `
		query($id: ID!) {
			node(id: $id) {
				... on Framework { id name }
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
					allowMsg: "owner should be able to create framework",
				},
				{
					name:     "admin can create",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to create framework",
				},
				{
					name:         "viewer cannot create",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to create framework",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							createFrameworkMutation,
							map[string]any{
								"input": map[string]any{
									"organizationId": client.GetOrganizationID().String(),
									"name":           factory.SafeName("RBAC create " + string(tt.role)),
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
					allowMsg:   "owner should be able to update framework",
				},
				{
					name:       "admin can update",
					role:       testutil.RoleAdmin,
					updateName: "Updated by Admin",
					allowMsg:   "admin should be able to update framework",
				},
				{
					name:         "viewer cannot update",
					role:         testutil.RoleViewer,
					forbidden:    true,
					updateName:   "Updated by Viewer",
					forbiddenMsg: "viewer should not be able to update framework",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC update " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							updateFrameworkMutation,
							map[string]any{
								"input": map[string]any{
									"id":   frameworkID,
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
					allowMsg: "owner should be able to delete framework",
				},
				{
					name:     "admin can delete",
					role:     testutil.RoleAdmin,
					allowMsg: "admin should be able to delete framework",
				},
				{
					name:         "viewer cannot delete",
					role:         testutil.RoleViewer,
					forbidden:    true,
					forbiddenMsg: "viewer should not be able to delete framework",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						_, err := client.Do(
							deleteFrameworkMutation,
							map[string]any{
								"input": map[string]any{
									"frameworkId": frameworkID,
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
					allowMsg:    "owner should be able to read framework",
					nodePresent: "owner should receive framework data",
				},
				{
					name:        "admin can read",
					role:        testutil.RoleAdmin,
					allowMsg:    "admin should be able to read framework",
					nodePresent: "admin should receive framework data",
				},
				{
					name:        "viewer can read",
					role:        testutil.RoleViewer,
					allowMsg:    "viewer should be able to read framework",
					nodePresent: "viewer should receive framework data",
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						ownerClient := org.Client(t, testutil.RoleOwner)

						frameworkID := factory.NewFramework(ownerClient).
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
							readFrameworkQuery,
							map[string]any{
								"id": frameworkID,
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
