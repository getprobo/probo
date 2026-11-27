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
	"fmt"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
)

func NewAssetSnapshot(snapshot *coredata.AssetSnapshot) *AssetSnapshot {
	return &AssetSnapshot{
		ID:             snapshot.ID,
		OrganizationID: snapshot.OrganizationID,
		CapturedAt:     snapshot.CapturedAt,
		ActorID:        snapshot.ActorID,
		Name:           snapshot.Name,
		Note:           snapshot.Note,
	}
}

func NewAssetSnapshotAsset(asset *coredata.AssetSnapshotAsset) (*AssetSnapshotAsset, error) {
	thirdParties, err := asset.ThirdParties()
	if err != nil {
		return nil, fmt.Errorf("cannot read snapshot third parties: %w", err)
	}

	businessFunctions, err := asset.BusinessFunctions()
	if err != nil {
		return nil, fmt.Errorf("cannot read snapshot business functions: %w", err)
	}

	tags, err := asset.Tags()
	if err != nil {
		return nil, fmt.Errorf("cannot read snapshot tags: %w", err)
	}

	return &AssetSnapshotAsset{
		ID:                  asset.ID,
		SnapshotID:          asset.SnapshotID,
		OrganizationID:      asset.OrganizationID,
		AssetID:             asset.AssetID,
		Name:                asset.Name,
		Amount:              asset.Amount,
		OwnerID:             asset.OwnerID,
		AssetType:           asset.AssetType,
		DataTypesStored:     asset.DataTypesStored,
		Kind:                asset.Kind,
		ExternalID:          asset.ExternalID,
		Region:              asset.Region,
		Configuration:       rawJSONText(asset.Configuration),
		SchemaVersion:       asset.SchemaVersion,
		LastSeenAt:          asset.LastSeenAt,
		Status:              asset.Status,
		ConnectorID:         asset.ConnectorID,
		ThirdPartyIds:       thirdParties,
		BusinessFunctionIds: businessFunctions,
		TagIds:              tags,
		AssetCreatedAt:      asset.AssetCreatedAt,
		AssetUpdatedAt:      asset.AssetUpdatedAt,
	}, nil
}

func NewListAssetSnapshotsOutput(
	snapshotPage *page.Page[*coredata.AssetSnapshot, coredata.AssetSnapshotOrderField],
) ListAssetSnapshotsOutput {
	snapshots := make([]*AssetSnapshot, 0, len(snapshotPage.Data))
	for _, snapshot := range snapshotPage.Data {
		snapshots = append(snapshots, NewAssetSnapshot(snapshot))
	}

	var nextCursor *page.CursorKey

	if snapshotPage.Info.HasNext && len(snapshotPage.Data) > 0 {
		cursorKey := snapshotPage.Data[len(snapshotPage.Data)-1].CursorKey(snapshotPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListAssetSnapshotsOutput{
		NextCursor:     nextCursor,
		AssetSnapshots: snapshots,
	}
}

func NewGetAssetSnapshotOutput(
	snapshot *coredata.AssetSnapshot,
	assetPage *page.Page[*coredata.AssetSnapshotAsset, coredata.AssetSnapshotAssetOrderField],
) (GetAssetSnapshotOutput, error) {
	assets := make([]*AssetSnapshotAsset, 0, len(assetPage.Data))
	for _, asset := range assetPage.Data {
		row, err := NewAssetSnapshotAsset(asset)
		if err != nil {
			return GetAssetSnapshotOutput{}, err
		}

		assets = append(assets, row)
	}

	var nextCursor *page.CursorKey

	if assetPage.Info.HasNext && len(assetPage.Data) > 0 {
		cursorKey := assetPage.Data[len(assetPage.Data)-1].CursorKey(assetPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return GetAssetSnapshotOutput{
		AssetSnapshot: NewAssetSnapshot(snapshot),
		Assets:        assets,
		NextCursor:    nextCursor,
	}, nil
}
