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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestDevice_DeleteOrphans(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)

	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)

	orphanPendingID := gid.New(tenantID, coredata.DeviceEntityType)
	validTokenID := gid.New(tenantID, coredata.DeviceEntityType)
	withKeyID := gid.New(tenantID, coredata.DeviceEntityType)
	orphanRevokedID := gid.New(tenantID, coredata.DeviceEntityType)
	softDeletedOrphanID := gid.New(tenantID, coredata.DeviceEntityType)
	softDeletedWithPostureID := gid.New(tenantID, coredata.DeviceEntityType)

	deviceIDs := []string{
		orphanPendingID.String(),
		validTokenID.String(),
		withKeyID.String(),
		orphanRevokedID.String(),
		softDeletedOrphanID.String(),
		softDeletedWithPostureID.String(),
	}

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "Orphan Devices GC Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		insertPending := func(id gid.GID, apiKeyHash []byte) error {
			device := coredata.Device{
				ID:             id,
				OrganizationID: organizationID,
				State:          coredata.DeviceStatePending,
				APIKeyHash:     apiKeyHash,
				CreatedAt:      now,
				UpdatedAt:      now,
			}

			return device.Insert(ctx, tx, scope)
		}

		if err := insertPending(orphanPendingID, nil); err != nil {
			return err
		}

		if err := insertPending(validTokenID, nil); err != nil {
			return err
		}

		if err := insertPending(withKeyID, []byte("orphan-gc-key-"+withKeyID.String())); err != nil {
			return err
		}

		revokedAt := now

		orphanRevoked := coredata.Device{
			ID:             orphanRevokedID,
			OrganizationID: organizationID,
			State:          coredata.DeviceStateRevoked,
			RevokedAt:      &revokedAt,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := orphanRevoked.Insert(ctx, tx, scope); err != nil {
			return err
		}

		softDeletedOrphan := coredata.Device{
			ID:             softDeletedOrphanID,
			OrganizationID: organizationID,
			State:          coredata.DeviceStateRevoked,
			RevokedAt:      &revokedAt,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := softDeletedOrphan.Insert(ctx, tx, scope); err != nil {
			return err
		}

		if err := softDeletedOrphan.SoftDelete(ctx, tx, scope); err != nil {
			return err
		}

		softDeletedWithPosture := coredata.Device{
			ID:             softDeletedWithPostureID,
			OrganizationID: organizationID,
			State:          coredata.DeviceStateRevoked,
			RevokedAt:      &revokedAt,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := softDeletedWithPosture.Insert(ctx, tx, scope); err != nil {
			return err
		}

		if err := softDeletedWithPosture.SoftDelete(ctx, tx, scope); err != nil {
			return err
		}

		posture := coredata.DevicePosture{
			ID:             gid.New(tenantID, coredata.DevicePostureEntityType),
			OrganizationID: organizationID,
			DeviceID:       softDeletedWithPostureID,
			CorrelationID:  gid.New(tenantID, coredata.DevicePostureReportEntityType),
			CheckKey:       "DISK_ENCRYPTION",
			Status:         coredata.DevicePostureStatusPass,
			ObservedAt:     now,
			CreatedAt:      now,
		}
		if err := posture.Insert(ctx, tx, scope); err != nil {
			return err
		}

		token := coredata.DeviceEnrollmentToken{
			ID:          gid.New(tenantID, coredata.DeviceEnrollmentTokenEntityType),
			DeviceID:    validTokenID,
			HashedValue: []byte("orphan-gc-token-" + validTokenID.String()),
			ExpiresAt:   now.Add(time.Hour),
			CreatedAt:   now,
		}

		return token.Insert(ctx, tx, scope)
	}))

	t.Cleanup(func() {
		_ = client.WithTx(context.Background(), func(ctx context.Context, tx pg.Tx) error {
			_, _ = tx.Exec(ctx, `DELETE FROM device_postures WHERE device_id = ANY($1)`, deviceIDs)
			_, _ = tx.Exec(ctx, `DELETE FROM device_enrollment_tokens WHERE device_id = ANY($1)`, deviceIDs)
			_, _ = tx.Exec(ctx, `DELETE FROM devices WHERE id = ANY($1)`, deviceIDs)
			_, _ = tx.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, organizationID)

			return nil
		})
	})

	require.NoError(t, client.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var device coredata.Device

		deleted, err := device.DeleteOrphans(ctx, tx, now)
		require.NoError(t, err)
		require.Equal(t, int64(3), deleted)

		var remaining coredata.Device
		require.ErrorIs(t, remaining.LoadByID(ctx, tx, scope, orphanPendingID), coredata.ErrResourceNotFound)
		require.ErrorIs(t, remaining.LoadByID(ctx, tx, scope, orphanRevokedID), coredata.ErrResourceNotFound)

		require.NoError(t, remaining.LoadByID(ctx, tx, scope, validTokenID))
		require.NoError(t, remaining.LoadByID(ctx, tx, scope, withKeyID))

		var softDeletedOrphanCount int

		err = tx.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM devices WHERE id = $1`,
			softDeletedOrphanID,
		).Scan(&softDeletedOrphanCount)
		require.NoError(t, err)
		require.Equal(t, 0, softDeletedOrphanCount)

		var softDeletedWithHistory int

		err = tx.QueryRow(
			ctx,
			`SELECT COUNT(*) FROM devices WHERE id = $1 AND deleted_at IS NOT NULL`,
			softDeletedWithPostureID,
		).Scan(&softDeletedWithHistory)
		require.NoError(t, err)
		require.Equal(t, 1, softDeletedWithHistory)

		return nil
	}))
}
