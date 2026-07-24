// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/page"
)

var (
	// ErrDeviceRevoked is returned when the authenticated device has
	// been revoked.
	ErrDeviceRevoked = errors.New("device is revoked")

	// ErrDeviceHardwareConflict is returned when activation would
	// duplicate an existing (organization_id, hardware_uuid) pair.
	ErrDeviceHardwareConflict = errors.New("device hardware uuid already enrolled")

	// ErrEnrollmentTokenExpired is returned when an enrollment token
	// has passed its expiry time.
	ErrEnrollmentTokenExpired = errors.New("enrollment token expired")

	// ErrEnrollmentTokenAlreadyUsed is returned when an enrollment
	// token has already been exchanged.
	ErrEnrollmentTokenAlreadyUsed = errors.New("enrollment token already used")

	// ErrEnrollmentTokenInvalid is returned when an enrollment token
	// cannot be exchanged for the device.
	ErrEnrollmentTokenInvalid = errors.New("enrollment token invalid")
)

const (
	// APIKeyRawLength is the random byte length of a device API key
	// secret (96 chars once hex-encoded).
	APIKeyRawLength = 48

	// EnrollmentTokenRawLength is the random byte length of a device
	// enrollment token secret (96 chars once hex-encoded).
	EnrollmentTokenRawLength = 48
)

type (
	// Service is the IT Asset Management service. Admin operations are
	// tenant-scoped via a caller-supplied scope; agent-facing operations
	// (authenticate, heartbeat, postures, unenroll) resolve their own
	// scope, since the agent does not know its tenant until activation.
	Service struct {
		pg                      *pg.Client
		logger                  *log.Logger
		enrollmentTokenValidity time.Duration
	}

	CreateDeviceRequest struct {
		OrganizationID gid.GID
		OwnerID        *gid.GID
	}

	EnrollDeviceRequest struct {
		OrganizationID gid.GID
		IdentityID     gid.GID
	}

	// CreateDeviceResult carries the device row and the plaintext
	// enrollment token the agent installer must exchange for an API key.
	// Only the hash is stored, so EnrollmentToken is available only at
	// this point.
	CreateDeviceResult struct {
		Device          *coredata.Device
		EnrollmentToken string
	}

	RecordHeartbeatRequest struct {
		HardwareUUID string
		SerialNumber *string
		Hostname     string
		Platform     coredata.DevicePlatform
		OSVersion    string
		AgentVersion string
	}

	RecordPostureResult struct {
		CheckKey   string
		Status     coredata.DevicePostureStatus
		Evidence   json.RawMessage
		ObservedAt time.Time
	}

	ServiceConfig struct {
		EnrollmentTokenValidity time.Duration
	}
)

func NewService(
	pgClient *pg.Client,
	iamSvc *iam.Service,
	cfg ServiceConfig,
	logger *log.Logger,
) *Service {
	iamSvc.Authorizer.RegisterPolicySet(ITAMPolicySet())

	validity := cfg.EnrollmentTokenValidity
	if validity <= 0 {
		validity = 7 * 24 * time.Hour
	}

	return &Service{
		pg:                      pgClient,
		logger:                  logger,
		enrollmentTokenValidity: validity,
	}
}

func (s *Service) CreateDevice(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateDeviceRequest,
) (*CreateDeviceResult, error) {
	enrollmentToken, err := rand.HexString(EnrollmentTokenRawLength)
	if err != nil {
		return nil, err
	}

	enrollmentTokenHash := hash.SHA256String(enrollmentToken)
	now := time.Now()

	device := &coredata.Device{
		ID:             gid.New(req.OrganizationID.TenantID(), coredata.DeviceEntityType),
		OrganizationID: req.OrganizationID,
		State:          coredata.DeviceStatePending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	token := &coredata.DeviceEnrollmentToken{
		ID:          gid.New(req.OrganizationID.TenantID(), coredata.DeviceEnrollmentTokenEntityType),
		DeviceID:    device.ID,
		HashedValue: enrollmentTokenHash,
		ExpiresAt:   now.Add(s.enrollmentTokenValidity),
		CreatedAt:   now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			ownerID, err := s.validateOwnerProfileID(
				ctx,
				conn,
				scope,
				req.OrganizationID,
				req.OwnerID,
			)
			if err != nil {
				return err
			}

			device.OwnerID = ownerID

			if err := device.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert device: %w", err)
			}

			if err := token.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert device enrollment token: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &CreateDeviceResult{
		Device:          device,
		EnrollmentToken: enrollmentToken,
	}, nil
}

// EnrollDevice creates a pending device owned by the caller's membership
// profile in the organization.
func (s *Service) EnrollDevice(
	ctx context.Context,
	scope coredata.Scoper,
	req EnrollDeviceRequest,
) (*CreateDeviceResult, error) {
	enrollmentToken, err := rand.HexString(EnrollmentTokenRawLength)
	if err != nil {
		return nil, err
	}

	enrollmentTokenHash := hash.SHA256String(enrollmentToken)
	now := time.Now()

	device := &coredata.Device{
		ID:             gid.New(req.OrganizationID.TenantID(), coredata.DeviceEntityType),
		OrganizationID: req.OrganizationID,
		State:          coredata.DeviceStatePending,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	token := &coredata.DeviceEnrollmentToken{
		ID:          gid.New(req.OrganizationID.TenantID(), coredata.DeviceEnrollmentTokenEntityType),
		DeviceID:    device.ID,
		HashedValue: enrollmentTokenHash,
		ExpiresAt:   now.Add(s.enrollmentTokenValidity),
		CreatedAt:   now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			profile := &coredata.MembershipProfile{}
			if err := profile.LoadByIdentityIDAndOrganizationID(
				ctx,
				conn,
				scope,
				req.IdentityID,
				req.OrganizationID,
			); err != nil {
				return fmt.Errorf("cannot load owner profile for identity: %w", err)
			}

			device.OwnerID = &profile.ID

			if err := device.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert device: %w", err)
			}

			if err := token.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert device enrollment token: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &CreateDeviceResult{
		Device:          device,
		EnrollmentToken: enrollmentToken,
	}, nil
}

// ExchangeEnrollmentToken redeems a one-shot enrollment token and returns
// the plaintext device API key. The token row is deleted on success.
func (s *Service) ExchangeEnrollmentToken(
	ctx context.Context,
	tokenString string,
) (string, error) {
	hashedValue := hash.SHA256String(tokenString)
	now := time.Now()

	var apiKey string

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			token := &coredata.DeviceEnrollmentToken{}
			if err := token.LoadByHashedValueForUpdate(ctx, conn, hashedValue); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrEnrollmentTokenAlreadyUsed
				}

				return fmt.Errorf("cannot load device enrollment token: %w", err)
			}

			if now.After(token.ExpiresAt) {
				if err := token.Delete(ctx, conn); err != nil {
					return fmt.Errorf("cannot delete expired device enrollment token: %w", err)
				}

				return ErrEnrollmentTokenExpired
			}

			scope := coredata.NewScope(token.TenantID)

			device := &coredata.Device{}
			if err := device.LoadByIDForUpdate(ctx, conn, scope, token.DeviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if device.State == coredata.DeviceStateRevoked {
				return ErrEnrollmentTokenInvalid
			}

			if len(device.APIKeyHash) > 0 {
				return ErrEnrollmentTokenInvalid
			}

			generatedKey, err := rand.HexString(APIKeyRawLength)
			if err != nil {
				return err
			}

			apiKeyHash := hash.SHA256String(generatedKey)
			if err := device.SetAPIKeyHash(ctx, conn, scope, apiKeyHash); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrEnrollmentTokenInvalid
				}

				return fmt.Errorf("cannot set device api key hash: %w", err)
			}

			if err := token.Delete(ctx, conn); err != nil {
				return fmt.Errorf("cannot delete device enrollment token: %w", err)
			}

			apiKey = generatedKey

			return nil
		},
	)
	if err != nil {
		return "", err
	}

	return apiKey, nil
}

func (s *Service) GetDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Service) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DeviceOrderField],
) (*page.Page[*coredata.Device, coredata.DeviceOrderField], error) {
	var devices coredata.Devices

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := devices.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
				return fmt.Errorf("cannot load devices: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(devices, cursor), nil
}

func (s *Service) ListForOrganizationIDAndOwnerID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	ownerID gid.GID,
	cursor *page.Cursor[coredata.DeviceOrderField],
) (*page.Page[*coredata.Device, coredata.DeviceOrderField], error) {
	var devices coredata.Devices

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			if err := devices.LoadByOrganizationIDAndOwnerID(
				ctx, conn, scope, organizationID, ownerID, cursor,
			); err != nil {
				return fmt.Errorf("cannot load devices by owner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(devices, cursor), nil
}

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			var ds coredata.Devices

			c, err := ds.CountByOrganizationID(ctx, conn, scope, organizationID)
			if err != nil {
				return fmt.Errorf("cannot count devices: %w", err)
			}

			count = c

			return nil
		},
	)

	return count, err
}

func (s *Service) CountForOrganizationIDAndOwnerID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	ownerID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
			}

			var ds coredata.Devices

			c, err := ds.CountByOrganizationIDAndOwnerID(
				ctx, conn, scope, organizationID, ownerID,
			)
			if err != nil {
				return fmt.Errorf("cannot count devices by owner: %w", err)
			}

			count = c

			return nil
		},
	)

	return count, err
}

func (s *Service) RevokeDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if err := device.Revoke(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot revoke device: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Service) SetDeviceOwner(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	ownerProfileID *gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			resolvedOwnerID, err := s.validateOwnerProfileID(
				ctx, conn, scope, device.OrganizationID, ownerProfileID,
			)
			if err != nil {
				return err
			}

			if err := device.AssignOwner(ctx, conn, scope, resolvedOwnerID); err != nil {
				return fmt.Errorf("cannot set device owner: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *Service) validateOwnerProfileID(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	organizationID gid.GID,
	ownerID *gid.GID,
) (*gid.GID, error) {
	if ownerID == nil {
		return nil, nil
	}

	if ownerID.EntityType() != coredata.MembershipProfileEntityType {
		return nil, fmt.Errorf("owner_id must be a membership profile")
	}

	profile := &coredata.MembershipProfile{}
	if err := profile.LoadByID(ctx, conn, scope, *ownerID); err != nil {
		return nil, fmt.Errorf("cannot load owner profile: %w", err)
	}

	if profile.OrganizationID != organizationID {
		return nil, fmt.Errorf("owner profile does not belong to organization")
	}

	return ownerID, nil
}

func (s *Service) GetLatestPostures(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (coredata.DevicePostures, error) {
	var postures coredata.DevicePostures

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			loaded, err := page.LoadAll(
				ctx,
				page.OrderBy[coredata.DevicePostureOrderField]{
					Field:     coredata.DevicePostureOrderFieldCheckKey,
					Direction: page.OrderDirectionAsc,
				},
				func(ctx context.Context, cursor *page.Cursor[coredata.DevicePostureOrderField]) ([]*coredata.DevicePosture, error) {
					var batch coredata.DevicePostures
					if err := batch.LoadLatestByDeviceID(ctx, conn, scope, deviceID, cursor); err != nil {
						return nil, fmt.Errorf("cannot load latest device postures: %w", err)
					}

					return batch, nil
				},
			)
			if err != nil {
				return fmt.Errorf("cannot load latest device postures: %w", err)
			}

			postures = loaded

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return postures, nil
}

func (s *Service) GetPostureHistory(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	checkKey string,
	limit int,
) (coredata.DevicePostures, error) {
	var postures coredata.DevicePostures

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := postures.LoadHistoryByDeviceIDAndCheckKey(ctx, conn, scope, deviceID, checkKey, limit); err != nil {
				return fmt.Errorf("cannot load device posture history: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return postures, nil
}

// AuthenticateDevice resolves a device API key to its device row.
// Returns coredata.ErrResourceNotFound when no non-revoked device
// matches the key. Revoked devices are treated as not found.
func (s *Service) AuthenticateDevice(
	ctx context.Context,
	apiKey string,
) (*coredata.Device, error) {
	if apiKey == "" {
		return nil, coredata.ErrResourceNotFound
	}

	hash := hash.SHA256String(apiKey)

	device := &coredata.Device{}

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return device.LoadByAPIKeyHash(ctx, conn, hash)
	})
	if err != nil {
		return nil, err
	}

	return device, nil
}

// RecordHeartbeat refreshes the device's last-seen timestamp and any
// version fields the agent sends. On the first heartbeat for a PENDING
// device, hardware metadata is recorded and the device is activated.
func (s *Service) RecordHeartbeat(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	req RecordHeartbeatRequest,
) (*coredata.Device, error) {
	if req.HardwareUUID == "" {
		return nil, fmt.Errorf("hardware_uuid is required")
	}

	if req.Hostname == "" {
		return nil, fmt.Errorf("hostname is required")
	}

	if !req.Platform.IsValid() {
		return nil, fmt.Errorf("invalid platform: %q", req.Platform)
	}

	if req.OSVersion == "" {
		return nil, fmt.Errorf("os_version is required")
	}

	if req.AgentVersion == "" {
		return nil, fmt.Errorf("agent_version is required")
	}

	device := &coredata.Device{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := device.LoadByIDForUpdate(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if device.State == coredata.DeviceStateRevoked {
				return ErrDeviceRevoked
			}

			hardwareUUID := req.HardwareUUID
			hostname := req.Hostname
			platform := req.Platform
			osVersion := req.OSVersion
			agentVersion := req.AgentVersion

			device.HardwareUUID = &hardwareUUID
			device.SerialNumber = req.SerialNumber
			device.Hostname = &hostname
			device.Platform = &platform
			device.OSVersion = &osVersion
			device.AgentVersion = &agentVersion

			switch device.State {
			case coredata.DeviceStatePending:
				if err := device.Activate(ctx, conn, scope); err != nil {
					if errors.Is(err, coredata.ErrResourceAlreadyExists) {
						return ErrDeviceHardwareConflict
					}

					if errors.Is(err, coredata.ErrResourceNotFound) {
						return ErrDeviceRevoked
					}

					return fmt.Errorf("cannot activate device: %w", err)
				}
			case coredata.DeviceStateActive:
				if err := device.UpdateHeartbeat(ctx, conn, scope); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return ErrDeviceRevoked
					}

					return fmt.Errorf("cannot update device heartbeat: %w", err)
				}
			default:
				return ErrDeviceRevoked
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return device, nil
}

// RecordPostures appends posture results for a device.
func (s *Service) RecordPostures(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	results []RecordPostureResult,
) error {
	if len(results) == 0 {
		return nil
	}

	now := time.Now()

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			device := &coredata.Device{}
			if err := device.LoadByIDForUpdate(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if device.State != coredata.DeviceStateActive {
				return ErrDeviceRevoked
			}

			for _, r := range results {
				posture := coredata.DevicePosture{
					ID:             gid.New(device.OrganizationID.TenantID(), coredata.DevicePostureEntityType),
					OrganizationID: device.OrganizationID,
					DeviceID:       device.ID,
					CheckKey:       r.CheckKey,
					Status:         r.Status,
					Evidence:       r.Evidence,
					ObservedAt:     r.ObservedAt,
					CreatedAt:      now,
				}
				if err := posture.Insert(ctx, conn, scope); err != nil {
					return fmt.Errorf("cannot insert device posture: %w", err)
				}
			}

			return nil
		},
	)
}

// UnenrollDevice revokes the device. The agent invokes this from its
// uninstaller before wiping its local API key.
func (s *Service) UnenrollDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			device := &coredata.Device{}
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if err := device.Revoke(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot revoke device: %w", err)
			}

			return nil
		},
	)
}
