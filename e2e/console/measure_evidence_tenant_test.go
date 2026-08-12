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

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMeasureEvidence_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1MeasureID := factory.NewMeasure(org1Owner).
		WithName(factory.SafeName("Org1 measure evidence")).
		Create()

	org1Evidence := uploadMeasureEvidence(
		t,
		org1Owner,
		org1MeasureID,
		"org1-evidence.pdf",
	)

	t.Run(
		"cannot upload measure evidence on another organization's measure",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := uploadMeasureEvidenceExpectError(
				t,
				org2,
				org1MeasureID,
				"cross-tenant.pdf",
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not upload measure evidence for another organization's measure",
			)
		},
	)

	t.Run(
		"cannot read evidence from another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			var result struct {
				Node *measureEvidenceWireNode `json:"node"`
			}

			err := org2.Execute(
				`
					query($id: ID!) {
						node(id: $id) {
							... on Evidence { id }
						}
					}
				`,
				map[string]any{"id": org1Evidence.ID},
				&result,
			)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "evidence")
		},
	)

	t.Run(
		"cannot list evidences on another organization's measure",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			var result struct {
				Node *struct {
					ID        string `json:"id"`
					Evidences *struct {
						TotalCount int `json:"totalCount"`
					} `json:"evidences"`
				} `json:"node"`
			}

			err := org2.Execute(
				`
					query($id: ID!) {
						node(id: $id) {
							... on Measure {
								id
								evidences(first: 10) {
									totalCount
								}
							}
						}
					}
				`,
				map[string]any{"id": org1MeasureID},
				&result,
			)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "measure")
		},
	)

	t.Run(
		"cannot delete evidence from another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := deleteMeasureEvidenceExpectError(t, org2, org1Evidence.ID)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete evidence from another organization",
			)
		},
	)
}
