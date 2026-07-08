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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTrustCenter_SubprocessorsFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	trustCenterID := activateTrustCenter(t, owner)

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

		result := querySubprocessors(t, owner, trustCenterID, nil)
		assert.Equal(t, 3, result.CurrentTrustCenter.Subprocessors.TotalCount)
		assert.Len(t, result.CurrentTrustCenter.Subprocessors.Edges, 3)
	})

	t.Run("category filter narrows to one category", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, trustCenterID, map[string]any{
			"category": "CLOUD_PROVIDER",
		})
		require.Equal(t, 1, result.CurrentTrustCenter.Subprocessors.TotalCount)
		require.Len(t, result.CurrentTrustCenter.Subprocessors.Edges, 1)
		assert.Equal(t, awsName, result.CurrentTrustCenter.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("country filter matches array membership", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, trustCenterID, map[string]any{
			"country": "IE",
		})
		require.Equal(t, 1, result.CurrentTrustCenter.Subprocessors.TotalCount)
		require.Len(t, result.CurrentTrustCenter.Subprocessors.Edges, 1)
		assert.Equal(t, stripeName, result.CurrentTrustCenter.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("query filter matches name substring", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, trustCenterID, map[string]any{
			"query": slackName,
		})
		require.Equal(t, 1, result.CurrentTrustCenter.Subprocessors.TotalCount)
		require.Len(t, result.CurrentTrustCenter.Subprocessors.Edges, 1)
		assert.Equal(t, slackName, result.CurrentTrustCenter.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("combined filters intersect", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, trustCenterID, map[string]any{
			"category": "FINANCE",
			"country":  "US",
		})
		require.Equal(t, 1, result.CurrentTrustCenter.Subprocessors.TotalCount)
		require.Len(t, result.CurrentTrustCenter.Subprocessors.Edges, 1)
		assert.Equal(t, stripeName, result.CurrentTrustCenter.Subprocessors.Edges[0].Node.Name)
	})

	t.Run("non-matching filter returns empty set", func(t *testing.T) {
		t.Parallel()

		result := querySubprocessors(t, owner, trustCenterID, map[string]any{
			"category": "SECURITY",
		})
		assert.Equal(t, 0, result.CurrentTrustCenter.Subprocessors.TotalCount)
		assert.Empty(t, result.CurrentTrustCenter.Subprocessors.Edges)
	})
}

type subprocessorsResult struct {
	CurrentTrustCenter struct {
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
	} `json:"currentTrustCenter"`
}

func querySubprocessors(
	t *testing.T,
	owner *testutil.Client,
	trustCenterID string,
	filter map[string]any,
) subprocessorsResult {
	t.Helper()

	const query = `
		query($filter: SubprocessorFilter) {
			currentTrustCenter {
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

	err := owner.ExecuteTrust(trustCenterID, query, map[string]any{"filter": filter}, &result)
	require.NoError(t, err)

	return result
}

func activateTrustCenter(t *testing.T, owner *testutil.Client) string {
	t.Helper()

	const trustCenterQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter { id }
				}
			}
		}
	`

	var lookup struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(trustCenterQuery, map[string]any{
		"organizationId": owner.GetOrganizationID().String(),
	}, &lookup)
	require.NoError(t, err)
	require.NotEmpty(t, lookup.Node.TrustCenter.ID)

	const activateMutation = `
		mutation($input: UpdateTrustCenterInput!) {
			updateTrustCenter(input: $input) {
				trustCenter { id active }
			}
		}
	`

	err = owner.Execute(activateMutation, map[string]any{
		"input": map[string]any{
			"trustCenterId": lookup.Node.TrustCenter.ID,
			"active":        true,
		},
	}, nil)
	require.NoError(t, err)

	return lookup.Node.TrustCenter.ID
}

func publishSubprocessor(t *testing.T, owner *testutil.Client, thirdPartyID string, countries []string) {
	t.Helper()

	const mutation = `
		mutation($input: UpdateThirdPartyInput!) {
			updateThirdParty(input: $input) {
				thirdParty { id showOnTrustCenter countries }
			}
		}
	`

	err := owner.Execute(mutation, map[string]any{
		"input": map[string]any{
			"id":                thirdPartyID,
			"showOnTrustCenter": true,
			"countries":         countries,
		},
	}, nil)
	require.NoError(t, err)
}
