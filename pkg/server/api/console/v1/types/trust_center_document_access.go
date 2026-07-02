// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	TrustCenterDocumentAccessOrderBy = OrderBy[coredata.TrustCenterDocumentAccessOrderField]

	TrustCenterDocumentAccessConnection struct {
		TotalCount int
		Edges      []*TrustCenterDocumentAccessEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	TrustCenterDocumentAccess struct {
		ID                gid.GID                                  `json:"id"`
		OrganizationID    gid.GID                                  `json:"-"`
		Status            coredata.TrustCenterDocumentAccessStatus `json:"status"`
		CreatedAt         time.Time                                `json:"createdAt"`
		UpdatedAt         time.Time                                `json:"updatedAt"`
		TrustCenterAccess *TrustCenterAccess                       `json:"trustCenterAccess"`
		Document          *Document                                `json:"document,omitempty"`
		ReportFile        *File                                    `json:"reportFile,omitempty"`
		TrustCenterFile   *TrustCenterFile                         `json:"trustCenterFile,omitempty"`

		// Internal fields used by resolvers
		TrustCenterAccessID gid.GID  `json:"-"`
		DocumentID          *gid.GID `json:"-"`
		ReportFileID        *gid.GID `json:"-"`
		TrustCenterFileID   *gid.GID `json:"-"`
	}
)

func NewTrustCenterDocumentAccess(tcda *coredata.TrustCenterDocumentAccess) *TrustCenterDocumentAccess {
	object := &TrustCenterDocumentAccess{
		ID:                  tcda.ID,
		OrganizationID:      tcda.OrganizationID,
		Status:              tcda.Status,
		CreatedAt:           tcda.CreatedAt,
		UpdatedAt:           tcda.UpdatedAt,
		TrustCenterAccessID: tcda.TrustCenterAccessID,
		DocumentID:          tcda.DocumentID,
		ReportFileID:        tcda.ReportFileID,
		TrustCenterFileID:   tcda.TrustCenterFileID,
	}

	if tcda.DocumentID != nil {
		object.Document = &Document{
			ID: *tcda.DocumentID,
		}
	}

	if tcda.ReportFileID != nil {
		object.ReportFile = &File{
			ID: *tcda.ReportFileID,
		}
	}

	if tcda.TrustCenterFileID != nil {
		object.TrustCenterFile = &TrustCenterFile{
			ID: *tcda.TrustCenterFileID,
		}
	}

	return object
}

func NewTrustCenterDocumentAccessConnection(
	p *page.Page[*coredata.TrustCenterDocumentAccess, coredata.TrustCenterDocumentAccessOrderField],
	parentType any,
	parentID gid.GID,
) *TrustCenterDocumentAccessConnection {
	var edges = make([]*TrustCenterDocumentAccessEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTrustCenterDocumentAccessEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TrustCenterDocumentAccessConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: parentType,
		ParentID: parentID,
	}
}

func NewTrustCenterDocumentAccessEdge(access *coredata.TrustCenterDocumentAccess, orderBy coredata.TrustCenterDocumentAccessOrderField) *TrustCenterDocumentAccessEdge {
	return &TrustCenterDocumentAccessEdge{
		Cursor: access.CursorKey(orderBy),
		Node:   NewTrustCenterDocumentAccess(access),
	}
}
