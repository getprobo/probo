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

package types

import (
	"encoding/json"

	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieBannerScanOrderBy OrderBy[coredata.CookieBannerScanOrderField]

	CookieBannerScanConnection struct {
		TotalCount int
		Edges      []*CookieBannerScanEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		State    *coredata.CookieBannerScanState
	}
)

func NewCookieBannerScanConnection(
	p *page.Page[*coredata.CookieBannerScan, coredata.CookieBannerScanOrderField],
	parentType any,
	parentID gid.GID,
	state *coredata.CookieBannerScanState,
) *CookieBannerScanConnection {
	edges := make([]*CookieBannerScanEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCookieBannerScanEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CookieBannerScanConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		State:    state,
	}
}

func NewCookieBannerScanEdge(
	scan *coredata.CookieBannerScan,
	orderBy coredata.CookieBannerScanOrderField,
) *CookieBannerScanEdge {
	return &CookieBannerScanEdge{
		Cursor: scan.CursorKey(orderBy),
		Node:   NewCookieBannerScan(scan),
	}
}

func NewCookieBannerScan(scan *coredata.CookieBannerScan) *CookieBannerScan {
	node := &CookieBannerScan{
		ID: scan.ID,
		CookieBanner: &CookieBanner{
			ID: scan.CookieBannerID,
		},
		State:          scan.State,
		FailureReason:  scan.FailureReason,
		Truncated:      scan.Truncated,
		PagesScanned:   scan.PagesScanned,
		CookiesFound:   scan.CookiesFound,
		StorageFound:   scan.StorageFound,
		ResourcesFound: scan.ResourcesFound,
		StartedAt:      scan.StartedAt,
		CompletedAt:    scan.CompletedAt,
		CreatedAt:      scan.CreatedAt,
		UpdatedAt:      scan.UpdatedAt,
		Checkpoint:     decodeCookieBannerScanCheckpoint(scan.Checkpoint),
	}

	return node
}

func decodeCookieBannerScanCheckpoint(raw *json.RawMessage) *CookieBannerScanCheckpoint {
	decoded, err := cookiebanner.UnmarshalScanCheckpoint(raw)
	if err != nil || decoded == nil {
		return nil
	}

	queue := make([]*CookieBannerScanCheckpointQueueItem, 0, len(decoded.Queue))
	for _, item := range decoded.Queue {
		if item.URL == "" {
			continue
		}

		queue = append(queue, &CookieBannerScanCheckpointQueueItem{URL: item.URL})
	}

	visited := decoded.VisitedTemplates
	if visited == nil {
		visited = []string{}
	}

	hosts := decoded.CoveredHosts
	if hosts == nil {
		hosts = []string{}
	}

	return &CookieBannerScanCheckpoint{
		VisitedTemplates: visited,
		Queue:            queue,
		CoveredHosts:     hosts,
		PagesScanned:     decoded.PagesScanned,
		ElapsedMs:        int(decoded.ElapsedMS),
		Truncated:        decoded.Truncated,
	}
}
