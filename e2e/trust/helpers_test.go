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

package trust_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// lookupTrustCenterID resolves the trust center ID of the owner's organization.
func lookupTrustCenterID(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	const query = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter { id }
				}
			}
		}
	`

	var result struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
	}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.TrustCenter.ID)

	return result.Node.TrustCenter.ID
}

// activateTrustCenter flips the trust center to active so its public surface
// (NDA, subprocessors, reports, branding) becomes reachable by visitors.
func activateTrustCenter(t *testing.T, owner *testutil.Client, trustCenterID string) {
	t.Helper()

	const query = `
		mutation($input: UpdateTrustCenterInput!) {
			updateTrustCenter(input: $input) {
				trustCenter { id active }
			}
		}
	`

	err := owner.Execute(query, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"active":        true,
		},
	}, nil)
	require.NoError(t, err)
}
