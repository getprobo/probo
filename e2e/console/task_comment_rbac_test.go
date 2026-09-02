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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTaskComment_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)
	owner := org.Client(t, testutil.RoleOwner)
	taskID := factory.NewTaskWithoutMeasure(owner).Create()

	const updateMutation = `
		mutation UpdateTaskComment($input: UpdateTaskCommentInput!) {
			updateTaskComment(input: $input) {
				taskComment { id }
			}
		}
	`

	const deleteMutation = `
		mutation DeleteTaskComment($input: DeleteTaskCommentInput!) {
			deleteTaskComment(input: $input) {
				deletedTaskCommentId
			}
		}
	`

	t.Run(
		"update",
		func(t *testing.T) {
			t.Parallel()

			t.Run(
				"author can update own comment",
				func(t *testing.T) {
					t.Parallel()

					tests := []struct {
						name       string
						authorRole testutil.TestRole
					}{
						{name: "owner", authorRole: testutil.RoleOwner},
						{name: "admin", authorRole: testutil.RoleAdmin},
						{name: "viewer", authorRole: testutil.RoleViewer},
					}

					for _, tt := range tests {
						t.Run(
							tt.name,
							func(t *testing.T) {
								t.Parallel()

								owner := org.Client(t, testutil.RoleOwner)
								author := org.Client(t, tt.authorRole)
								commentID := factory.NewTaskComment(owner, taskID).
									WithContent("Own comment").
									WithOwnerID(author.GetProfileID().String()).
									Create()

								_, err := author.Do(
									updateMutation,
									map[string]any{
										"input": map[string]any{
											"taskCommentId": commentID,
											"content":       factory.ProseMirrorPlainText("Edited by author"),
										},
									},
								)
								require.NoError(t, err, "%s should update their own comment", tt.name)
							},
						)
					}
				},
			)

			t.Run(
				"nobody can update another user's comment",
				func(t *testing.T) {
					t.Parallel()

					tests := []struct {
						name       string
						actorRole  testutil.TestRole
						authorRole testutil.TestRole
					}{
						{
							name:       "admin cannot update owner comment",
							actorRole:  testutil.RoleAdmin,
							authorRole: testutil.RoleOwner,
						},
						{
							name:       "owner cannot update admin comment",
							actorRole:  testutil.RoleOwner,
							authorRole: testutil.RoleAdmin,
						},
						{
							name:       "viewer cannot update owner comment",
							actorRole:  testutil.RoleViewer,
							authorRole: testutil.RoleOwner,
						},
					}

					for _, tt := range tests {
						t.Run(
							tt.name,
							func(t *testing.T) {
								t.Parallel()

								owner := org.Client(t, testutil.RoleOwner)
								author := org.Client(t, tt.authorRole)
								actor := org.Client(t, tt.actorRole)
								commentID := factory.NewTaskComment(owner, taskID).
									WithContent("Someone else's comment").
									WithOwnerID(author.GetProfileID().String()).
									Create()

								_, err := actor.Do(
									updateMutation,
									map[string]any{
										"input": map[string]any{
											"taskCommentId": commentID,
											"content":       factory.ProseMirrorPlainText("Hijacked comment"),
										},
									},
								)
								testutil.RequireForbiddenError(t, err, tt.name)
							},
						)
					}
				},
			)
		},
	)

	t.Run(
		"delete",
		func(t *testing.T) {
			t.Parallel()

			t.Run(
				"author can delete own comment",
				func(t *testing.T) {
					t.Parallel()

					tests := []struct {
						name       string
						authorRole testutil.TestRole
					}{
						{name: "owner", authorRole: testutil.RoleOwner},
						{name: "admin", authorRole: testutil.RoleAdmin},
						{name: "viewer", authorRole: testutil.RoleViewer},
					}

					for _, tt := range tests {
						t.Run(
							tt.name,
							func(t *testing.T) {
								t.Parallel()

								owner := org.Client(t, testutil.RoleOwner)
								author := org.Client(t, tt.authorRole)
								commentID := factory.NewTaskComment(owner, taskID).
									WithContent("Deletable own comment").
									WithOwnerID(author.GetProfileID().String()).
									Create()

								_, err := author.Do(
									deleteMutation,
									map[string]any{
										"input": map[string]any{
											"taskCommentId": commentID,
										},
									},
								)
								require.NoError(t, err, "%s should delete their own comment", tt.name)
							},
						)
					}
				},
			)

			t.Run(
				"owner and admin can delete another user's comment",
				func(t *testing.T) {
					t.Parallel()

					tests := []struct {
						name      string
						actorRole testutil.TestRole
					}{
						{name: "owner", actorRole: testutil.RoleOwner},
						{name: "admin", actorRole: testutil.RoleAdmin},
					}

					for _, tt := range tests {
						t.Run(
							tt.name,
							func(t *testing.T) {
								t.Parallel()

								owner := org.Client(t, testutil.RoleOwner)
								viewer := org.Client(t, testutil.RoleViewer)
								actor := org.Client(t, tt.actorRole)
								commentID := factory.NewTaskComment(owner, taskID).
									WithContent("Comment to moderate").
									WithOwnerID(viewer.GetProfileID().String()).
									Create()

								_, err := actor.Do(
									deleteMutation,
									map[string]any{
										"input": map[string]any{
											"taskCommentId": commentID,
										},
									},
								)
								require.NoError(t, err, "%s should delete another user's comment", tt.name)
							},
						)
					}
				},
			)

			t.Run(
				"viewer cannot delete another user's comment",
				func(t *testing.T) {
					t.Parallel()

					owner := org.Client(t, testutil.RoleOwner)
					viewer := org.Client(t, testutil.RoleViewer)
					commentID := factory.NewTaskComment(owner, taskID).
						WithContent("Owner comment").
						Create()

					_, err := viewer.Do(
						deleteMutation,
						map[string]any{
							"input": map[string]any{
								"taskCommentId": commentID,
							},
						},
					)
					testutil.RequireForbiddenError(t, err, "viewer cannot delete another user's comment")
				},
			)
		},
	)

	t.Run(
		"permission field",
		func(t *testing.T) {
			t.Parallel()

			const permissionQuery = `
				query($id: ID!) {
					node(id: $id) {
						... on TaskComment {
							canUpdate: permission(action: "core:task-comment:update")
							canDelete: permission(action: "core:task-comment:delete")
						}
					}
				}
			`

			type permissionResult struct {
				Node struct {
					CanUpdate bool `json:"canUpdate"`
					CanDelete bool `json:"canDelete"`
				} `json:"node"`
			}

			owner := org.Client(t, testutil.RoleOwner)
			admin := org.Client(t, testutil.RoleAdmin)
			viewer := org.Client(t, testutil.RoleViewer)
			ownerCommentID := factory.NewTaskComment(owner, taskID).
				WithContent("Owner permission comment").
				Create()
			viewerCommentID := factory.NewTaskComment(owner, taskID).
				WithContent("Viewer permission comment").
				WithOwnerID(viewer.GetProfileID().String()).
				Create()

			t.Run(
				"author can update and delete",
				func(t *testing.T) {
					t.Parallel()

					var result permissionResult

					err := owner.ForTest(t).Execute(
						permissionQuery,
						map[string]any{"id": ownerCommentID},
						&result,
					)
					require.NoError(t, err)
					assert.True(t, result.Node.CanUpdate)
					assert.True(t, result.Node.CanDelete)
				},
			)

			t.Run(
				"admin cannot update another user's comment but can delete it",
				func(t *testing.T) {
					t.Parallel()

					var result permissionResult

					err := admin.ForTest(t).Execute(
						permissionQuery,
						map[string]any{"id": ownerCommentID},
						&result,
					)
					require.NoError(t, err)
					assert.False(t, result.Node.CanUpdate)
					assert.True(t, result.Node.CanDelete)
				},
			)

			t.Run(
				"viewer can update and delete own comment",
				func(t *testing.T) {
					t.Parallel()

					var result permissionResult

					err := viewer.ForTest(t).Execute(
						permissionQuery,
						map[string]any{"id": viewerCommentID},
						&result,
					)
					require.NoError(t, err)
					assert.True(t, result.Node.CanUpdate)
					assert.True(t, result.Node.CanDelete)
				},
			)

			t.Run(
				"viewer cannot update or delete another user's comment",
				func(t *testing.T) {
					t.Parallel()

					var result permissionResult

					err := viewer.ForTest(t).Execute(
						permissionQuery,
						map[string]any{"id": ownerCommentID},
						&result,
					)
					require.NoError(t, err)
					assert.False(t, result.Node.CanUpdate)
					assert.False(t, result.Node.CanDelete)
				},
			)
		},
	)
}
