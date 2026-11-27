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

package coredata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/gid"
)

func TestDiffAssetChange_Create(t *testing.T) {
	t.Parallel()

	asset := sampleAsset()

	changeType, before, after, write, err := DiffAssetChange(nil, asset, AssetChangeInput{})
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeCreated, changeType)
	assert.Nil(t, before)
	assert.Contains(t, string(after), `"name":"cdn"`)
	assert.Contains(t, string(after), `"status":"ACTIVE"`)
}

func TestDiffAssetChange_UnchangedFieldsWriteNothing(t *testing.T) {
	t.Parallel()

	asset := sampleAsset()

	_, _, _, write, err := DiffAssetChange(asset, asset, AssetChangeInput{})
	require.NoError(t, err)
	assert.False(t, write)
}

func TestDiffAssetChange_RefreshRecordsLastSeen(t *testing.T) {
	t.Parallel()

	asset := sampleAsset()

	changeType, before, after, write, err := DiffAssetChange(
		asset,
		asset,
		AssetChangeInput{RecordLastSeen: true},
	)
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeUpdated, changeType)
	assert.Contains(t, string(before), "last_seen_at")
	assert.Contains(t, string(after), "last_seen_at")
	assert.NotContains(t, string(after), `"name"`)
}

func TestDiffAssetChange_FieldUpdate(t *testing.T) {
	t.Parallel()

	beforeAsset := sampleAsset()
	afterAsset := *beforeAsset
	afterAsset.Name = "edge"
	afterAsset.Amount = 3

	changeType, before, after, write, err := DiffAssetChange(
		beforeAsset,
		&afterAsset,
		AssetChangeInput{},
	)
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeUpdated, changeType)
	assert.JSONEq(t, `{"amount":1,"name":"cdn"}`, string(before))
	assert.JSONEq(t, `{"amount":3,"name":"edge"}`, string(after))
}

func TestDiffAssetChange_LinkSet(t *testing.T) {
	t.Parallel()

	asset := sampleAsset()
	tenantID := gid.NewTenantID()
	removed := gid.New(tenantID, ThirdPartyEntityType)
	added := gid.New(tenantID, ThirdPartyEntityType)
	kept := gid.New(tenantID, BusinessFunctionEntityType)
	linked := gid.New(tenantID, BusinessFunctionEntityType)

	changeType, before, after, write, err := DiffAssetChange(
		asset,
		asset,
		AssetChangeInput{
			ThirdParties: &AssetLinkChange{
				Before: []gid.GID{removed},
				After:  []gid.GID{added},
			},
		},
	)
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeUpdated, changeType)
	assert.JSONEq(t, `{"third_party_ids":["`+removed.String()+`"]}`, string(before))
	assert.JSONEq(t, `{"third_party_ids":["`+added.String()+`"]}`, string(after))

	changeType, before, after, write, err = DiffAssetChange(
		asset,
		asset,
		AssetChangeInput{
			BusinessFunctions: &AssetLinkChange{
				Before: []gid.GID{kept},
				After:  []gid.GID{kept, linked},
			},
		},
	)
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeUpdated, changeType)
	assert.Contains(t, string(before), kept.String())
	assert.NotContains(t, string(before), linked.String())
	assert.Contains(t, string(after), kept.String())
	assert.Contains(t, string(after), linked.String())
}

func TestDiffAssetChange_Deactivate(t *testing.T) {
	t.Parallel()

	beforeAsset := sampleAsset()
	afterAsset := *beforeAsset
	afterAsset.Status = AssetStatusDeactivated

	changeType, before, after, write, err := DiffAssetChange(
		beforeAsset,
		&afterAsset,
		AssetChangeInput{},
	)
	require.NoError(t, err)
	require.True(t, write)
	assert.Equal(t, AssetChangeTypeDeactivated, changeType)
	assert.Contains(t, string(before), `"status":"ACTIVE"`)
	assert.Contains(t, string(after), `"status":"DEACTIVATED"`)
}

func sampleAsset() *Asset {
	seen := time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC)
	kind := "aws.cloudfront.distribution"
	externalID := "dist-1"

	return &Asset{
		Name:       "cdn",
		Amount:     1,
		Type:       AssetTypeVirtual,
		Status:     AssetStatusActive,
		Kind:       &kind,
		ExternalID: &externalID,
		LastSeenAt: &seen,
	}
}
