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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	RegistryAssetOrderBy OrderBy[coredata.RegistryAssetOrderField]

	RegistryAssetConnection struct {
		TotalCount int
		Edges      []*RegistryAssetEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}
)

func NewRegistryAssetConnection(
	p *page.Page[*coredata.RegistryAsset, coredata.RegistryAssetOrderField],
	resolver any,
	parentID gid.GID,
) *RegistryAssetConnection {
	edges := make([]*RegistryAssetEdge, len(p.Data))
	for i, asset := range p.Data {
		edges[i] = NewRegistryAssetEdge(asset, p.Cursor.OrderBy.Field)
	}

	return &RegistryAssetConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),

		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewRegistryAssetEdge(asset *coredata.RegistryAsset, orderField coredata.RegistryAssetOrderField) *RegistryAssetEdge {
	return &RegistryAssetEdge{
		Node:   NewRegistryAsset(asset),
		Cursor: asset.CursorKey(orderField),
	}
}

func NewRegistryAsset(asset *coredata.RegistryAsset) *RegistryAsset {
	return &RegistryAsset{
		ID:     asset.ID,
		Name:   asset.Name,
		Amount: asset.Amount,
		Owner: &Profile{
			ID: asset.OwnerID,
		},
		RegistryAssetType: asset.RegistryAssetType,
		DataTypesStored:   asset.DataTypesStored,
		CreatedAt:         asset.CreatedAt,
		UpdatedAt:         asset.UpdatedAt,
	}
}
