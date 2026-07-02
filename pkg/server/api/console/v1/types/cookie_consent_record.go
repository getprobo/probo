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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieConsentRecordOrderBy OrderBy[coredata.CookieConsentRecordOrderField]

	CookieConsentRecordConnection struct {
		TotalCount int
		Edges      []*CookieConsentRecordEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
		Filter   *coredata.CookieConsentRecordFilter
	}
)

func NewCookieConsentRecordConnection(
	p *page.Page[*coredata.CookieConsentRecord, coredata.CookieConsentRecordOrderField],
	parentType any,
	parentID gid.GID,
	filter *coredata.CookieConsentRecordFilter,
) *CookieConsentRecordConnection {
	edges := make([]*CookieConsentRecordEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCookieConsentRecordEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &CookieConsentRecordConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
		Filter:   filter,
	}
}

func NewCookieConsentRecordEdge(
	r *coredata.CookieConsentRecord,
	orderBy coredata.CookieConsentRecordOrderField,
) *CookieConsentRecordEdge {
	return &CookieConsentRecordEdge{
		Cursor: r.CursorKey(orderBy),
		Node:   NewCookieConsentRecord(r),
	}
}

func NewCookieConsentRecord(r *coredata.CookieConsentRecord) *CookieConsentRecord {
	return &CookieConsentRecord{
		ID: r.ID,
		CookieBanner: &CookieBanner{
			ID: r.CookieBannerID,
		},
		CookieBannerVersion: &CookieBannerVersion{
			ID: r.CookieBannerVersionID,
		},
		VisitorID:        r.VisitorID,
		IPAddress:        r.IPAddress,
		UserAgent:        r.UserAgent,
		ConsentData:      string(r.ConsentData),
		Action:           r.Action,
		SdkVersion:       r.SdkVersion,
		Regulation:       r.Regulation,
		RegulationSource: r.RegulationSource,
		CountryCode:      r.CountryCode,
		CreatedAt:        r.CreatedAt,
	}
}
