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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// insertMDMFixture creates an organization with an Iru connector and a
// revoked MDM device owned by ownerID.
func insertMDMFixture(
	t *testing.T,
	client *pg.Client,
) (coredata.Scoper, gid.GID, gid.GID, gid.GID) {
	t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)
	organizationID := gid.New(tenantID, coredata.OrganizationEntityType)
	connectorID := gid.New(tenantID, coredata.ConnectorEntityType)
	deviceID := gid.New(tenantID, coredata.DeviceEntityType)

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        organizationID,
			TenantID:  tenantID,
			Name:      "MDM Devices Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		connector := &coredata.Connector{
			ID:             connectorID,
			OrganizationID: organizationID,
			Provider:       coredata.ConnectorProviderIru,
			Protocol:       coredata.ConnectorProtocolAPIKey,
			Connection:     &connectorprotocol.APIKeyConnection{APIKey: "test-token"},
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		var key cipher.EncryptionKey
		if err := connector.Insert(ctx, tx, scope, key); err != nil {
			return err
		}

		provider := coredata.ConnectorProviderIru
		device := coredata.Device{
			ID:             deviceID,
			OrganizationID: organizationID,
			State:          coredata.DeviceStateRevoked,
			Source:         coredata.DeviceSourceMDM,
			MDMProvider:    &provider,
			MDMConnectorID: &connectorID,
			MDMDeviceID:    new("iru-device"),
			RevokedAt:      &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return device.Insert(ctx, tx, scope)
	}))

	return scope, organizationID, connectorID, deviceID
}

func TestDevice_DeleteOrphans_KeepsMDMDevices(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	scope, _, _, deviceID := insertMDMFixture(t, client)

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		var device coredata.Device

		_, err := device.DeleteOrphans(ctx, tx, time.Now())

		return err
	}))

	// A revoked MDM device has no API key and no postures, which is exactly
	// what an agent orphan looks like. It must stay listed until an admin
	// deletes it.
	device := &coredata.Device{}

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return device.LoadByID(ctx, conn, scope, deviceID)
	}))
	assert.Equal(t, coredata.DeviceSourceMDM, device.Source)
}

func TestDevice_MDMRowConstraints(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	scope, organizationID, connectorID, _ := insertMDMFixture(t, client)
	now := time.Now().UTC().Truncate(time.Microsecond)
	provider := coredata.ConnectorProviderIru

	insert := func(device coredata.Device) error {
		return client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
			return device.Insert(ctx, tx, scope)
		})
	}

	base := func() coredata.Device {
		return coredata.Device{
			ID:             gid.New(organizationID.TenantID(), coredata.DeviceEntityType),
			OrganizationID: organizationID,
			State:          coredata.DeviceStateActive,
			Source:         coredata.DeviceSourceMDM,
			MDMProvider:    &provider,
			MDMConnectorID: &connectorID,
			MDMDeviceID:    new("iru-" + gid.New(organizationID.TenantID(), coredata.DeviceEntityType).String()),
			CreatedAt:      now,
			UpdatedAt:      now,
		}
	}

	// An ACTIVE MDM row needs none of the agent's fields.
	require.NoError(t, insert(base()))

	// An MDM row can never hold an agent credential.
	withKey := base()
	withKey.APIKeyHash = []byte("key-" + withKey.ID.String())
	assert.Error(t, insert(withKey))

	// An MDM row must name the device its MDM lists.
	withoutDeviceID := base()
	withoutDeviceID.MDMDeviceID = nil
	assert.Error(t, insert(withoutDeviceID))

	// The agent invariants still hold for agent rows.
	agent := base()
	agent.Source = coredata.DeviceSourceAgent
	agent.MDMProvider = nil
	agent.MDMConnectorID = nil
	agent.MDMDeviceID = nil
	assert.Error(t, insert(agent))
}

func TestDevices_LoadByOrganizationIDAndOwnerID_ExcludesMDMDevices(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	scope, organizationID, connectorID, _ := insertMDMFixture(t, client)
	now := time.Now().UTC().Truncate(time.Microsecond)
	provider := coredata.ConnectorProviderIru
	ownerID := gid.New(organizationID.TenantID(), coredata.MembershipProfileEntityType)

	// The owner column references a profile, so the MDM row is inserted
	// without one and the query is checked to filter on source rather than
	// on the owner alone.
	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		device := coredata.Device{
			ID:             gid.New(organizationID.TenantID(), coredata.DeviceEntityType),
			OrganizationID: organizationID,
			State:          coredata.DeviceStateActive,
			Source:         coredata.DeviceSourceMDM,
			MDMProvider:    &provider,
			MDMConnectorID: &connectorID,
			MDMDeviceID:    new("iru-active"),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return device.Insert(ctx, tx, scope)
	}))

	var devices coredata.Devices

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		count, err := devices.CountByOrganizationIDAndOwnerID(ctx, conn, scope, organizationID, ownerID)
		if err != nil {
			return err
		}

		assert.Equal(t, 0, count)

		return devices.LoadByOrganizationIDAndOwnerID(
			ctx,
			conn,
			scope,
			organizationID,
			ownerID,
			page.NewCursor(
				10,
				nil,
				page.Head,
				page.OrderBy[coredata.DeviceOrderField]{
					Field:     coredata.DeviceOrderFieldCreatedAt,
					Direction: page.OrderDirectionAsc,
				},
			),
		)
	}))
	assert.Empty(t, devices)

	sources := []coredata.DeviceSource{coredata.DeviceSourceMDM}

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		count, err := devices.CountByOrganizationID(
			ctx,
			conn,
			scope,
			organizationID,
			coredata.NewDeviceFilter(nil, nil, sources, nil, nil, nil),
		)
		if err != nil {
			return err
		}

		assert.Equal(t, 2, count)

		agentOnly, err := devices.CountByOrganizationID(
			ctx,
			conn,
			scope,
			organizationID,
			coredata.NewDeviceFilter(nil, nil, []coredata.DeviceSource{coredata.DeviceSourceAgent}, nil, nil, nil),
		)
		if err != nil {
			return err
		}

		assert.Equal(t, 0, agentOnly)

		return nil
	}))
}

func TestDeviceSync_Schedule(t *testing.T) {
	t.Parallel()

	client := test.PGClient(t)
	scope, organizationID, connectorID, _ := insertMDMFixture(t, client)
	now := time.Now().UTC().Truncate(time.Microsecond)

	sync := &coredata.DeviceSync{
		ConnectorID:    connectorID,
		OrganizationID: organizationID,
		NextSyncAt:     now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		if err := sync.Schedule(ctx, tx, scope); err != nil {
			return err
		}

		return sync.RecordAttempt(ctx, tx, scope, now, now.Add(time.Hour))
	}))

	loaded := &coredata.DeviceSync{}

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByConnectorID(ctx, conn, scope, connectorID)
	}))
	assert.Equal(t, 1, loaded.Attempts)
	assert.True(t, loaded.NextSyncAt.After(now))

	// Scheduling again, as a reconnect does, resets the budget and makes the
	// sync due at once.
	sync.NextSyncAt = now
	sync.UpdatedAt = now

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return sync.Schedule(ctx, tx, scope)
	}))

	require.NoError(t, client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByConnectorID(ctx, conn, scope, connectorID)
	}))
	assert.Equal(t, 0, loaded.Attempts)
	assert.True(t, loaded.NextSyncAt.Equal(now))

	// Deleting the connector deletes its schedule.
	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		connector := &coredata.Connector{ID: connectorID}

		return connector.Delete(ctx, tx, scope)
	}))

	err := client.WithConn(t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return loaded.LoadByConnectorID(ctx, conn, scope, connectorID)
	})
	assert.True(t, errors.Is(err, coredata.ErrResourceNotFound))
}
