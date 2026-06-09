// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package itam

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/page"
)

var (
	// ErrEnrollmentTokenInvalid is returned when the presented
	// enrollment token is unknown, revoked, expired, or exhausted.
	ErrEnrollmentTokenInvalid = errors.New("enrollment token is invalid")

	// ErrDeviceRevoked is returned when the authenticated device has
	// been revoked.
	ErrDeviceRevoked = errors.New("device is revoked")
)

const (
	// EnrollmentTokenRawLength is the random byte length of an
	// enrollment token secret (43 chars once base64url-encoded).
	EnrollmentTokenRawLength = 32

	// APIKeyRawLength is the random byte length of a device API key
	// secret (64 chars once base64url-encoded).
	APIKeyRawLength = 48
)

type (
	// Service is the IT Asset Management service. Admin operations are
	// tenant-scoped via a caller-supplied scope; agent-facing operations
	// (enroll, authenticate, heartbeat, postures, unenroll) resolve
	// their own scope, since the agent does not know its tenant until
	// enrollment.
	Service struct {
		pg     *pg.Client
		logger *log.Logger
	}

	CreateEnrollmentTokenRequest struct {
		OrganizationID      gid.GID
		Name                string
		Validity            time.Duration
		MaxUses             *int
		CreatedByIdentityID *gid.GID
	}

	// CreateEnrollmentTokenResult carries the persisted token and its
	// plaintext secret. Only the hash is stored, so Secret is available
	// only at this point.
	CreateEnrollmentTokenResult struct {
		Token  *coredata.DeviceEnrollmentToken
		Secret string
	}

	EnrollDeviceRequest struct {
		EnrollmentSecret string
		HardwareUUID     string
		SerialNumber     *string
		Hostname         string
		Platform         coredata.DevicePlatform
		OSVersion        string
		AgentVersion     string
	}

	// EnrollDeviceResult carries the device row and the plaintext API
	// key the agent must persist. Only the hash is stored, so APIKey is
	// available only at this point.
	EnrollDeviceResult struct {
		Device *coredata.Device
		APIKey string
	}

	RecordPostureResult struct {
		CheckKey   string
		Status     coredata.DevicePostureStatus
		Evidence   json.RawMessage
		ObservedAt time.Time
	}

	EnrollmentStatusResult struct {
		Token           *coredata.DeviceEnrollmentToken
		Device          *coredata.Device
		FirstActivityAt *time.Time
	}
)

func NewService(pgClient *pg.Client, iamSvc *iam.Service, logger *log.Logger) *Service {
	iamSvc.Authorizer.RegisterPolicySet(ITAMPolicySet())

	return &Service{
		pg:     pgClient,
		logger: logger,
	}
}

// hashSecret hashes an enrollment token or device API key for storage
// and lookup. Unsalted SHA-256 is sufficient: the input is a random
// secret with at least 256 bits of entropy.
func hashSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// generateSecret returns a base64url-encoded random secret of rawLen
// bytes.
func generateSecret(rawLen int) (string, error) {
	buf := make([]byte, rawLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("cannot generate random secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (s *Service) CreateEnrollmentToken(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateEnrollmentTokenRequest,
) (*CreateEnrollmentTokenResult, error) {
	if req.OrganizationID == gid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Validity <= 0 {
		req.Validity = 7 * 24 * time.Hour
	}

	secret, err := generateSecret(EnrollmentTokenRawLength)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	token := &coredata.DeviceEnrollmentToken{
		ID:                  gid.New(req.OrganizationID.TenantID(), coredata.DeviceEnrollmentTokenEntityType),
		OrganizationID:      req.OrganizationID,
		Name:                req.Name,
		TokenHash:           hashSecret(secret),
		CreatedByIdentityID: req.CreatedByIdentityID,
		ExpiresAt:           now.Add(req.Validity),
		MaxUses:             req.MaxUses,
		UsedCount:           0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			organization := &coredata.Organization{}
			if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
				return fmt.Errorf("cannot load organization: %w", err)
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

	return &CreateEnrollmentTokenResult{Token: token, Secret: secret}, nil
}

func (s *Service) RevokeEnrollmentToken(
	ctx context.Context,
	scope coredata.Scoper,
	tokenID gid.GID,
) (*coredata.DeviceEnrollmentToken, error) {
	token := &coredata.DeviceEnrollmentToken{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := token.LoadByID(ctx, conn, scope, tokenID); err != nil {
				return fmt.Errorf("cannot load device enrollment token: %w", err)
			}

			if err := token.Revoke(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot revoke device enrollment token: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) GetEnrollmentStatus(
	ctx context.Context,
	scope coredata.Scoper,
	enrollmentTokenID gid.GID,
) (*EnrollmentStatusResult, error) {
	result := &EnrollmentStatusResult{}
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			token := &coredata.DeviceEnrollmentToken{}
			if err := token.LoadByID(ctx, conn, scope, enrollmentTokenID); err != nil {
				return fmt.Errorf("cannot load device enrollment token: %w", err)
			}
			result.Token = token

			device := &coredata.Device{}
			if err := device.LoadLatestByEnrollmentTokenID(ctx, conn, scope, enrollmentTokenID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}
				return fmt.Errorf("cannot load latest device by enrollment token: %w", err)
			}
			result.Device = device

			posture := &coredata.DevicePosture{}
			if err := posture.LoadFirstObservedAtByDeviceID(ctx, conn, scope, device.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}
				return fmt.Errorf("cannot load first device activity posture: %w", err)
			}

			result.FirstActivityAt = &posture.ObservedAt

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
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

func (s *Service) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
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

func (s *Service) RevokeDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}

	err := s.pg.WithTx(
		ctx, func(
			ctx context.Context, conn pg.Tx) error {
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

func (s *Service) AssignDeviceToUser(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	identityID *gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if err := device.AssignUser(ctx, conn, scope, identityID); err != nil {
				return fmt.Errorf("cannot assign device user: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return device, nil
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
			if err := postures.LoadLatestByDeviceID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load latest device postures: %w", err)
			}

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

func (s *Service) EnrollDevice(
	ctx context.Context,
	req EnrollDeviceRequest,
) (*EnrollDeviceResult, error) {
	if req.EnrollmentSecret == "" {
		return nil, ErrEnrollmentTokenInvalid
	}

	if req.HardwareUUID == "" {
		return nil, fmt.Errorf("hardware_uuid is required")
	}

	if !req.Platform.IsValid() {
		return nil, fmt.Errorf("invalid platform: %q", req.Platform)
	}

	tokenHash := hashSecret(req.EnrollmentSecret)
	apiKey, err := generateSecret(APIKeyRawLength)
	if err != nil {
		return nil, err
	}

	apiKeyHash := hashSecret(apiKey)

	var device *coredata.Device
	now := time.Now()

	err = s.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		token := &coredata.DeviceEnrollmentToken{}
		if err := token.LoadByTokenHashForUpdate(ctx, conn, tokenHash); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrEnrollmentTokenInvalid
			}
			return fmt.Errorf("cannot load device enrollment token: %w", err)
		}
		if !token.IsUsable(now) {
			return ErrEnrollmentTokenInvalid
		}
		enrollmentTokenID := token.ID

		scope := coredata.NewScope(token.OrganizationID.TenantID())

		existing := &coredata.Device{}
		err := existing.LoadByHardwareUUID(ctx, conn, scope, token.OrganizationID, req.HardwareUUID)
		switch {
		case err == nil:
			// Re-enrollment: rotate the API key, refresh metadata,
			// and clear any prior revocation so installer re-runs
			// are idempotent.
			existing.APIKeyHash = apiKeyHash
			existing.Hostname = req.Hostname
			existing.SerialNumber = req.SerialNumber
			existing.Platform = req.Platform
			existing.OSVersion = req.OSVersion
			existing.AgentVersion = req.AgentVersion
			existing.EnrollmentTokenID = &enrollmentTokenID
			existing.LastSeenAt = now
			existing.UpdatedAt = now
			existing.RevokedAt = nil
			if err := existing.Reenroll(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot re-enroll device: %w", err)
			}
			device = existing
		case errors.Is(err, coredata.ErrResourceNotFound):
			device = &coredata.Device{
				ID:                gid.New(token.OrganizationID.TenantID(), coredata.DeviceEntityType),
				TenantID:          token.OrganizationID.TenantID(),
				OrganizationID:    token.OrganizationID,
				HardwareUUID:      req.HardwareUUID,
				SerialNumber:      req.SerialNumber,
				Hostname:          req.Hostname,
				Platform:          req.Platform,
				OSVersion:         req.OSVersion,
				AgentVersion:      req.AgentVersion,
				EnrollmentTokenID: &enrollmentTokenID,
				APIKeyHash:        apiKeyHash,
				EnrolledAt:        now,
				LastSeenAt:        now,
				CreatedAt:         now,
				UpdatedAt:         now,
			}
			if err := device.Insert(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot insert device: %w", err)
			}
		default:
			return fmt.Errorf("cannot check existing device: %w", err)
		}

		if err := token.IncrementUsage(ctx, conn); err != nil {
			return fmt.Errorf("cannot increment enrollment token usage: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &EnrollDeviceResult{Device: device, APIKey: apiKey}, nil
}

// AuthenticateDevice resolves a device API key to its device row.
// Returns coredata.ErrResourceNotFound when no device matches the key
// and ErrDeviceRevoked when the matching device has been revoked.
func (s *Service) AuthenticateDevice(
	ctx context.Context,
	apiKey string,
) (*coredata.Device, error) {
	if apiKey == "" {
		return nil, coredata.ErrResourceNotFound
	}
	hash := hashSecret(apiKey)

	device := &coredata.Device{}
	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return device.LoadByAPIKeyHash(ctx, conn, hash)
	})
	if err != nil {
		return nil, err
	}
	if device.RevokedAt != nil {
		return nil, ErrDeviceRevoked
	}
	return device, nil
}

// RecordHeartbeat refreshes the device's last-seen timestamp and any
// version fields the agent sends.
func (s *Service) RecordHeartbeat(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	hostname string,
	osVersion string,
	agentVersion string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, conn pg.Tx) error {
			device := &coredata.Device{}
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if device.RevokedAt != nil {
				return ErrDeviceRevoked
			}

			if hostname != "" {
				device.Hostname = hostname
			}

			if osVersion != "" {
				device.OSVersion = osVersion
			}

			if agentVersion != "" {
				device.AgentVersion = agentVersion
			}

			if err := device.UpdateHeartbeat(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot update device heartbeat: %w", err)
			}

			return nil
		},
	)
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
			if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
				return fmt.Errorf("cannot load device: %w", err)
			}

			if device.RevokedAt != nil {
				return ErrDeviceRevoked
			}

			for _, r := range results {
				observed := r.ObservedAt
				if observed.IsZero() {
					observed = now
				}

				posture := coredata.DevicePosture{
					ID:             gid.New(device.OrganizationID.TenantID(), coredata.DevicePostureEntityType),
					OrganizationID: device.OrganizationID,
					DeviceID:       device.ID,
					CheckKey:       r.CheckKey,
					Status:         r.Status,
					Evidence:       r.Evidence,
					ObservedAt:     observed,
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
