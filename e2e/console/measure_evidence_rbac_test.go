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

func TestMeasureEvidence_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	t.Run(
		"upload",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{
					name: "owner can upload measure evidence",
					role: testutil.RoleOwner,
				},
				{
					name: "admin can upload measure evidence",
					role: testutil.RoleAdmin,
				},
				{
					name:      "viewer cannot upload measure evidence",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						measureID := factory.NewMeasure(owner).
							WithName(factory.SafeName("RBAC upload " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := uploadMeasureEvidenceExpectError(
								t,
								client,
								measureID,
								"rbac-upload.pdf",
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not upload measure evidence",
							)

							return
						}

						uploaded := uploadMeasureEvidence(
							t,
							client,
							measureID,
							"rbac-upload.pdf",
						)
						assert.NotEmpty(t, uploaded.ID)
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
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{
					name: "owner can delete measure evidence",
					role: testutil.RoleOwner,
				},
				{
					name: "admin can delete measure evidence",
					role: testutil.RoleAdmin,
				},
				{
					name:      "viewer cannot delete measure evidence",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						measureID := factory.NewMeasure(owner).
							WithName(factory.SafeName("RBAC delete " + string(tt.role))).
							Create()

						evidenceID := uploadMeasureEvidence(
							t,
							owner,
							measureID,
							"rbac-delete.pdf",
						).ID

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := deleteMeasureEvidenceExpectError(t, client, evidenceID)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not delete measure evidence",
							)

							return
						}

						deletedID := deleteMeasureEvidence(t, client, evidenceID)
						assert.Equal(t, evidenceID, deletedID)
					},
				)
			}
		},
	)

	t.Run(
		"viewer read and list",
		func(t *testing.T) {
			t.Parallel()

			owner := org.Client(t, testutil.RoleOwner)
			viewer := org.Client(t, testutil.RoleViewer)

			measureID := factory.NewMeasure(owner).
				WithName(factory.SafeName("RBAC viewer read")).
				Create()

			uploaded := uploadMeasureEvidence(t, owner, measureID, "viewer-read.pdf")

			node := queryEvidenceNode(t, viewer, uploaded.ID)
			require.NotNil(t, node)
			assert.Equal(t, uploaded.ID, node.ID)
			assert.False(t, node.CanDelete)
			require.NotNil(t, node.File)
			assert.NotEmpty(t, node.File.DownloadURL)

			list := queryMeasureEvidences(t, viewer, measureID)
			assert.Equal(t, 1, list.TotalCount)
			require.Len(t, list.Edges, 1)
			assert.Equal(t, uploaded.ID, list.Edges[0].Node.ID)
		},
	)
}
