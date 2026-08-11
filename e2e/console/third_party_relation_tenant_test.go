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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyRelation_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	parentID := factory.NewThirdParty(org1Owner).WithName("Org1 Parent").Create()
	createChildThirdParty(t, org1Owner, parentID, "Org1 Child")

	t.Run("cannot list children of other org third party", func(t *testing.T) {
		t.Parallel()

		const query = `
			query($id: ID!) {
				node(id: $id) {
					... on ThirdParty {
						childThirdParties(first: 10) {
							totalCount
						}
					}
				}
			}
		`

		var result struct {
			Node *struct {
				ChildThirdParties *struct {
					TotalCount int `json:"totalCount"`
				} `json:"childThirdParties"`
			} `json:"node"`
		}

		err := org2Owner.Execute(query, map[string]any{"id": parentID}, &result)

		nodeInaccessible := err != nil || result.Node == nil || result.Node.ChildThirdParties == nil
		emptyResult := result.Node != nil && result.Node.ChildThirdParties != nil && result.Node.ChildThirdParties.TotalCount == 0
		assert.True(t, nodeInaccessible || emptyResult, "expected either inaccessible node or zero children")
	})
}
