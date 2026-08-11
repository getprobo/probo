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
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestCompliancePortal_CreateAssignsSlugAutomatically(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	const entityName = "Acme Trust"

	portalID := factory.CreateCompliancePortal(owner, factory.Attrs{
		"entityName": entityName,
	})

	const query = `
		query($compliancePortalId: ID!) {
			node(id: $compliancePortalId) {
				... on CompliancePortal {
					slug
					entityName
				}
			}
		}
	`

	var result struct {
		Node struct {
			Slug       string `json:"slug"`
			EntityName string `json:"entityName"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{
		"compliancePortalId": portalID,
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, entityName, result.Node.EntityName)
	assert.Regexp(t, regexp.MustCompile(`^acme-trust-[0-9a-f]{8}$`), result.Node.Slug)
}
