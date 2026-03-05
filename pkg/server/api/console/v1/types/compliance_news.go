package types

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	ComplianceNewsConnection struct {
		TotalCount int
		Edges      []*ComplianceNewsEdge
		PageInfo   *PageInfo

		Resolver      any
		TrustCenterID gid.GID
	}
)

func NewComplianceNews(cn *coredata.ComplianceNews) *ComplianceNews {
	return &ComplianceNews{
		ID:        cn.ID,
		Title:     cn.Title,
		Body:      cn.Body,
		Status:    cn.Status,
		CreatedAt: cn.CreatedAt,
		UpdatedAt: cn.UpdatedAt,
	}
}

func NewComplianceNewsEdge(cn *coredata.ComplianceNews, orderBy coredata.ComplianceNewsOrderField) *ComplianceNewsEdge {
	return &ComplianceNewsEdge{
		Cursor: cn.CursorKey(orderBy),
		Node:   NewComplianceNews(cn),
	}
}

func NewComplianceNewsConnection(
	p *page.Page[*coredata.ComplianceNews, coredata.ComplianceNewsOrderField],
	resolver any,
	trustCenterID gid.GID,
) *ComplianceNewsConnection {
	edges := make([]*ComplianceNewsEdge, len(p.Data))

	for i := range edges {
		edges[i] = NewComplianceNewsEdge(p.Data[i], p.Cursor.OrderBy.Field)
	}

	return &ComplianceNewsConnection{
		Edges:         edges,
		PageInfo:      NewPageInfo(p),
		Resolver:      resolver,
		TrustCenterID: trustCenterID,
	}
}
