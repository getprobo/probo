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

package types

import (
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
)

func NewCustomDomainConnection(
	p *page.Page[*coredata.CustomDomain, coredata.CustomDomainOrderField],
	parentType any,
	parentID gid.GID,
	cnameTarget string,
) *CustomDomainConnection {
	var edges = make([]*CustomDomainEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewCustomDomainEdge(p.Data[i], p.Cursor.OrderBy.Field, cnameTarget)
	}

	return &CustomDomainConnection{
		Edges:      edges,
		PageInfo:   NewPageInfo(p),
		TotalCount: len(p.Data),
	}
}

func NewCustomDomainEdge(
	d *coredata.CustomDomain,
	orderBy coredata.CustomDomainOrderField,
	cnameTarget string,
) *CustomDomainEdge {
	return &CustomDomainEdge{
		Cursor: d.CursorKey(orderBy),
		Node:   NewCustomDomain(d, cnameTarget),
	}
}

func NewCustomDomain(d *coredata.CustomDomain, cnameTarget string) *CustomDomain {
	result := &CustomDomain{
		ID:           d.ID,
		Domain:       d.Domain,
		IsActive:     d.IsActive,
		SslStatus:    d.SSLStatus,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
		SslExpiresAt: d.SSLExpiresAt,
	}

	// Convert DNS records
	result.DNSRecords = convertDNSRecords(d, cnameTarget)

	return result
}

func convertDNSRecords(d *coredata.CustomDomain, cnameTarget string) []*DNSRecordInstruction {
	var records []*DNSRecordInstruction

	// For HTTP-01 challenges, we just need the domain to point to our servers via CNAME
	record := &DNSRecordInstruction{
		Type:    "CNAME",
		Name:    d.Domain,
		Value:   cnameTarget,
		TTL:     300,
		Purpose: "Point domain to Probo servers",
	}
	records = append(records, record)

	return records
}
