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
	"go.probo.inc/probo/pkg/page"
)

func NewRegistryAsset(a *coredata.RegistryAsset) *RegistryAsset {
	return &RegistryAsset{
		ID:                a.ID,
		Name:              a.Name,
		Amount:            a.Amount,
		OwnerID:           a.OwnerID,
		OrganizationID:    a.OrganizationID,
		RegistryAssetType: a.RegistryAssetType,
		DataTypesStored:   a.DataTypesStored,
		CreatedAt:         a.CreatedAt,
		UpdatedAt:         a.UpdatedAt,
	}
}

func NewListRegistryAssetsOutput(assetPage *page.Page[*coredata.RegistryAsset, coredata.RegistryAssetOrderField]) ListRegistryAssetsOutput {
	assets := make([]*RegistryAsset, 0, len(assetPage.Data))
	for _, v := range assetPage.Data {
		assets = append(assets, NewRegistryAsset(v))
	}

	var nextCursor *page.CursorKey

	if len(assetPage.Data) > 0 {
		cursorKey := assetPage.Data[len(assetPage.Data)-1].CursorKey(assetPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListRegistryAssetsOutput{
		NextCursor:     nextCursor,
		RegistryAssets: assets,
	}
}
