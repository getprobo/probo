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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TrackerPatternOrderBy OrderBy[coredata.TrackerPatternOrderField]

	// TrackerPattern is the Go model bound to the GraphQL TrackerPattern
	// type via @goModel. The first block contains the fields gqlgen
	// fulfills directly from the model; resolver-only fields
	// (cookieCategory, detectedTrackers, thirdParty, commonThirdParty,
	// detectedCount, permission) are populated by the resolver.
	//
	// ThirdPartyID is not exposed in GraphQL — it is a foreign-key
	// handle the resolver uses to load the linked org-scoped third
	// party without re-querying coredata.
	//
	// CommonTrackerPatternID is exposed directly as the
	// commonTrackerPatternId field: a non-null value indicates the
	// pattern is linked to the common tracker-pattern catalog, which is
	// used to debug the provenance of agent-generated descriptions.
	TrackerPattern struct {
		ID            gid.GID                          `json:"id"`
		TrackerType   coredata.TrackerType             `json:"trackerType"`
		Pattern       string                           `json:"pattern"`
		MatchType     coredata.TrackerPatternMatchType `json:"matchType"`
		DisplayName   string                           `json:"displayName"`
		MaxAgeSeconds *int                             `json:"maxAgeSeconds,omitempty"`
		Description   string                           `json:"description"`
		Source        *coredata.CookieSource           `json:"source,omitempty"`
		Excluded      bool                             `json:"excluded"`
		LastMatchedAt *time.Time                       `json:"lastMatchedAt,omitempty"`
		CreatedAt     time.Time                        `json:"createdAt"`
		UpdatedAt     time.Time                        `json:"updatedAt"`

		CookieCategory   *CookieCategory            `json:"cookieCategory,omitempty"`
		DetectedTrackers *DetectedTrackerConnection `json:"detectedTrackers,omitempty"`
		DetectedCount    int                        `json:"detectedCount"`
		Permission       bool                       `json:"permission"`

		ThirdPartyID           *gid.GID `json:"-"`
		CommonTrackerPatternID *gid.GID `json:"commonTrackerPatternId,omitempty"`
	}

	TrackerPatternConnection struct {
		TotalCount int
		Edges      []*TrackerPatternEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *TrackerPatternFilter
	}

	TrackerPatternFilter struct {
		Query            *string
		Source           *coredata.CookieSource
		TrackerType      *coredata.TrackerType
		CookieCategoryID *gid.GID
		ThirdPartyID     *gid.GID
	}
)

func (TrackerPattern) IsNode()          {}
func (t TrackerPattern) GetID() gid.GID { return t.ID }

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
		TrackerType:            tp.TrackerType,
		Pattern:                tp.Pattern,
		MatchType:              tp.MatchType,
		DisplayName:            tp.DisplayName,
		MaxAgeSeconds:          tp.MaxAgeSeconds,
		Description:            tp.Description,
		Source:                 tp.Source,
		Excluded:               tp.Excluded,
		LastMatchedAt:          tp.LastMatchedAt,
		CreatedAt:              tp.CreatedAt,
		UpdatedAt:              tp.UpdatedAt,
		ThirdPartyID:           tp.ThirdPartyID,
		CommonTrackerPatternID: tp.CommonTrackerPatternID,
	}
}
