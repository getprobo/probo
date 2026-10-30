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

package mcp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_AccessReviewCampaign_RetryFailedSourceFetch(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// A CSV without the email column fails every fetch.
	sourceID := factory.NewAccessReviewSource(owner, orgID).
		WithName("Broken CSV Source").
		WithCsvData("full_name,role\nJane Smith,admin").
		Create()

	campaignID := factory.NewAccessReviewCampaign(owner, orgID).
		WithName("MCP Retry Fetch Campaign").
		WithAccessReviewSourceIDs([]string{sourceID}).
		Create()

	var started struct {
		Campaign struct {
			Status string `json:"status"`
		} `json:"campaign"`
	}
	mc.CallToolInto("startAccessReviewCampaign", map[string]any{
		"campaign_id": campaignID,
	}, &started)
	require.Equal(t, "IN_PROGRESS", started.Campaign.Status)

	waitForCampaignStatus(t, owner, campaignID, "PENDING_ACTIONS")

	closeErr := mc.CallToolExpectToolError("closeAccessReviewCampaign", map[string]any{
		"campaign_id": campaignID,
	})
	assert.Contains(t, closeErr, "Broken CSV Source")
	assert.Contains(t, closeErr, sourceID)

	var retried struct {
		Campaign struct {
			Status string `json:"status"`
		} `json:"campaign"`
	}
	mc.CallToolInto("retryAccessReviewCampaignSourceFetch", map[string]any{
		"campaign_id":             campaignID,
		"access_review_source_id": sourceID,
	}, &retried)
	assert.Equal(t, "IN_PROGRESS", retried.Campaign.Status)

	waitForCampaignStatus(t, owner, campaignID, "PENDING_ACTIONS")

	var cancelled struct {
		Campaign struct {
			Status string `json:"status"`
		} `json:"campaign"`
	}
	mc.CallToolInto("cancelAccessReviewCampaign", map[string]any{
		"campaign_id": campaignID,
	}, &cancelled)
	assert.Equal(t, "CANCELLED", cancelled.Campaign.Status)
}

func waitForCampaignStatus(t *testing.T, owner *testutil.Client, campaignID string, want string) {
	t.Helper()

	require.Eventually(t, func() bool {
		var r struct {
			Node struct {
				Status string `json:"status"`
			} `json:"node"`
		}

		err := owner.Execute(
			`query($id: ID!) { node(id: $id) { ... on AccessReviewCampaign { status } } }`,
			map[string]any{"id": campaignID},
			&r,
		)

		return err == nil && r.Node.Status == want
	}, 60*time.Second, 1*time.Second, "campaign should reach %s", want)
}
