package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

type ComplianceExternalURLOrderBy = OrderBy[coredata.ComplianceExternalURLOrderField]

type ComplianceExternalURLConnection struct {
	Edges    []*ComplianceExternalURLEdge `json:"edges"`
	PageInfo *PageInfo                    `json:"pageInfo"`
}

func NewComplianceExternalURL(c *coredata.ComplianceExternalURL) *ComplianceExternalURL {
	return &ComplianceExternalURL{
		ID:        c.ID,
		Name:      c.Name,
		URL:       c.URL,
		Rank:      c.Rank,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func NewComplianceExternalURLConnection(
	p *page.Page[*coredata.ComplianceExternalURL, coredata.ComplianceExternalURLOrderField],
) *ComplianceExternalURLConnection {
	edges := make([]*ComplianceExternalURLEdge, len(p.Data))
	for i := range edges {
		edges[i] = NewComplianceExternalURLEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}
	return &ComplianceExternalURLConnection{
		Edges:    edges,
		PageInfo: NewPageInfo(p),
	}
}

func NewComplianceExternalURLEdge(c *coredata.ComplianceExternalURL, orderBy coredata.ComplianceExternalURLOrderField) *ComplianceExternalURLEdge {
	return &ComplianceExternalURLEdge{
		Cursor: c.CursorKey(orderBy),
		Node:   NewComplianceExternalURL(c),
	}
}
