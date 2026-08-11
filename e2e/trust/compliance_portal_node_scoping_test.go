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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const subprocessorNodeQuery = `
	query Node($id: ID!) {
		node(id: $id) {
			__typename
			... on Subprocessor {
				id
			}
		}
	}
`

// TestCompliancePortal_SubprocessorNodeScoping verifies that node(id:) only
// hands out third parties the visited portal actually publishes. An organization
// can run several portals, so publishing a subprocessor on one of them must not
// make it readable through a sibling portal, and an unpublished third party must
// not be readable at all.
func TestCompliancePortal_SubprocessorNodeScoping(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	publishingPortalID := factory.CreateCompliancePortal(owner)
	siblingPortalID := factory.CreateCompliancePortal(owner)

	publishingHost := lookupTrustHost(t, owner, publishingPortalID)
	siblingHost := lookupTrustHost(t, owner, siblingPortalID)

	testutil.WaitForCompliancePortalHTTPS(t, publishingHost)
	testutil.WaitForCompliancePortalHTTPS(t, siblingHost)

	publishedID := factory.NewThirdParty(owner).
		WithName(factory.SafeName("Published")).
		WithCategory("CLOUD_PROVIDER").
		Create()
	publishSubprocessor(t, owner, publishingPortalID, publishedID, []string{"US"})

	unpublishedID := factory.NewThirdParty(owner).
		WithName(factory.SafeName("Unpublished")).
		WithCategory("CLOUD_PROVIDER").
		Create()

	t.Run("publishing portal resolves its own subprocessor", func(t *testing.T) {
		t.Parallel()

		var result struct {
			Node struct {
				Typename string `json:"__typename"`
				ID       string `json:"id"`
			} `json:"node"`
		}

		err := owner.ExecuteTrust(publishingHost, subprocessorNodeQuery, map[string]any{
			"id": publishedID,
		}, &result)
		require.NoError(t, err, "the publishing portal must resolve the subprocessor it published")
		assert.Equal(t, "Subprocessor", result.Node.Typename)
		assert.Equal(t, publishedID, result.Node.ID)
	})

	t.Run("unpublished third party is not resolvable", func(t *testing.T) {
		t.Parallel()

		err := owner.ExecuteTrust(publishingHost, subprocessorNodeQuery, map[string]any{
			"id": unpublishedID,
		}, nil)
		require.Error(t, err, "an unpublished third party must not be readable through a portal")
		assert.Contains(
			t,
			err.Error(),
			"not found",
			"an unpublished third party GID must be rejected as not found",
		)
	})

	t.Run("sibling portal cannot resolve another portal's subprocessor", func(t *testing.T) {
		t.Parallel()

		err := owner.ExecuteTrust(siblingHost, subprocessorNodeQuery, map[string]any{
			"id": publishedID,
		}, nil)
		require.Error(t, err, "a portal must not expose a subprocessor published only on a sibling portal")
		assert.Contains(
			t,
			err.Error(),
			"not found",
			"a sibling portal's publication must be rejected as not found",
		)
	})
}
