// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
