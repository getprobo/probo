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
	"fmt"
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

func insertDevicePostureWithEvidence(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx devicePostureFixture,
	checkKey string,
	status coredata.DevicePostureStatus,
	evidence map[string]any,
	correlationID gid.GID,
	createdAt time.Time,
) {
	t.Helper()

	raw, err := json.Marshal(evidence)
	require.NoError(t, err)

	posture := coredata.DevicePosture{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.DevicePostureEntityType),
		OrganizationID: fx.organizationID,
		DeviceID:       fx.deviceID,
		CorrelationID:  correlationID,
		CheckKey:       checkKey,
		Status:         status,
		Evidence:       raw,
		ObservedAt:     createdAt,
		CreatedAt:      createdAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return posture.Insert(ctx, tx, fx.scope)
	}))
}

func TestDevicePostureReport_LoadByDeviceID_GroupsByCorrelationID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedDevicePostureFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	older := now.Add(-time.Hour)
	newer := now
	olderCorr := gid.New(fx.scope.GetTenantID(), coredata.DevicePostureReportEntityType)
	newerCorr := gid.New(fx.scope.GetTenantID(), coredata.DevicePostureReportEntityType)

	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"OS_VERSION",
		coredata.DevicePostureStatusPass,
		map[string]any{"product_version": "14.0"},
		olderCorr,
		older,
	)
	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"DISK_ENCRYPTION",
		coredata.DevicePostureStatusPass,
		map[string]any{"raw": "FileVault is On."},
		olderCorr,
		older,
	)
	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"OS_VERSION",
		coredata.DevicePostureStatusPass,
		map[string]any{"product_version": "15.4"},
		newerCorr,
		newer,
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		orderBy := page.OrderBy[coredata.DevicePostureReportOrderField]{
			Field:     coredata.DevicePostureReportOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		}
		cursor := page.NewCursor(10, nil, page.Head, orderBy)

		var reports coredata.DevicePostureReports
		require.NoError(t, reports.LoadByDeviceID(ctx, conn, fx.scope, fx.deviceID, cursor))

		p := page.NewPage(reports, cursor)
		require.Len(t, p.Data, 2)
		assert.Equal(t, newerCorr, p.Data[0].ID)
		assert.Equal(t, olderCorr, p.Data[1].ID)
		assert.True(t, p.Data[0].CreatedAt.Equal(newer))
		assert.True(t, p.Data[1].CreatedAt.Equal(older))

		correlationIDs := []gid.GID{p.Data[0].ID, p.Data[1].ID}

		var postures coredata.DevicePostures
		require.NoError(t, postures.LoadByDeviceIDAndCorrelationIDs(
			ctx, conn, fx.scope, fx.deviceID, correlationIDs,
		))
		require.Len(t, postures, 3)

		var counter coredata.DevicePostureReports

		count, err := counter.CountByDeviceID(ctx, conn, fx.scope, fx.deviceID)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		return nil
	}))
}

func TestDevicePostureReport_LoadByDeviceID_IDIsCorrelationID(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedDevicePostureFixture(t, ctx, client)

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	correlationID := gid.New(fx.scope.GetTenantID(), coredata.DevicePostureReportEntityType)

	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"OS_VERSION",
		coredata.DevicePostureStatusPass,
		map[string]any{"product_version": "15.4"},
		correlationID,
		createdAt,
	)
	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"DISK_ENCRYPTION",
		coredata.DevicePostureStatusPass,
		map[string]any{"raw": "FileVault is On."},
		correlationID,
		createdAt,
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		reports := loadDevicePostureReports(t, ctx, conn, fx, 10, nil)
		require.Len(t, reports, 1)

		report := reports[0]
		assert.Equal(t, correlationID, report.ID)

		var postures coredata.DevicePostures
		require.NoError(t, postures.LoadByDeviceIDAndCorrelationIDs(
			ctx, conn, fx.scope, fx.deviceID, []gid.GID{report.ID},
		))
		require.Len(t, postures, 2)

		for _, posture := range postures {
			assert.Equal(t, correlationID, posture.CorrelationID)
		}

		return nil
	}))
}

func TestDevicePostureReport_LoadByDeviceID_PaginatesAcrossPages(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedDevicePostureFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)
	createdAts := []time.Time{
		now.Add(-2 * time.Hour),
		now.Add(-time.Hour),
		now,
	}

	correlationIDs := make([]gid.GID, len(createdAts))
	for i := range createdAts {
		correlationIDs[i] = gid.New(fx.scope.GetTenantID(), coredata.DevicePostureReportEntityType)
	}

	for i, createdAt := range createdAts {
		insertDevicePostureWithEvidence(
			t, ctx, client, fx,
			"OS_VERSION",
			coredata.DevicePostureStatusPass,
			map[string]any{"product_version": fmt.Sprintf("15.%d", i)},
			correlationIDs[i],
			createdAt,
		)
		insertDevicePostureWithEvidence(
			t, ctx, client, fx,
			"DISK_ENCRYPTION",
			coredata.DevicePostureStatusPass,
			map[string]any{"raw": "FileVault is On."},
			correlationIDs[i],
			createdAt,
		)
	}

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		orderBy := page.OrderBy[coredata.DevicePostureReportOrderField]{
			Field:     coredata.DevicePostureReportOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		}

		first := loadDevicePostureReports(t, ctx, conn, fx, 2, nil)
		require.Len(t, first, 2)
		assert.Equal(t, correlationIDs[2], first[0].ID)
		assert.Equal(t, correlationIDs[1], first[1].ID)
		assert.True(t, first[0].CreatedAt.Equal(createdAts[2]))
		assert.True(t, first[1].CreatedAt.Equal(createdAts[1]))

		after := first[1].CursorKey(orderBy.Field)

		second := loadDevicePostureReports(t, ctx, conn, fx, 2, &after)
		require.Len(t, second, 1)
		assert.Equal(t, correlationIDs[0], second[0].ID)
		assert.True(t, second[0].CreatedAt.Equal(createdAts[0]))

		return nil
	}))
}

func TestDevicePostureReport_LoadByDeviceID_IsTenantScoped(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedDevicePostureFixture(t, ctx, client)
	other := seedDevicePostureFixture(t, ctx, client)

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	correlationID := gid.New(fx.scope.GetTenantID(), coredata.DevicePostureReportEntityType)

	insertDevicePostureWithEvidence(
		t, ctx, client, fx,
		"OS_VERSION",
		coredata.DevicePostureStatusPass,
		map[string]any{"product_version": "15.4"},
		correlationID,
		createdAt,
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		orderBy := page.OrderBy[coredata.DevicePostureReportOrderField]{
			Field:     coredata.DevicePostureReportOrderFieldCreatedAt,
			Direction: page.OrderDirectionDesc,
		}
		cursor := page.NewCursor(10, nil, page.Head, orderBy)

		var reports coredata.DevicePostureReports
		require.NoError(t, reports.LoadByDeviceID(
			ctx, conn, other.scope, fx.deviceID, cursor,
		))
		assert.Empty(t, reports)

		count, err := reports.CountByDeviceID(ctx, conn, other.scope, fx.deviceID)
		require.NoError(t, err)
		assert.Zero(t, count)

		var postures coredata.DevicePostures
		require.NoError(t, postures.LoadByDeviceIDAndCorrelationIDs(
			ctx, conn, other.scope, fx.deviceID, []gid.GID{correlationID},
		))
		assert.Empty(t, postures)

		return nil
	}))
}

func loadDevicePostureReports(
	t *testing.T,
	ctx context.Context,
	conn pg.Querier,
	fx devicePostureFixture,
	size int,
	from *page.CursorKey,
) coredata.DevicePostureReports {
	t.Helper()

	orderBy := page.OrderBy[coredata.DevicePostureReportOrderField]{
		Field:     coredata.DevicePostureReportOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}
	cursor := page.NewCursor(size, from, page.Head, orderBy)

	var reports coredata.DevicePostureReports
	require.NoError(t, reports.LoadByDeviceID(ctx, conn, fx.scope, fx.deviceID, cursor))

	return page.NewPage(reports, cursor).Data
}
