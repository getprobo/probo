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
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/itam/mdm"
	"go.probo.inc/probo/pkg/mail"
)

type (
	// DeviceSyncResult counts what one sync changed.
	DeviceSyncResult struct {
		Created  int
		Updated  int
		Linked   int
		Revoked  int64
		Unlinked int64
	}
)

// ApplyDeviceSync makes the organization's device list match what an MDM
// connector lists, in one transaction:
//
//   - A device that matches a live agent device by serial number is linked to
//     that row, so one machine stays one row. An MDM row the device had
//     before the agent enrolled is retired.
//   - Any other device gets, or keeps, an MDM row of its own. A revoked MDM
//     row is left revoked: a sync never re-activates a device an admin or a
//     previous sync revoked.
//   - A device without an owner gets the active member whose email matches
//     the MDM's assigned user. An owner already set is never changed.
//   - MDM rows the connector no longer lists are revoked, and agent rows it
//     no longer lists lose their link.
//
// When claim is not nil, the claim is locked and checked first, in the same
// transaction, so only the worker that holds the sync writes its devices,
// and nothing fetched with credentials the connector has since replaced is
// written. It then returns coredata.ErrResourceNotFound or
// coredata.ErrDeviceSyncSuperseded and changes nothing.
func (s *Service) ApplyDeviceSync(
	ctx context.Context,
	claim *coredata.DeviceSync,
	connector *coredata.Connector,
	records []mdm.DeviceRecord,
) (*DeviceSyncResult, error) {
	scope := coredata.NewScopeFromObjectID(connector.ID)
	result := &DeviceSyncResult{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if claim != nil {
				if err := claim.LockClaim(ctx, tx, scope); err != nil {
					return err
				}
			}

			listedIDs := make([]string, 0, len(records))
			seen := make(map[string]struct{}, len(records))

			for _, record := range records {
				// A provider listing one device twice would otherwise be
				// written twice, and the second write could undo the first.
				if _, dup := seen[record.ExternalID]; dup {
					continue
				}

				seen[record.ExternalID] = struct{}{}
				listedIDs = append(listedIDs, record.ExternalID)

				if err := s.applyDeviceRecord(ctx, tx, scope, connector, record, result); err != nil {
					return err
				}
			}

			var devices coredata.Devices

			revoked, err := devices.RevokeUnlistedMDM(ctx, tx, scope, connector.ID, listedIDs)
			if err != nil {
				return fmt.Errorf("cannot revoke unlisted devices: %w", err)
			}

			unlinked, err := devices.UnlinkUnlistedMDM(ctx, tx, scope, connector.ID, listedIDs)
			if err != nil {
				return fmt.Errorf("cannot unlink unlisted devices: %w", err)
			}

			result.Revoked = revoked
			result.Unlinked = unlinked

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) applyDeviceRecord(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connector *coredata.Connector,
	record mdm.DeviceRecord,
	result *DeviceSyncResult,
) error {
	var linked *coredata.Device

	existing := &coredata.Device{}
	if err := existing.LoadByMDMDeviceIDForUpdate(ctx, tx, scope, connector.ID, record.ExternalID); err != nil {
		if !errors.Is(err, coredata.ErrResourceNotFound) {
			return fmt.Errorf("cannot load synced device: %w", err)
		}
	} else {
		linked = existing
	}

	agent, err := s.loadAgentDeviceForRecord(ctx, tx, scope, connector.OrganizationID, record)
	if err != nil {
		return err
	}

	if agent != nil {
		return s.linkAgentDevice(ctx, tx, scope, connector, record, agent, linked, result)
	}

	if linked != nil && linked.Source == coredata.DeviceSourceAgent {
		// The agent row this device was linked to no longer matches it: it
		// was revoked, deleted or reports another serial. The link moves to a
		// row of the device's own.
		if err := linked.UnlinkMDM(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot unlink agent device: %w", err)
		}

		linked = nil
	}

	if linked == nil {
		if err := s.insertMDMDevice(ctx, tx, scope, connector, record); err != nil {
			return err
		}

		result.Created++

		return nil
	}

	if linked.State == coredata.DeviceStateRevoked {
		return nil
	}

	// A field the provider leaves out this time keeps what an earlier sync
	// stored: losing a serial would stop a later agent from merging with
	// this row, and losing the check-in would reset the last-seen time.
	linked.Hostname = keepUnlessSet(linked.Hostname, record.Name)
	linked.SerialNumber = keepUnlessSet(linked.SerialNumber, record.SerialNumber)
	linked.Platform = keepUnlessSet(linked.Platform, record.Platform)
	linked.OSVersion = keepUnlessSet(linked.OSVersion, record.OSVersion)
	linked.Model = keepUnlessSet(linked.Model, record.Model)
	linked.MDMLastCheckInAt = keepUnlessSet(linked.MDMLastCheckInAt, record.LastCheckInAt)

	if err := linked.UpdateFromMDM(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot update synced device: %w", err)
	}

	if err := s.assignOwnerFromMDM(ctx, tx, scope, linked, record); err != nil {
		return err
	}

	result.Updated++

	return nil
}

func (s *Service) loadAgentDeviceForRecord(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	record mdm.DeviceRecord,
) (*coredata.Device, error) {
	if record.SerialNumber == nil {
		return nil, nil
	}

	agent := &coredata.Device{}
	if err := agent.LoadAgentBySerialNumberForUpdate(ctx, tx, scope, organizationID, *record.SerialNumber); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load agent device by serial number: %w", err)
	}

	return agent, nil
}

func (s *Service) linkAgentDevice(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connector *coredata.Connector,
	record mdm.DeviceRecord,
	agent *coredata.Device,
	linked *coredata.Device,
	result *DeviceSyncResult,
) error {
	alreadyLinked := linked != nil && linked.ID == agent.ID

	if linked != nil && !alreadyLinked {
		// The device was synced before the agent enrolled on it, or was
		// linked to another agent row. Free the link first: it is unique.
		switch linked.Source {
		case coredata.DeviceSourceMDM:
			if err := retireMDMDevice(ctx, tx, scope, linked); err != nil {
				return err
			}
		case coredata.DeviceSourceAgent:
			if err := linked.UnlinkMDM(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot unlink agent device: %w", err)
			}
		}
	}

	if err := agent.LinkMDM(
		ctx,
		tx,
		scope,
		connector.Provider,
		connector.ID,
		record.ExternalID,
		record.Model,
		record.LastCheckInAt,
	); err != nil {
		return fmt.Errorf("cannot link agent device: %w", err)
	}

	if err := s.assignOwnerFromMDM(ctx, tx, scope, agent, record); err != nil {
		return err
	}

	if alreadyLinked {
		result.Updated++
	} else {
		result.Linked++
	}

	return nil
}

// retireMDMDevice revokes and soft-deletes an MDM row whose device is now an
// agent row. It is the same machine, so the row is a duplicate, not a device
// that left the MDM.
func retireMDMDevice(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	device *coredata.Device,
) error {
	if device.State != coredata.DeviceStateRevoked {
		if err := device.Revoke(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot revoke duplicate synced device: %w", err)
		}
	}

	if err := device.SoftDelete(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot delete duplicate synced device: %w", err)
	}

	return nil
}

func (s *Service) insertMDMDevice(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	connector *coredata.Connector,
	record mdm.DeviceRecord,
) error {
	now := time.Now()
	provider := connector.Provider
	connectorID := connector.ID
	externalID := record.ExternalID

	device := &coredata.Device{
		ID:               gid.New(connector.OrganizationID.TenantID(), coredata.DeviceEntityType),
		OrganizationID:   connector.OrganizationID,
		State:            coredata.DeviceStateActive,
		Source:           coredata.DeviceSourceMDM,
		Hostname:         record.Name,
		SerialNumber:     record.SerialNumber,
		Platform:         record.Platform,
		OSVersion:        record.OSVersion,
		Model:            record.Model,
		LastSeenAt:       record.LastCheckInAt,
		MDMProvider:      &provider,
		MDMConnectorID:   &connectorID,
		MDMDeviceID:      &externalID,
		MDMLastCheckInAt: record.LastCheckInAt,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	ownerID, err := s.ownerFromMDM(ctx, tx, scope, connector.OrganizationID, record)
	if err != nil {
		return err
	}

	device.OwnerID = ownerID

	if err := device.Insert(ctx, tx, scope); err != nil {
		return fmt.Errorf("cannot insert synced device: %w", err)
	}

	return nil
}

func (s *Service) assignOwnerFromMDM(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	device *coredata.Device,
	record mdm.DeviceRecord,
) error {
	if device.OwnerID != nil {
		return nil
	}

	ownerID, err := s.ownerFromMDM(ctx, tx, scope, device.OrganizationID, record)
	if err != nil || ownerID == nil {
		return err
	}

	if err := device.AssignOwner(ctx, tx, scope, ownerID); err != nil {
		return fmt.Errorf("cannot assign synced device owner: %w", err)
	}

	return nil
}

// ownerFromMDM resolves the MDM's assigned user to an active member of the
// organization by email. No match, or more than one, leaves the device
// without an owner rather than guessing.
func (s *Service) ownerFromMDM(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	organizationID gid.GID,
	record mdm.DeviceRecord,
) (*gid.GID, error) {
	if record.UserEmail == nil {
		return nil, nil
	}

	email, err := mail.ParseAddr(*record.UserEmail)
	if err != nil {
		return nil, nil
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByOrganizationIDAndEmail(ctx, tx, scope, organizationID, email); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("cannot load owner profile by email: %w", err)
	}

	return &profile.ID, nil
}

// keepUnlessSet returns next when the provider reported it, and current
// otherwise.
func keepUnlessSet[T any](current, next *T) *T {
	if next != nil {
		return next
	}

	return current
}
