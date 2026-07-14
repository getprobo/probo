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

type devicePostureFixture struct {
	scope          *coredata.Scope
	organizationID gid.GID
	deviceID       gid.GID
}

func seedDevicePostureFixture(t *testing.T, ctx context.Context, client *pg.Client) devicePostureFixture {
	t.Helper()

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	deviceID := gid.New(tenantID, coredata.DeviceEntityType)
	now := time.Now().UTC().Truncate(time.Microsecond)

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Device Posture Test Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		device := coredata.Device{
			ID:             deviceID,
			OrganizationID: organizationID,
			State:          coredata.DeviceStatePending,
			APIKeyHash:     []byte("device-posture-test-" + deviceID.String()),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := device.Insert(ctx, tx, scope); err != nil {
			return err
		}

		return nil
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			if _, err := tx.Exec(ctx, `DELETE FROM device_postures WHERE device_id = $1`, deviceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM devices WHERE id = $1`, deviceID); err != nil {
				return err
			}

			if _, err := tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID); err != nil {
				return err
			}

			return nil
		})
	})

	return devicePostureFixture{
		scope:          scope,
		organizationID: organizationID,
		deviceID:       deviceID,
	}
}

func insertDevicePosture(
	t *testing.T,
	ctx context.Context,
	client *pg.Client,
	fx devicePostureFixture,
	checkKey string,
	status coredata.DevicePostureStatus,
	observedAt time.Time,
) {
	t.Helper()

	posture := coredata.DevicePosture{
		ID:             gid.New(fx.scope.GetTenantID(), coredata.DevicePostureEntityType),
		OrganizationID: fx.organizationID,
		DeviceID:       fx.deviceID,
		CheckKey:       checkKey,
		Status:         status,
		ObservedAt:     observedAt,
		CreatedAt:      observedAt,
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return posture.Insert(ctx, tx, fx.scope)
	}))
}

func TestDevicePosture_LoadLatestByDeviceID_ReturnsLatestPerCheckKey(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	fx := seedDevicePostureFixture(t, ctx, client)

	now := time.Now().UTC().Truncate(time.Microsecond)

	insertDevicePosture(t, ctx, client, fx, "AUTO_UPDATE", coredata.DevicePostureStatusFail, now.Add(-2*time.Hour))
	insertDevicePosture(t, ctx, client, fx, "AUTO_UPDATE", coredata.DevicePostureStatusPass, now.Add(-time.Hour))
	insertDevicePosture(t, ctx, client, fx, "DISK_ENCRYPTION", coredata.DevicePostureStatusUnknown, now.Add(-30*time.Minute))
	insertDevicePosture(t, ctx, client, fx, "FIREWALL_ENABLED", coredata.DevicePostureStatusFail, now.Add(-time.Minute))

	var (
		pageOne coredata.DevicePostures
		hasNext bool
	)

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		orderBy := page.OrderBy[coredata.DevicePostureOrderField]{
			Field:     coredata.DevicePostureOrderFieldCheckKey,
			Direction: page.OrderDirectionAsc,
		}
		cursor := page.NewCursor(2, nil, page.Head, orderBy)

		var batch coredata.DevicePostures
		if err := batch.LoadLatestByDeviceID(ctx, conn, fx.scope, fx.deviceID, cursor); err != nil {
			return err
		}

		p := page.NewPage(batch, cursor)
		pageOne = p.Data
		hasNext = p.Info.HasNext

		return nil
	}))
	require.Len(t, pageOne, 2)
	require.True(t, hasNext)
	assert.Equal(t, "AUTO_UPDATE", pageOne[0].CheckKey)
	assert.Equal(t, coredata.DevicePostureStatusPass, pageOne[0].Status)
	assert.Equal(t, "DISK_ENCRYPTION", pageOne[1].CheckKey)

	var all coredata.DevicePostures

	require.NoError(t, client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		loaded, err := page.LoadAll(
			ctx,
			page.OrderBy[coredata.DevicePostureOrderField]{
				Field:     coredata.DevicePostureOrderFieldCheckKey,
				Direction: page.OrderDirectionAsc,
			},
			func(ctx context.Context, cursor *page.Cursor[coredata.DevicePostureOrderField]) ([]*coredata.DevicePosture, error) {
				var batch coredata.DevicePostures
				if err := batch.LoadLatestByDeviceID(ctx, conn, fx.scope, fx.deviceID, cursor); err != nil {
					return nil, fmt.Errorf("cannot load latest device postures: %w", err)
				}

				return batch, nil
			},
		)
		if err != nil {
			return err
		}

		all = loaded

		return nil
	}))
	require.Len(t, all, 3)

	byKey := make(map[string]coredata.DevicePostureStatus, len(all))
	for _, posture := range all {
		byKey[posture.CheckKey] = posture.Status
	}

	assert.Equal(t, coredata.DevicePostureStatusPass, byKey["AUTO_UPDATE"])
	assert.Equal(t, coredata.DevicePostureStatusUnknown, byKey["DISK_ENCRYPTION"])
	assert.Equal(t, coredata.DevicePostureStatusFail, byKey["FIREWALL_ENABLED"])
}
