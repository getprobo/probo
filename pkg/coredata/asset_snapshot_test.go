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

package coredata_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

func TestAssetSnapshot_CopyActiveAssets(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	now := time.Now().UTC().Round(time.Microsecond)
	org := insertSnapshotOrganization(t, client, now)
	other := insertSnapshotOrganization(t, client, now)
	scope := coredata.NewScope(org.tenantID, gid.Nil)

	manualID := gid.New(org.tenantID, coredata.AssetEntityType)
	discoveredID := gid.New(org.tenantID, coredata.AssetEntityType)
	retiredID := gid.New(org.tenantID, coredata.AssetEntityType)
	otherID := gid.New(other.tenantID, coredata.AssetEntityType)
	thirdPartyID := gid.New(org.tenantID, coredata.ThirdPartyEntityType)
	businessFunctionID := gid.New(org.tenantID, coredata.BusinessFunctionEntityType)
	tagID := gid.New(org.tenantID, coredata.TagEntityType)
	connectorID := gid.New(org.tenantID, coredata.ConnectorEntityType)
	snapshotID := gid.New(org.tenantID, coredata.AssetSnapshotEntityType)

	kind := "object-storage"
	externalID := "bucket-1"
	dataTypes := "customer records"
	name := "quarter"
	note := "kept as written"

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		manual := &coredata.Asset{
			ID:              manualID,
			Name:            "alpha-manual",
			Amount:          4,
			OrganizationID:  org.organizationID,
			Type:            coredata.AssetTypePhysical,
			DataTypesStored: &dataTypes,
			Status:          coredata.AssetStatusActive,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := manual.Insert(ctx, tx, scope); err != nil {
			return err
		}

		discovered := &coredata.Asset{
			ID:             discoveredID,
			Name:           "zeta-discovered",
			Amount:         2,
			OrganizationID: org.organizationID,
			Type:           coredata.AssetTypeVirtual,
			Kind:           &kind,
			ExternalID:     &externalID,
			Configuration:  json.RawMessage(`{"encrypted":true}`),
			Status:         coredata.AssetStatusActive,
			ConnectorID:    &connectorID,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := discovered.Insert(ctx, tx, scope); err != nil {
			return err
		}

		retired := &coredata.Asset{
			ID:             retiredID,
			Name:           "middle-retired",
			Amount:         9,
			OrganizationID: org.organizationID,
			Type:           coredata.AssetTypePhysical,
			Status:         coredata.AssetStatusDeactivated,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := retired.Insert(ctx, tx, scope); err != nil {
			return err
		}

		otherScope := coredata.NewScope(other.tenantID, gid.Nil)
		otherAsset := &coredata.Asset{
			ID:             otherID,
			Name:           "other-active",
			Amount:         1,
			OrganizationID: other.organizationID,
			Type:           coredata.AssetTypeVirtual,
			Status:         coredata.AssetStatusActive,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := otherAsset.Insert(ctx, tx, otherScope); err != nil {
			return err
		}

		thirdParty := coredata.ThirdParty{
			ID:             thirdPartyID,
			OrganizationID: org.organizationID,
			Name:           "Copied vendor",
			Category:       coredata.ThirdPartyCategoryOther,
			Certifications: []string{},
			Countries:      coredata.CountryCodes{},
			Level:          1,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := thirdParty.Insert(ctx, tx, scope); err != nil {
			return err
		}

		if err := (coredata.AssetThirdParties{}).Insert(
			ctx,
			tx,
			scope,
			manualID,
			org.organizationID,
			[]gid.GID{thirdPartyID},
		); err != nil {
			return err
		}

		businessFunction := &coredata.BusinessFunction{
			ID:             businessFunctionID,
			OrganizationID: org.organizationID,
			Name:           "Billing",
			Classification: coredata.BusinessFunctionClassificationStandard,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := businessFunction.Insert(ctx, tx, scope); err != nil {
			return err
		}

		if err := (coredata.BusinessFunctionAssets{}).Insert(
			ctx,
			tx,
			scope,
			businessFunctionID,
			org.organizationID,
			[]gid.GID{manualID},
		); err != nil {
			return err
		}

		tag := &coredata.Tag{
			ID:             tagID,
			OrganizationID: org.organizationID,
			Name:           "copied-tag",
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tag.Insert(ctx, tx, scope); err != nil {
			return err
		}

		assignment := &coredata.TagAssignment{
			ID:             gid.New(org.tenantID, coredata.TagAssignmentEntityType),
			OrganizationID: org.organizationID,
			ResourceID:     manualID,
			TagID:          tagID,
			CreatedAt:      now,
		}
		if err := assignment.Insert(ctx, tx, scope); err != nil {
			return err
		}

		snapshot := &coredata.AssetSnapshot{
			ID:             snapshotID,
			OrganizationID: org.organizationID,
			CapturedAt:     now,
			Name:           &name,
			Note:           &note,
		}
		if err := snapshot.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return snapshot.CopyActiveAssets(ctx, tx, scope)
	}))

	copied := loadSnapshotAssets(t, client, scope, snapshotID)
	require.Len(t, copied, 2)
	assert.Equal(t, manualID, copied[0].AssetID)
	assert.Equal(t, "alpha-manual", copied[0].Name)
	assert.Equal(t, 4, copied[0].Amount)
	assert.Equal(t, discoveredID, copied[1].AssetID)
	assert.Equal(t, "zeta-discovered", copied[1].Name)
	assert.Equal(t, &kind, copied[1].Kind)
	assert.Equal(t, &externalID, copied[1].ExternalID)

	var configuration map[string]any
	require.NoError(t, json.Unmarshal(copied[1].Configuration, &configuration))
	assert.Equal(t, true, configuration["encrypted"])

	thirdParties, err := copied[0].ThirdParties()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{thirdPartyID}, thirdParties)

	functions, err := copied[0].BusinessFunctions()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{businessFunctionID}, functions)

	tags, err := copied[0].Tags()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{tagID}, tags)

	discoveredParties, err := copied[1].ThirdParties()
	require.NoError(t, err)
	assert.Empty(t, discoveredParties)

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		if _, err := tx.Exec(
			ctx,
			`UPDATE assets SET name = 'renamed', amount = 99, status = 'DEACTIVATED' WHERE id = $1`,
			manualID,
		); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM asset_third_parties WHERE asset_id = $1`, manualID); err != nil {
			return err
		}

		if _, err := tx.Exec(ctx, `DELETE FROM tag_assignments WHERE resource_id = $1`, manualID); err != nil {
			return err
		}

		_, err := tx.Exec(
			ctx,
			`DELETE FROM business_function_assets WHERE asset_id = $1`,
			manualID,
		)

		return err
	}))

	after := loadSnapshotAssets(t, client, scope, snapshotID)
	require.Len(t, after, 2)
	assert.Equal(t, "alpha-manual", after[0].Name)
	assert.Equal(t, 4, after[0].Amount)
	assert.Equal(t, coredata.AssetStatusActive, after[0].Status)

	thirdParties, err = after[0].ThirdParties()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{thirdPartyID}, thirdParties)

	tags, err = after[0].Tags()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{tagID}, tags)
}

type snapshotOrganization struct {
	tenantID       gid.TenantID
	organizationID gid.GID
}

func insertSnapshotOrganization(
	t *testing.T,
	client *pg.Client,
	now time.Time,
) snapshotOrganization {
	t.Helper()

	tenantID := gid.NewTenantID()
	org := snapshotOrganization{
		tenantID:       tenantID,
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		_, err := tx.Exec(
			ctx,
			`INSERT INTO organizations (id, tenant_id, name, created_at, updated_at, asset_inventory_enabled)
			 VALUES ($1, $2, $3, $4, $5, false)`,
			org.organizationID,
			org.tenantID,
			"asset-snapshot-"+org.organizationID.String(),
			now,
			now,
		)

		return err
	}))

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, org.organizationID)

			return err
		})
	})

	return org
}

func loadSnapshotAssets(
	t *testing.T,
	client *pg.Client,
	scope coredata.Scoper,
	snapshotID gid.GID,
) coredata.AssetSnapshotAssets {
	t.Helper()

	var assets coredata.AssetSnapshotAssets

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return assets.LoadBySnapshotID(
			ctx,
			conn,
			scope,
			snapshotID,
			page.NewCursor(
				50,
				nil,
				page.Head,
				page.OrderBy[coredata.AssetSnapshotAssetOrderField]{
					Field:     coredata.AssetSnapshotAssetOrderFieldName,
					Direction: page.OrderDirectionAsc,
				},
			),
		)
	}))

	return assets
}
