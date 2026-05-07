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
		VisitorID:   r.VisitorID,
		IPAddress:   r.IPAddress,
		UserAgent:   r.UserAgent,
		ConsentData: string(r.ConsentData),
		Action:      r.Action,
		SdkVersion:  r.SdkVersion,
		Regulation:  r.Regulation.String(),
		CreatedAt:   r.CreatedAt,
	}
}
