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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestCookieBannerScan_ListAndRequest(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	bannerID := factory.CreateCookieBanner(owner)

	completedID := seedCompletedScan(t, bannerID, time.Now().Add(-2*time.Hour))

	const requestQuery = `
		mutation RequestCookieBannerScan($input: RequestCookieBannerScanInput!) {
			requestCookieBannerScan(input: $input) {
				cookieBannerScan {
					id
					state
					startedAt
					completedAt
				}
				cookieBanner {
					id
					activeScan { id state }
				}
			}
		}
	`

	var requested struct {
		RequestCookieBannerScan struct {
			CookieBannerScan struct {
				ID          string  `json:"id"`
				State       string  `json:"state"`
				StartedAt   *string `json:"startedAt"`
				CompletedAt *string `json:"completedAt"`
			} `json:"cookieBannerScan"`
			CookieBanner struct {
				ID         string `json:"id"`
				ActiveScan *struct {
					ID    string `json:"id"`
					State string `json:"state"`
				} `json:"activeScan"`
			} `json:"cookieBanner"`
		} `json:"requestCookieBannerScan"`
	}

	err := owner.Execute(requestQuery, map[string]any{
		"input": map[string]any{"cookieBannerId": bannerID},
	}, &requested)
	require.NoError(t, err)

	scan := requested.RequestCookieBannerScan.CookieBannerScan
	assert.NotEmpty(t, scan.ID)
	assert.Equal(t, "PENDING", scan.State)
	assert.Nil(t, scan.StartedAt)
	assert.Nil(t, scan.CompletedAt)
	require.NotNil(t, requested.RequestCookieBannerScan.CookieBanner.ActiveScan)
	assert.Equal(t, scan.ID, requested.RequestCookieBannerScan.CookieBanner.ActiveScan.ID)
	assert.Equal(t, "PENDING", requested.RequestCookieBannerScan.CookieBanner.ActiveScan.State)

	err = owner.Execute(requestQuery, map[string]any{
		"input": map[string]any{"cookieBannerId": bannerID},
	}, &struct{}{})
	testutil.RequireErrorCode(t, err, "CONFLICT")

	const listQuery = `
		query ListCookieBannerScans($id: ID!, $filter: CookieBannerScanFilter) {
			node(id: $id) {
				... on CookieBanner {
					scans(first: 10, filter: $filter) {
						edges {
							node {
								id
								state
								truncated
								pagesScanned
								cookiesFound
								storageFound
								resourcesFound
								checkpoint {
									visitedTemplates
									queue { url }
									coveredHosts
									pagesScanned
									elapsedMs
									truncated
								}
							}
						}
					}
				}
			}
		}
	`

	type scanNode struct {
		ID             string `json:"id"`
		State          string `json:"state"`
		Truncated      bool   `json:"truncated"`
		PagesScanned   int    `json:"pagesScanned"`
		CookiesFound   int    `json:"cookiesFound"`
		StorageFound   int    `json:"storageFound"`
		ResourcesFound int    `json:"resourcesFound"`
		Checkpoint     *struct {
			VisitedTemplates []string `json:"visitedTemplates"`
			Queue            []struct {
				URL string `json:"url"`
			} `json:"queue"`
			CoveredHosts []string `json:"coveredHosts"`
			PagesScanned int      `json:"pagesScanned"`
			ElapsedMs    int      `json:"elapsedMs"`
			Truncated    bool     `json:"truncated"`
		} `json:"checkpoint"`
	}

	var listed struct {
		Node struct {
			Scans struct {
				Edges []struct {
					Node scanNode `json:"node"`
				} `json:"edges"`
			} `json:"scans"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(listQuery, map[string]any{"id": bannerID}, &listed))
	require.Len(t, listed.Node.Scans.Edges, 2)
	assert.Equal(t, scan.ID, listed.Node.Scans.Edges[0].Node.ID)
	assert.Equal(t, "PENDING", listed.Node.Scans.Edges[0].Node.State)
	assert.Equal(t, completedID, listed.Node.Scans.Edges[1].Node.ID)

	completed := listed.Node.Scans.Edges[1].Node
	assert.Equal(t, "COMPLETED", completed.State)
	assert.True(t, completed.Truncated)
	assert.Equal(t, 3, completed.PagesScanned)
	assert.Equal(t, 2, completed.CookiesFound)
	assert.Equal(t, 1, completed.StorageFound)
	assert.Equal(t, 4, completed.ResourcesFound)
	require.NotNil(t, completed.Checkpoint)
	assert.Equal(t, []string{"https://example.com/"}, completed.Checkpoint.VisitedTemplates)
	require.Len(t, completed.Checkpoint.Queue, 1)
	assert.Equal(t, "https://example.com/a", completed.Checkpoint.Queue[0].URL)
	assert.Equal(t, []string{"example.com"}, completed.Checkpoint.CoveredHosts)
	assert.Equal(t, 2, completed.Checkpoint.PagesScanned)
	assert.Equal(t, 1500, completed.Checkpoint.ElapsedMs)
	assert.True(t, completed.Checkpoint.Truncated)

	var filtered struct {
		Node struct {
			Scans struct {
				Edges []struct {
					Node struct {
						ID    string `json:"id"`
						State string `json:"state"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"scans"`
		} `json:"node"`
	}

	require.NoError(t, owner.Execute(listQuery, map[string]any{
		"id":     bannerID,
		"filter": map[string]any{"state": "COMPLETED"},
	}, &filtered))
	require.Len(t, filtered.Node.Scans.Edges, 1)
	assert.Equal(t, completedID, filtered.Node.Scans.Edges[0].Node.ID)
	assert.Equal(t, "COMPLETED", filtered.Node.Scans.Edges[0].Node.State)
}

func seedCompletedScan(t *testing.T, bannerID string, createdAt time.Time) string {
	t.Helper()

	bannerGID, err := gid.ParseGID(bannerID)
	require.NoError(t, err)

	ctx := context.Background()
	conn := dialTestPg(t, ctx)
	t.Cleanup(func() { _ = conn.Close(ctx) })

	var organizationID string
	err = conn.QueryRow(ctx, `SELECT organization_id FROM cookie_banners WHERE id = $1`, bannerID).Scan(&organizationID)
	require.NoError(t, err)

	id := gid.New(bannerGID.TenantID(), coredata.CookieBannerScanEntityType)
	startedAt := createdAt.Add(time.Minute)
	completedAt := createdAt.Add(2 * time.Minute)
	checkpoint := `{
		"visited_templates": ["https://example.com/"],
		"queue": [{"url": "https://example.com/a", "depth": 1}],
		"covered_hosts": ["example.com"],
		"pages_scanned": 2,
		"elapsed_ms": 1500,
		"truncated": true
	}`

	_, err = conn.Exec(
		ctx,
		`
		INSERT INTO cookie_banner_scans (
			id, tenant_id, organization_id, cookie_banner_id,
			state, failure_reason, truncated, checkpoint,
			pages_scanned, cookies_found, storage_found, resources_found,
			started_at, completed_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			'COMPLETED', NULL, TRUE, $5::jsonb,
			3, 2, 1, 4,
			$6, $7, $8, $9
		)
	`,
		id,
		bannerGID.TenantID(),
		organizationID,
		bannerID,
		checkpoint,
		startedAt,
		completedAt,
		createdAt,
		completedAt,
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cleanupConn := dialTestPg(t, cleanupCtx)
		defer func() { _ = cleanupConn.Close(cleanupCtx) }()

		_, err := cleanupConn.Exec(cleanupCtx, `DELETE FROM cookie_banner_scans WHERE id = $1`, id)
		assert.NoError(t, err)
	})

	return id.String()
}
