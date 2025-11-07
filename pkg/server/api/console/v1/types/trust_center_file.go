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
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type TrustCenterFile struct {
	ID                    gid.GID                        `json:"id"`
	OrganizationID        gid.GID                        `json:"-"`
	Name                  string                         `json:"name"`
	Category              string                         `json:"category"`
	FileURL               string                         `json:"fileUrl"`
	TrustCenterVisibility coredata.TrustCenterVisibility `json:"trustCenterVisibility"`
	CreatedAt             time.Time                      `json:"createdAt"`
	UpdatedAt             time.Time                      `json:"updatedAt"`
	Organization          *Organization                  `json:"organization"`
}

func (TrustCenterFile) IsNode()             {}
func (this TrustCenterFile) GetID() gid.GID { return this.ID }

type TrustCenterFileOrderBy = OrderBy[coredata.TrustCenterFileOrderField]

type TrustCenterFileConnection struct {
	TotalCount int                    `json:"totalCount"`
	Edges      []*TrustCenterFileEdge `json:"edges"`
	PageInfo   *PageInfo              `json:"pageInfo"`
	ParentID   gid.GID                `json:"-"`
}

func NewTrustCenterFile(tcf *coredata.TrustCenterFile) *TrustCenterFile {
	return &TrustCenterFile{
		ID:                    tcf.ID,
		OrganizationID:        tcf.OrganizationID,
		Name:                  tcf.Name,
		Category:              tcf.Category,
		TrustCenterVisibility: tcf.TrustCenterVisibility,
		CreatedAt:             tcf.CreatedAt,
		UpdatedAt:             tcf.UpdatedAt,
	}
}

func NewTrustCenterFileConnection(
	p *page.Page[*coredata.TrustCenterFile, coredata.TrustCenterFileOrderField],
	parentID gid.GID,
) *TrustCenterFileConnection {
	var edges = make([]*TrustCenterFileEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewTrustCenterFileEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &TrustCenterFileConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
		ParentID: parentID,
	}
}

func NewTrustCenterFileEdge(tcf *coredata.TrustCenterFile, orderBy coredata.TrustCenterFileOrderField) *TrustCenterFileEdge {
	return &TrustCenterFileEdge{
		Cursor: tcf.CursorKey(orderBy),
		Node:   NewTrustCenterFile(tcf),
	}
}
