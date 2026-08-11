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

func TestControlApplicability_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	t.Run(
		"update applicability statement",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{name: "owner can update", role: testutil.RoleOwner},
				{name: "admin can update", role: testutil.RoleAdmin},
				{
					name:      "viewer cannot update",
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
						frameworkID := factory.CreateFramework(owner)
						controlID := factory.CreateControl(owner, frameworkID)
						soaID := factory.NewStatementOfApplicability(owner).
							WithName(factory.SafeName("RBAC update SOA " + string(tt.role))).
							Create()
						asID := factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := updateApplicabilityStatementExpectError(
								t,
								client,
								asID,
								false,
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not update applicability statement",
							)

							return
						}

						updated := updateApplicabilityStatementLeaf(
							t,
							client,
							asID,
							false,
							nil,
						)
						require.Equal(t, asID, updated.ID)
						require.False(t, updated.Applicability)
					},
				)
			}
		},
	)

	t.Run(
		"delete applicability statement",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{name: "owner can delete", role: testutil.RoleOwner},
				{name: "admin can delete", role: testutil.RoleAdmin},
				{
					name:      "viewer cannot delete",
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
						frameworkID := factory.CreateFramework(owner)
						controlID := factory.CreateControl(owner, frameworkID)
						soaID := factory.NewStatementOfApplicability(owner).
							WithName(factory.SafeName("RBAC delete AS SOA " + string(tt.role))).
							Create()
						asID := factory.CreateApplicabilityStatement(owner, soaID, controlID, true, nil)

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := deleteApplicabilityStatementExpectError(t, client, asID)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not delete applicability statement",
							)

							return
						}

						deletedID := deleteApplicabilityStatementLeaf(t, client, asID)
						require.Equal(t, asID, deletedID)
					},
				)
			}
		},
	)

	t.Run(
		"control obligation mapping delete",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{name: "owner can delete mapping", role: testutil.RoleOwner},
				{name: "admin can delete mapping", role: testutil.RoleAdmin},
				{
					name:      "viewer cannot delete mapping",
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
						frameworkID := factory.CreateFramework(owner)
						controlID := factory.CreateControl(owner, frameworkID)
						obligationID := createObligationLeaf(
							t,
							owner,
							factory.SafeName("RBAC mapping "+string(tt.role)),
							"LEGAL",
						)
						createControlObligationMappingLeaf(t, owner, controlID, obligationID)

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := deleteControlObligationMappingExpectError(
								t,
								client,
								controlID,
								obligationID,
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not delete control obligation mapping",
							)

							return
						}

						deleteControlObligationMappingLeaf(t, client, controlID, obligationID)
					},
				)
			}
		},
	)
}
