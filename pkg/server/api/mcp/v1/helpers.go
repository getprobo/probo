// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package mcp_v1

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func allApproversCursor() *page.Cursor[coredata.MembershipProfileOrderField] {
	return page.NewCursor(100, nil, page.Head, page.OrderBy[coredata.MembershipProfileOrderField]{
		Field:     coredata.MembershipProfileOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	})
}

func profileIDs(p *page.Page[*coredata.MembershipProfile, coredata.MembershipProfileOrderField]) []gid.GID {
	ids := make([]gid.GID, len(p.Data))
	for i, profile := range p.Data {
		ids[i] = profile.ID
	}
	return ids
}
