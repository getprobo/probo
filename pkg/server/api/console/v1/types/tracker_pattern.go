// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrackerPatternOrderBy OrderBy[coredata.TrackerPatternOrderField]

	TrackerPatternConnection struct {
		TotalCount int
		Edges      []*TrackerPatternEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *TrackerPatternFilter
	}

	TrackerPatternFilter struct {
		Query       *string
		Source      *coredata.CookieSource
		TrackerType *coredata.TrackerType
	}
)

func NewTrackerPatternConnection(
	p *page.Page[*coredata.TrackerPattern, coredata.TrackerPatternOrderField],
	parentType any,
	parentID gid.GID,
) *TrackerPatternConnection {
	edges := make([]*TrackerPatternEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTrackerPatternEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TrackerPatternConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTrackerPatternConnectionWithFilter(
	p *page.Page[*coredata.TrackerPattern, coredata.TrackerPatternOrderField],
	parentType any,
	parentID gid.GID,
	filter *TrackerPatternFilter,
) *TrackerPatternConnection {
	conn := NewTrackerPatternConnection(p, parentType, parentID)
	conn.Filter = filter
	return conn
}

func NewTrackerPatternEdge(tp *coredata.TrackerPattern, orderBy coredata.TrackerPatternOrderField) *TrackerPatternEdge {
	return &TrackerPatternEdge{
		Cursor: tp.CursorKey(orderBy),
		Node:   NewTrackerPatternNode(tp),
	}
}

func NewTrackerPatternNode(tp *coredata.TrackerPattern) *TrackerPattern {
	return &TrackerPattern{
		ID: tp.ID,
		CookieCategory: &CookieCategory{
			ID: tp.CookieCategoryID,
			CookieBanner: &CookieBanner{
				ID: tp.CookieBannerID,
			},
		},
		TrackerType:   tp.TrackerType,
		Pattern:       tp.Pattern,
		MatchType:     tp.MatchType,
		DisplayName:   tp.DisplayName,
		MaxAgeSeconds: tp.MaxAgeSeconds,
		Description:   tp.Description,
		Source:        tp.Source,
		Excluded:      tp.Excluded,
		LastMatchedAt: tp.LastMatchedAt,
		CreatedAt:     tp.CreatedAt,
		UpdatedAt:     tp.UpdatedAt,
	}
}
