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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// A CSV without the email column fails every fetch until it is fixed.
const brokenCsvData = "full_name,role\nJane Smith,admin"

const retryFetchMutation = `
	mutation($input: RetryAccessReviewCampaignSourceFetchInput!) {
		retryAccessReviewCampaignSourceFetch(input: $input) {
			accessReviewCampaign { id status }
		}
	}
`

func TestAccessReviewCampaign_RetryFailedSourceFetch(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	orgID := owner.GetOrganizationID().String()

	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName("Broken CSV Source").
		WithCsvData(brokenCsvData).
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("Retry Fetch Campaign").
		WithAccessReviewSourceIDs([]string{sourceID}).
		Create()

	err := owner.Execute(`
		mutation($input: StartAccessReviewCampaignInput!) {
			startAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{"accessReviewCampaignId": campaignID},
	}, nil)
	require.NoError(t, err)

	waitForLatestFetch(t, owner, campaignID, "FAILED", 1)

	_, err = owner.Do(`
		mutation($input: CloseAccessReviewCampaignInput!) {
			closeAccessReviewCampaign(input: $input) {
				accessReviewCampaign { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{"accessReviewCampaignId": campaignID},
	})
	testutil.RequireErrorCode(t, err, "INVALID", "a campaign with a failed source must not close")

	retryInput := map[string]any{
		"input": map[string]any{
			"accessReviewCampaignId": campaignID,
			"accessReviewSourceId":   sourceID,
		},
	}

	_, err = viewer.Do(retryFetchMutation, retryInput)
	testutil.RequireForbiddenError(t, err, "a viewer must not retry a fetch")

	stranger := testutil.NewClient(t, testutil.RoleOwner)
	_, err = stranger.Do(retryFetchMutation, retryInput)
	testutil.RequireForbiddenError(t, err, "another organization must not retry a fetch")

	err = owner.Execute(`
		mutation($input: UpdateAccessReviewSourceInput!) {
			updateAccessReviewSource(input: $input) {
				accessReviewSource { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"accessReviewSourceId": sourceID,
			"csvData":              testCsvData,
		},
	}, nil)
	require.NoError(t, err)

	var retryResult struct {
		RetryAccessReviewCampaignSourceFetch struct {
			AccessReviewCampaign struct {
				Status string `json:"status"`
			} `json:"accessReviewCampaign"`
		} `json:"retryAccessReviewCampaignSourceFetch"`
	}

	err = admin.Execute(retryFetchMutation, retryInput, &retryResult)
	require.NoError(t, err)
	assert.Equal(t, "IN_PROGRESS", retryResult.RetryAccessReviewCampaignSourceFetch.AccessReviewCampaign.Status)

	waitForLatestFetch(t, owner, campaignID, "SUCCESS", 2)

	_, err = owner.Do(retryFetchMutation, retryInput)
	testutil.RequireErrorCode(t, err, "INVALID", "a fetch that succeeded must not be retried")
}

func waitForLatestFetch(
	t *testing.T,
	client *testutil.Client,
	campaignID string,
	wantStatus string,
	wantAttempt int,
) {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on AccessReviewCampaign {
					status
					sources {
						fetchAttempts(first: 1) {
							edges { node { status attemptNumber } }
						}
					}
				}
			}
		}
	`

	require.Eventually(t, func() bool {
		var r struct {
			Node struct {
				Status  string `json:"status"`
				Sources []struct {
					FetchAttempts struct {
						Edges []struct {
							Node struct {
								Status        string `json:"status"`
								AttemptNumber int    `json:"attemptNumber"`
							} `json:"node"`
						} `json:"edges"`
					} `json:"fetchAttempts"`
				} `json:"sources"`
			} `json:"node"`
		}
		if err := client.Execute(query, map[string]any{"id": campaignID}, &r); err != nil {
			return false
		}

		if r.Node.Status != "PENDING_ACTIONS" || len(r.Node.Sources) != 1 {
			return false
		}

		edges := r.Node.Sources[0].FetchAttempts.Edges

		return len(edges) == 1 &&
			edges[0].Node.Status == wantStatus &&
			edges[0].Node.AttemptNumber == wantAttempt
	}, 60*time.Second, 1*time.Second, "campaign should settle with a %s fetch attempt %d", wantStatus, wantAttempt)
}
