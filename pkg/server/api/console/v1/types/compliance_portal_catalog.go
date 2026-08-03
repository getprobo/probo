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
	"go.probo.inc/probo/pkg/complianceportal/management"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	CompliancePortalDocument struct {
		ID         gid.GID
		Document   *Document
		Visibility coredata.CompliancePortalVisibility
	}

	CompliancePortalDocumentConnection struct {
		TotalCount int
		Edges      []*CompliancePortalDocumentEdge
		PageInfo   *PageInfo
		ParentID   gid.GID
	}

	CompliancePortalAudit struct {
		ID         gid.GID
		Audit      *Audit
		Visibility coredata.CompliancePortalVisibility
	}

	CompliancePortalAuditConnection struct {
		TotalCount int
		Edges      []*CompliancePortalAuditEdge
		PageInfo   *PageInfo
		ParentID   gid.GID
	}

	CompliancePortalThirdParty struct {
		ID         gid.GID
		ThirdParty *ThirdParty
	}

	CompliancePortalThirdPartyConnection struct {
		TotalCount int
		Edges      []*CompliancePortalThirdPartyEdge
		PageInfo   *PageInfo
		ParentID   gid.GID
	}
)

func NewCompliancePortalDocument(entry *management.PortalDocument) *CompliancePortalDocument {
	return &CompliancePortalDocument{
		ID:         entry.ID,
		Document:   NewDocument(entry.Document),
		Visibility: entry.Visibility,
	}
}

func NewCompliancePortalDocumentEdge(entry *management.PortalDocument, orderBy coredata.DocumentOrderField) *CompliancePortalDocumentEdge {
	return &CompliancePortalDocumentEdge{
		Cursor: entry.Document.CursorKey(orderBy),
		Node:   NewCompliancePortalDocument(entry),
	}
}

func NewCompliancePortalDocumentConnection(
	p *page.Page[*management.PortalDocument, coredata.DocumentOrderField],
	parentID gid.GID,
) *CompliancePortalDocumentConnection {
	edges := make([]*CompliancePortalDocumentEdge, len(p.Data))
	for i, entry := range p.Data {
		edges[i] = NewCompliancePortalDocumentEdge(entry, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalDocumentConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewCompliancePortalAudit(entry *management.PortalAudit) *CompliancePortalAudit {
	return &CompliancePortalAudit{
		ID:         entry.ID,
		Audit:      NewAudit(entry.Audit),
		Visibility: entry.Visibility,
	}
}

func NewCompliancePortalAuditEdge(entry *management.PortalAudit, orderBy coredata.AuditOrderField) *CompliancePortalAuditEdge {
	return &CompliancePortalAuditEdge{
		Cursor: entry.Audit.CursorKey(orderBy),
		Node:   NewCompliancePortalAudit(entry),
	}
}

func NewCompliancePortalAuditConnection(
	p *page.Page[*management.PortalAudit, coredata.AuditOrderField],
	parentID gid.GID,
) *CompliancePortalAuditConnection {
	edges := make([]*CompliancePortalAuditEdge, len(p.Data))
	for i, entry := range p.Data {
		edges[i] = NewCompliancePortalAuditEdge(entry, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalAuditConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewCompliancePortalThirdParty(entry *management.PortalThirdParty) *CompliancePortalThirdParty {
	return &CompliancePortalThirdParty{
		ID:         entry.ID,
		ThirdParty: NewThirdParty(entry.ThirdParty),
	}
}

func NewCompliancePortalThirdPartyEdge(entry *management.PortalThirdParty, orderBy coredata.ThirdPartyOrderField) *CompliancePortalThirdPartyEdge {
	return &CompliancePortalThirdPartyEdge{
		Cursor: entry.ThirdParty.CursorKey(orderBy),
		Node:   NewCompliancePortalThirdParty(entry),
	}
}

func NewCompliancePortalThirdPartyConnection(
	p *page.Page[*management.PortalThirdParty, coredata.ThirdPartyOrderField],
	parentID gid.GID,
) *CompliancePortalThirdPartyConnection {
	edges := make([]*CompliancePortalThirdPartyEdge, len(p.Data))
	for i, entry := range p.Data {
		edges[i] = NewCompliancePortalThirdPartyEdge(entry, p.Cursor.OrderBy.Field)
	}

	return &CompliancePortalThirdPartyConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}
