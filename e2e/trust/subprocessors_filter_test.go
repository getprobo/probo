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

func TestCompliancePortal_SubprocessorsFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := lookupCompliancePortalID(t, owner)
	activateCompliancePortal(t, owner, compliancePortalID)

	awsName := factory.SafeName("AWS")
	awsID := factory.NewThirdParty(owner).WithName(awsName).WithCategory("CLOUD_PROVIDER").Create()
	publishSubprocessor(t, owner, awsID, []string{"US"})

	stripeName := factory.SafeName("Stripe")
	stripeID := factory.NewThirdParty(owner).WithName(stripeName).WithCategory("FINANCE").Create()
	publishSubprocessor(t, owner, stripeID, []string{"US", "IE"})

	slackName := factory.SafeName("Slack")
	slackID := factory.NewThirdParty(owner).WithName(slackName).WithCategory("COLLABORATION").Create()
	publishSubprocessor(t, owner, slackID, []string{"US"})

	t.Run("no filter returns every published subprocessor", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, nil)
		assert.Equal(t, 3, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		assert.Len(t, result.CurrentCompliancePortal.Subprocessors.Edges, 3)
	})

	t.Run("category filter narrows to one category", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, map[string]any{
			"category": "CLOUD_PROVIDER",
		})
		require.Equal(t, 1, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		require.Len(t, result.CurrentCompliancePortal.Subprocessors.Edges, 1)
		assert.Equal(t, awsName, result.CurrentCompliancePortal.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("country filter matches array membership", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, map[string]any{
			"country": "IE",
		})
		require.Equal(t, 1, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		require.Len(t, result.CurrentCompliancePortal.Subprocessors.Edges, 1)
		assert.Equal(t, stripeName, result.CurrentCompliancePortal.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("query filter matches name substring", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, map[string]any{
			"query": slackName,
		})
		require.Equal(t, 1, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		require.Len(t, result.CurrentCompliancePortal.Subprocessors.Edges, 1)
		assert.Equal(t, slackName, result.CurrentCompliancePortal.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("combined filters intersect", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, map[string]any{
			"category": "FINANCE",
			"country":  "US",
		})
		require.Equal(t, 1, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		require.Len(t, result.CurrentCompliancePortal.Subprocessors.Edges, 1)
		assert.Equal(t, stripeName, result.CurrentCompliancePortal.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("non-matching filter returns empty set", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, compliancePortalID, map[string]any{
			"category": "SECURITY",
		})
		assert.Equal(t, 0, result.CurrentCompliancePortal.Subprocessors.TotalCount)
		assert.Empty(t, result.CurrentCompliancePortal.Subprocessors.Edges)
	})
}

type subprocessorsResult struct {
	CurrentCompliancePortal struct {
		Subprocessors struct {
			TotalCount int `json:"totalCount"`
			Edges      []struct {
				Node struct {
					ID        string   `json:"id"`
					Name      string   `json:"name"`
					Category  string   `json:"category"`
					Countries []string `json:"countries"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"subprocessors"`
	} `json:"currentCompliancePortal"`
}

func querySubprocessors(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	filter map[string]any,
) subprocessorsResult {
	t.Helper()

	const query = `
		query($filter: SubprocessorFilter) {
			currentCompliancePortal {
				subprocessors(first: 50, filter: $filter) {
					totalCount
					edges {
						node {
							id
							name
							category
							countries
						}
					}
				}
			}
		}
	`

	var result subprocessorsResult

	err := owner.ExecuteTrust(compliancePortalID, query, map[string]any{"filter": filter}, &result)
	require.NoError(t, err)

	return result
}

func publishSubprocessor(t *testing.T, owner *testutil.Client, thirdPartyID string, countries []string) {
	t.Helper()

	const mutation = `
		mutation($input: UpdateThirdPartyInput!) {
			updateThirdParty(input: $input) {
				thirdParty { id showOnCompliancePortal countries }
			}
		}
	`

	err := owner.Execute(mutation, map[string]any{
		"input": map[string]any{
			"id":                     thirdPartyID,
			"showOnCompliancePortal": true,
			"countries":              countries,
		},
	}, nil)
	require.NoError(t, err)
}
