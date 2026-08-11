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

func TestStatementOfApplicability_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	soaID := factory.NewStatementOfApplicability(org1Owner).Create()

	t.Run(
		"cannot create document for another org SOA",
		func(t *testing.T) {
			t.Parallel()

			const query = `
				mutation($input: PublishStatementOfApplicabilityInput!) {
					publishStatementOfApplicability(input: $input) {
						documentEdge {
							node { id }
						}
						documentVersionEdge {
							node { id }
						}
					}
				}
			`

			err := org2Owner.ExecuteShouldFail(
				query,
				map[string]any{
					"input": map[string]any{
						"minor":                      false,
						"statementOfApplicabilityId": soaID,
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not publish another organization's statement of applicability",
			)
		},
	)

	t.Run(
		"cannot create applicability statement referencing a control from another organization",
		func(t *testing.T) {
			t.Parallel()

			org2FrameworkID := factory.CreateFramework(org2Owner)
			org2ControlID := factory.CreateControl(org2Owner, org2FrameworkID)

			_, err := org1Owner.Do(`
				mutation($input: CreateApplicabilityStatementInput!) {
					createApplicabilityStatement(input: $input) {
						applicabilityStatementEdge { node { id } }
					}
				}
			`, map[string]any{
				"input": map[string]any{
					"statementOfApplicabilityId": soaID,
					"controlId":                  org2ControlID,
					"applicability":              true,
				},
			})
			testutil.RequireForbiddenError(
				t,
				err,
				"must not create applicability statement for another organization's control",
			)
		},
	)
}
