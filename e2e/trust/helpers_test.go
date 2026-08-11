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
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// lookupCompliancePortalID creates a compliance portal for the owner's organization.
func lookupCompliancePortalID(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	return factory.CreateCompliancePortal(owner)
}

func lookupTrustHost(t *testing.T, owner *testutil.Client, compliancePortalID string) string {
	t.Helper()

	testutil.ActivateCompliancePortal(t, owner, compliancePortalID)

	const query = `
		query($compliancePortalId: ID!) {
			node(id: $compliancePortalId) {
				... on CompliancePortal {
					publicUrl
				}
			}
		}
	`

	var result struct {
		Node struct {
			PublicURL string `json:"publicUrl"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"compliancePortalId": compliancePortalID,
	}, &result)
	require.NoError(t, err)
	require.NotEmpty(t, result.Node.PublicURL)

	publicURL, err := url.Parse(result.Node.PublicURL)
	require.NoError(t, err)
	require.NotEmpty(t, publicURL.Host)

	return publicURL.Host
}

// activateCompliancePortal flips the compliance portal to active so its public surface
// (NDA, subprocessors, reports, branding) becomes reachable by visitors.
func activateCompliancePortal(t *testing.T, owner *testutil.Client, compliancePortalID string) {
	t.Helper()

	testutil.ActivateCompliancePortal(t, owner, compliancePortalID)
}
