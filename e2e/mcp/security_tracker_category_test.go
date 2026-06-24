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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

// createCookieBannerAndCategory creates a cookie banner and one category
// under it via the MCP API, returning the category id.
func createCookieBannerAndCategory(t *testing.T, mc *testutil.MCPClient, orgID string) string {
	t.Helper()

	var bannerResult struct {
		CookieBanner struct {
			ID string `json:"id"`
		} `json:"cookie_banner"`
	}
	mc.CallToolInto("addCookieBanner", map[string]any{
		"organizationId":    orgID,
		"name":              factory.SafeName("CookieBanner"),
		"origin":            "https://example.com",
		"cookiePolicyUrl":   "https://example.com/cookies",
		"consentExpiryDays": 365,
	}, &bannerResult)
	require.NotEmpty(t, bannerResult.CookieBanner.ID)

	var categoryResult struct {
		CookieCategory struct {
			ID string `json:"id"`
		} `json:"cookie_category"`
	}
	mc.CallToolInto("addCookieCategory", map[string]any{
		"cookieBannerId": bannerResult.CookieBanner.ID,
		"name":           factory.SafeName("Category"),
		"slug":           factory.SafeName("category"),
		"description":    "Test category",
		"rank":           1,
	}, &categoryResult)
	require.NotEmpty(t, categoryResult.CookieCategory.ID)

	return categoryResult.CookieCategory.ID
}

// TestSecurity_MCP_MoveTrackerPatternToCategory_TenantIsolation covers a
// defense-in-depth gap found while auditing GHSA-c74x-79w6-63jh's blast
// radius: MoveTrackerPatternToCategoryTool (MCP) only authorized
// input.TrackerPatternID, never input.TargetCookieCategoryID, unlike the
// equivalent console/v1 GraphQL mutation which authorizes both. Not
// independently exploitable (pkg/cookiebanner.Service.MoveTrackerPatternToCategory
// loads the target category in the caller's own scope and asserts
// pattern.CookieBannerID == target.CookieBannerID), but hardened anyway and
// pinned here.
func TestSecurity_MCP_MoveTrackerPatternToCategory_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	org1MC := testutil.NewMCPClient(t, org1Owner)
	org2MC := testutil.NewMCPClient(t, org2Owner)

	org1CategoryID := createCookieBannerAndCategory(t, org1MC, org1Owner.GetOrganizationID().String())
	org2CategoryID := createCookieBannerAndCategory(t, org2MC, org2Owner.GetOrganizationID().String())

	var patternResult struct {
		TrackerPattern struct {
			ID string `json:"id"`
		} `json:"tracker_pattern"`
	}
	org1MC.CallToolInto("addTrackerPattern", map[string]any{
		"cookieCategoryId": org1CategoryID,
		"pattern":          "org1-tracker",
		"matchType":        "EXACT",
		"displayName":      "Org1 Tracker",
	}, &patternResult)
	require.NotEmpty(t, patternResult.TrackerPattern.ID)

	errText := org1MC.CallToolExpectToolError("moveTrackerPatternToCategory", map[string]any{
		"trackerPatternId":       patternResult.TrackerPattern.ID,
		"targetCookieCategoryId": org2CategoryID,
	})
	require.NotEmpty(t, errText, "must not accept a targetCookieCategoryId belonging to another organization")
}

// TestSecurity_MCP_MoveTrackerResourceToCategory_TenantIsolation is the
// resource-side sibling of the pattern test above.
func TestSecurity_MCP_MoveTrackerResourceToCategory_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)
	org1MC := testutil.NewMCPClient(t, org1Owner)
	org2MC := testutil.NewMCPClient(t, org2Owner)

	org1CategoryID := createCookieBannerAndCategory(t, org1MC, org1Owner.GetOrganizationID().String())
	org2CategoryID := createCookieBannerAndCategory(t, org2MC, org2Owner.GetOrganizationID().String())

	var resourceResult struct {
		TrackerResource struct {
			ID string `json:"id"`
		} `json:"tracker_resource"`
	}
	org1MC.CallToolInto("addTrackerResource", map[string]any{
		"cookieCategoryId": org1CategoryID,
		"url":              "https://org1.example.com/tracker.js",
		"displayName":      "Org1 Resource",
	}, &resourceResult)
	require.NotEmpty(t, resourceResult.TrackerResource.ID)

	errText := org1MC.CallToolExpectToolError("moveTrackerResourceToCategory", map[string]any{
		"trackerResourceId":      resourceResult.TrackerResource.ID,
		"targetCookieCategoryId": org2CategoryID,
	})
	require.NotEmpty(t, errText, "must not accept a targetCookieCategoryId belonging to another organization")
}
