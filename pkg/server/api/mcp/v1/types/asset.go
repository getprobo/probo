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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewAsset(a *coredata.Asset) *Asset {
	return &Asset{
		ID:              a.ID,
		Name:            a.Name,
		Amount:          a.Amount,
		OwnerID:         a.OwnerID,
		OrganizationID:  a.OrganizationID,
		AssetType:       a.AssetType,
		DataTypesStored: a.DataTypesStored,
		CreatedAt:       a.CreatedAt,
		UpdatedAt:       a.UpdatedAt,
	}
}

func NewListAssetsOutput(assetPage *page.Page[*coredata.Asset, coredata.AssetOrderField]) ListAssetsOutput {
	assets := make([]*Asset, 0, len(assetPage.Data))
	for _, v := range assetPage.Data {
		assets = append(assets, NewAsset(v))
	}

	var nextCursor *page.CursorKey
	if len(assetPage.Data) > 0 {
		cursorKey := assetPage.Data[len(assetPage.Data)-1].CursorKey(assetPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAssetsOutput{
		NextCursor: nextCursor,
		Assets:     assets,
	}
}
