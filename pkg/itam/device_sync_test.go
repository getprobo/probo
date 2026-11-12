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

package itam

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/itam/mdm"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/page"
)

type deviceSyncFixture struct {
	t              *testing.T
	svc            *Service
	scope          coredata.Scoper
	organizationID gid.GID
	connector      *coredata.Connector
}

// newDeviceSyncFixture creates an organization with an Iru connector in a
// tenant of its own, so parallel subtests never see each other's devices.
func newDeviceSyncFixture(t *testing.T) *deviceSyncFixture {
	t.Helper()

	client := test.PGClient(t)
	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := gid.NewTenantID()
	scope := coredata.NewScope(tenantID, gid.Nil)

	f := &deviceSyncFixture{
		t:              t,
		svc:            &Service{pg: client, logger: log.NewLogger(log.WithName("test"))},
		scope:          scope,
		organizationID: gid.New(tenantID, coredata.OrganizationEntityType),
	}

	f.connector = &coredata.Connector{
		ID:             gid.New(tenantID, coredata.ConnectorEntityType),
		OrganizationID: f.organizationID,
		Provider:       coredata.ConnectorProviderIru,
		Protocol:       coredata.ConnectorProtocolAPIKey,
		Connection:     &connectorprotocol.APIKeyConnection{APIKey: "test-token"},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(t, client.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
		org := &coredata.Organization{
			ID:        f.organizationID,
			TenantID:  tenantID,
			Name:      "Device Sync Org",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := org.Insert(ctx, tx); err != nil {
			return err
		}

		if err := f.connector.SetSettings(coredata.IruConnectorSettings{APIHost: "acme.api.kandji.io"}); err != nil {
			return err
		}

		var key cipher.EncryptionKey

		return f.connector.Insert(ctx, tx, scope, key)
	}))

	return f
}

// member creates an active member of the organization and returns its
// profile ID. The email is made unique per tenant, as identities are global.
func (f *deviceSyncFixture) member(localPart string) (gid.GID, string) {
	f.t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	tenantID := f.organizationID.TenantID()
	email := fmt.Sprintf("%s.%s@example.com", localPart, tenantID)
	profileID := gid.New(tenantID, coredata.MembershipProfileEntityType)

	require.NoError(f.t, f.svc.pg.WithTx(f.t.Context(), func(ctx context.Context, tx pg.Tx) error {
		identity := coredata.Identity{
			ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
			EmailAddress:         mail.Addr(email),
			FullName:             localPart,
			EmailAddressVerified: true,
			CreatedAt:            now,
			UpdatedAt:            now,
		}
		if err := identity.Insert(ctx, tx); err != nil {
			return err
		}

		membership := coredata.Membership{
			ID:             gid.New(tenantID, coredata.MembershipEntityType),
			IdentityID:     identity.ID,
			OrganizationID: f.organizationID,
			Role:           coredata.MembershipRoleEmployee,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := membership.Insert(ctx, tx, f.scope); err != nil {
			return err
		}

		profile := coredata.MembershipProfile{
			ID:             profileID,
			IdentityID:     identity.ID,
			OrganizationID: f.organizationID,
			Source:         coredata.ProfileSourceManual,
			State:          coredata.ProfileStateActive,
			FullName:       localPart,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		return profile.Insert(ctx, tx)
	}))

	return profileID, email
}

// agentDevice enrolls an ACTIVE agent device reporting serialNumber.
func (f *deviceSyncFixture) agentDevice(organizationID gid.GID, serialNumber string, ownerID *gid.GID) gid.GID {
	f.t.Helper()

	now := time.Now().UTC().Truncate(time.Microsecond)
	apiKeyHash := []byte(rand.MustHexString(32))

	platform := coredata.DevicePlatformDarwin
	device := coredata.Device{
		ID:             gid.New(organizationID.TenantID(), coredata.DeviceEntityType),
		OrganizationID: organizationID,
		State:          coredata.DeviceStateActive,
		Source:         coredata.DeviceSourceAgent,
		HardwareUUID:   new(gid.New(organizationID.TenantID(), coredata.DeviceEntityType).String()),
		SerialNumber:   &serialNumber,
		Hostname:       new("agent-host"),
		Platform:       &platform,
		OSVersion:      new("15.1"),
		AgentVersion:   new("1.0.0"),
		APIKeyHash:     apiKeyHash,
		OwnerID:        ownerID,
		EnrolledAt:     &now,
		LastSeenAt:     &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	require.NoError(f.t, f.svc.pg.WithTx(f.t.Context(), func(ctx context.Context, tx pg.Tx) error {
		return device.Insert(ctx, tx, coredata.NewScopeFromObjectID(organizationID))
	}))

	return device.ID
}

func (f *deviceSyncFixture) sync(records ...mdm.DeviceRecord) *DeviceSyncResult {
	f.t.Helper()

	result, err := f.svc.ApplyDeviceSync(f.t.Context(), f.connector, records)
	require.NoError(f.t, err)

	return result
}

// devices lists the organization's live devices, oldest first.
func (f *deviceSyncFixture) devices() coredata.Devices {
	f.t.Helper()

	var devices coredata.Devices

	cursor := page.NewCursor(
		100,
		nil,
		page.Head,
		page.OrderBy[coredata.DeviceOrderField]{
			Field:     coredata.DeviceOrderFieldCreatedAt,
			Direction: page.OrderDirectionAsc,
		},
	)

	require.NoError(f.t, f.svc.pg.WithConn(f.t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return devices.LoadByOrganizationID(
			ctx,
			conn,
			f.scope,
			f.organizationID,
			cursor,
			coredata.NewDeviceFilter(nil, nil, nil, nil, nil, nil),
		)
	}))

	return devices
}

func (f *deviceSyncFixture) device(id gid.GID) *coredata.Device {
	f.t.Helper()

	device := &coredata.Device{}

	require.NoError(f.t, f.svc.pg.WithConn(f.t.Context(), func(ctx context.Context, conn pg.Querier) error {
		return device.LoadByID(ctx, conn, f.scope, id)
	}))

	return device
}

func macRecord(id, serialNumber string, email *string) mdm.DeviceRecord {
	platform := coredata.DevicePlatformDarwin
	checkIn := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

	return mdm.DeviceRecord{
		ExternalID:    id,
		Name:          new("Mac " + id),
		SerialNumber:  &serialNumber,
		Platform:      &platform,
		Model:         new("MacBook Air (M2, 2022)"),
		OSVersion:     new("15.1"),
		UserEmail:     email,
		LastCheckInAt: &checkIn,
	}
}

func TestService_ApplyDeviceSync(t *testing.T) {
	t.Parallel()

	t.Run(
		"creates a device per listed device and updates it on the next sync",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			ownerID, email := f.member("ada")

			result := f.sync(macRecord("iru-1", "SERIAL-1", &email))
			assert.Equal(t, 1, result.Created)

			devices := f.devices()
			require.Len(t, devices, 1)

			device := devices[0]
			assert.Equal(t, coredata.DeviceSourceMDM, device.Source)
			assert.Equal(t, coredata.DeviceStateActive, device.State)
			assert.Nil(t, device.APIKeyHash)
			require.NotNil(t, device.MDMProvider)
			assert.Equal(t, coredata.ConnectorProviderIru, *device.MDMProvider)
			require.NotNil(t, device.MDMDeviceID)
			assert.Equal(t, "iru-1", *device.MDMDeviceID)
			require.NotNil(t, device.Model)
			assert.Equal(t, "MacBook Air (M2, 2022)", *device.Model)
			require.NotNil(t, device.OwnerID)
			assert.Equal(t, ownerID, *device.OwnerID)
			require.NotNil(t, device.LastSeenAt)

			updated := macRecord("iru-1", "SERIAL-1", &email)
			updated.OSVersion = new("15.2")

			result = f.sync(updated)
			assert.Equal(t, 0, result.Created)
			assert.Equal(t, 1, result.Updated)

			devices = f.devices()
			require.Len(t, devices, 1)
			require.NotNil(t, devices[0].OSVersion)
			assert.Equal(t, "15.2", *devices[0].OSVersion)
		},
	)

	t.Run(
		"links an agent device with the same serial instead of adding one",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			agentID := f.agentDevice(f.organizationID, "SERIAL-A", nil)
			ownerID, email := f.member("grace")

			result := f.sync(macRecord("iru-a", "SERIAL-A", &email))
			assert.Equal(t, 1, result.Linked)
			assert.Equal(t, 0, result.Created)

			devices := f.devices()
			require.Len(t, devices, 1)

			agent := devices[0]
			assert.Equal(t, agentID, agent.ID)
			assert.Equal(t, coredata.DeviceSourceAgent, agent.Source)
			require.NotNil(t, agent.MDMDeviceID)
			assert.Equal(t, "iru-a", *agent.MDMDeviceID)
			require.NotNil(t, agent.MDMLastCheckInAt)
			// The agent keeps reporting its own hostname.
			require.NotNil(t, agent.Hostname)
			assert.Equal(t, "agent-host", *agent.Hostname)
			require.NotNil(t, agent.OwnerID)
			assert.Equal(t, ownerID, *agent.OwnerID)
		},
	)

	t.Run(
		"merges a synced device into the agent that enrolls on it later",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)

			f.sync(macRecord("iru-b", "SERIAL-B", nil))
			require.Len(t, f.devices(), 1)

			agentID := f.agentDevice(f.organizationID, "SERIAL-B", nil)

			result := f.sync(macRecord("iru-b", "SERIAL-B", nil))
			assert.Equal(t, 1, result.Linked)

			devices := f.devices()
			require.Len(t, devices, 1, "the synced duplicate must be retired")
			assert.Equal(t, agentID, devices[0].ID)
			require.NotNil(t, devices[0].MDMDeviceID)
			assert.Equal(t, "iru-b", *devices[0].MDMDeviceID)
		},
	)

	t.Run(
		"revokes devices no longer listed and never re-activates them",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			agentID := f.agentDevice(f.organizationID, "SERIAL-C", nil)

			f.sync(
				macRecord("iru-c", "SERIAL-C", nil),
				macRecord("iru-d", "SERIAL-D", nil),
			)

			devices := f.devices()
			require.Len(t, devices, 2)

			var syncedID gid.GID

			for _, d := range devices {
				if d.Source == coredata.DeviceSourceMDM {
					syncedID = d.ID
				}
			}

			require.False(t, syncedID == gid.Nil)

			result := f.sync()
			assert.Equal(t, int64(1), result.Revoked)
			assert.Equal(t, int64(1), result.Unlinked)

			synced := f.device(syncedID)
			assert.Equal(t, coredata.DeviceStateRevoked, synced.State)
			assert.NotNil(t, synced.RevokedAt)

			agent := f.device(agentID)
			assert.Equal(t, coredata.DeviceStateActive, agent.State, "an agent device is unlinked, never revoked")
			assert.Nil(t, agent.MDMConnectorID)
			assert.Nil(t, agent.MDMDeviceID)
			assert.Nil(t, agent.MDMProvider)

			f.sync(macRecord("iru-d", "SERIAL-D", nil))
			assert.Equal(t, coredata.DeviceStateRevoked, f.device(syncedID).State)
		},
	)

	t.Run(
		"keeps an owner already set and leaves unknown users unassigned",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			setOwnerID, _ := f.member("linus")
			_, otherEmail := f.member("margaret")
			agentID := f.agentDevice(f.organizationID, "SERIAL-E", &setOwnerID)
			stranger := "stranger@example.com"

			f.sync(
				macRecord("iru-e", "SERIAL-E", &otherEmail),
				macRecord("iru-f", "SERIAL-F", &stranger),
			)

			agent := f.device(agentID)
			require.NotNil(t, agent.OwnerID)
			assert.Equal(t, setOwnerID, *agent.OwnerID)

			for _, d := range f.devices() {
				if d.Source == coredata.DeviceSourceMDM {
					assert.Nil(t, d.OwnerID)
				}
			}
		},
	)

	t.Run(
		"never links an agent device of another organization",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			other := newDeviceSyncFixture(t)
			otherAgentID := other.agentDevice(other.organizationID, "SERIAL-G", nil)

			result := f.sync(macRecord("iru-g", "SERIAL-G", nil))
			assert.Equal(t, 1, result.Created)
			assert.Equal(t, 0, result.Linked)

			assert.Nil(t, other.device(otherAgentID).MDMDeviceID)
		},
	)

	t.Run(
		"revokes synced devices once their connector is deleted",
		func(t *testing.T) {
			t.Parallel()

			f := newDeviceSyncFixture(t)
			f.sync(macRecord("iru-h", "SERIAL-H", nil))

			require.NoError(t, f.svc.pg.WithTx(t.Context(), func(ctx context.Context, tx pg.Tx) error {
				if err := f.connector.Delete(ctx, tx, f.scope); err != nil {
					return err
				}

				var device coredata.Device

				_, err := device.RevokeDetachedMDM(ctx, tx, time.Now())

				return err
			}))

			devices := f.devices()
			require.Len(t, devices, 1)
			assert.Equal(t, coredata.DeviceStateRevoked, devices[0].State)
			assert.Nil(t, devices[0].MDMConnectorID)
			require.NotNil(t, devices[0].MDMProvider, "the device still names its MDM")
		},
	)
}
