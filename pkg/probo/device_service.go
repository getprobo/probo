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

package probo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

// DeviceService exposes admin (tenant-scoped) operations on devices and
// enrollment tokens. Operations driven by the agent itself (Enroll,
// Heartbeat, RecordPostures, Unenroll) live on the parent Service since
// the agent does not know which tenant it belongs to until enrolment
// resolves it.
type DeviceService struct {
	svc *Service
}

var (
	// ErrEnrollmentTokenInvalid is returned when an /enroll request is
	// made with a token that does not exist, is revoked, expired, or has
	// hit its max-uses cap.
	ErrEnrollmentTokenInvalid = errors.New("enrollment token is invalid")

	// ErrDeviceRevoked is returned when an agent authenticates with a
	// device API key for a revoked device.
	ErrDeviceRevoked = errors.New("device is revoked")
)

const (
	// DeviceEnrollmentTokenRawLength is the number of random bytes used
	// for an enrollment token secret. base64url-encoded that's 43 chars.
	DeviceEnrollmentTokenRawLength = 32

	// DeviceAPIKeyRawLength is the number of random bytes for a device
	// API key secret. base64url-encoded that's 64 chars.
	DeviceAPIKeyRawLength = 48
)

type (
	CreateDeviceEnrollmentTokenRequest struct {
		OrganizationID      gid.GID
		Name                string
		Validity            time.Duration
		MaxUses             *int
		CreatedByIdentityID *gid.GID
	}

	// CreateDeviceEnrollmentTokenResult is returned to the caller once,
	// at creation time. The plain token is never persisted; only its hash
	// is stored.
	CreateDeviceEnrollmentTokenResult struct {
		Token *coredata.DeviceEnrollmentToken
		// Secret is the unhashed token value the admin needs to give the
		// agent installer. Shown once.
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

	EnrollDeviceResult struct {
		Device *coredata.Device
		// APIKey is the unhashed device-API-key the agent must persist
		// and present on every subsequent call. Shown once.
		APIKey string
	}

	RecordPostureResult struct {
		CheckKey   string
		Status     coredata.DevicePostureStatus
		Evidence   json.RawMessage
		ObservedAt time.Time
	}
)

// hashDeviceSecret hashes a token/api-key value the same way for storage
// and lookup. SHA-256 is sufficient: the secret is high-entropy random.
func hashDeviceSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// generateDeviceSecret returns a base64url-encoded random secret of the
// given raw byte length.
func generateDeviceSecret(rawLen int) (string, error) {
	buf := make([]byte, rawLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("cannot generate random secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ------------------------------------------------------------------
// Admin-facing operations (tenant-scoped)
// ------------------------------------------------------------------

func (s DeviceService) CreateEnrollmentToken(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateDeviceEnrollmentTokenRequest,
) (*CreateDeviceEnrollmentTokenResult, error) {
	if req.OrganizationID == gid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Validity <= 0 {
		req.Validity = 7 * 24 * time.Hour
	}

	secret, err := generateDeviceSecret(DeviceEnrollmentTokenRawLength)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	token := &coredata.DeviceEnrollmentToken{
		ID:                  gid.New(req.OrganizationID.TenantID(), coredata.DeviceEnrollmentTokenEntityType),
		OrganizationID:      req.OrganizationID,
		Name:                req.Name,
		TokenHash:           hashDeviceSecret(secret),
		CreatedByIdentityID: req.CreatedByIdentityID,
		ExpiresAt:           now.Add(req.Validity),
		MaxUses:             req.MaxUses,
		UsedCount:           0,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	err = s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, conn, scope, req.OrganizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}
		if err := token.Insert(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot insert device enrollment token: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreateDeviceEnrollmentTokenResult{Token: token, Secret: secret}, nil
}

func (s DeviceService) RevokeEnrollmentToken(
	ctx context.Context,
	scope coredata.Scoper,
	tokenID gid.GID,
) (*coredata.DeviceEnrollmentToken, error) {
	token := &coredata.DeviceEnrollmentToken{}

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		if err := token.LoadByID(ctx, conn, scope, tokenID); err != nil {
			return fmt.Errorf("cannot load device enrollment token: %w", err)
		}
		if err := token.Revoke(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot revoke device enrollment token: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s DeviceService) ListEnrollmentTokens(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (coredata.DeviceEnrollmentTokens, error) {
	var tokens coredata.DeviceEnrollmentTokens
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return tokens.LoadByOrganizationID(ctx, conn, scope, organizationID)
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (s DeviceService) GetDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return device.LoadByID(ctx, conn, scope, deviceID)
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s DeviceService) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DeviceOrderField],
) (*page.Page[*coredata.Device, coredata.DeviceOrderField], error) {
	var devices coredata.Devices
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return devices.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
	})
	if err != nil {
		return nil, err
	}
	return page.NewPage(devices, cursor), nil
}

func (s DeviceService) CountForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var ds coredata.Devices
		c, err := ds.CountByOrganizationID(ctx, conn, scope, organizationID)
		if err != nil {
			return err
		}
		count = c
		return nil
	})
	return count, err
}

func (s DeviceService) RevokeDevice(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}
	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
			return fmt.Errorf("cannot load device: %w", err)
		}
		if err := device.Revoke(ctx, conn, scope); err != nil {
			return fmt.Errorf("cannot revoke device: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s DeviceService) AssignDeviceToUser(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	identityID *gid.GID,
) (*coredata.Device, error) {
	device := &coredata.Device{}
	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
			return fmt.Errorf("cannot load device: %w", err)
		}
		if err := device.AssignUser(ctx, conn, scope, identityID); err != nil {
			return fmt.Errorf("cannot assign device user: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s DeviceService) GetLatestPostures(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
) (coredata.DevicePostures, error) {
	var postures coredata.DevicePostures
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return postures.LoadLatestByDeviceID(ctx, conn, scope, deviceID)
	})
	if err != nil {
		return nil, err
	}
	return postures, nil
}

func (s DeviceService) GetPostureHistory(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	checkKey string,
	limit int,
) (coredata.DevicePostures, error) {
	var postures coredata.DevicePostures
	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return postures.LoadHistoryByDeviceIDAndCheckKey(ctx, conn, scope, deviceID, checkKey, limit)
	})
	if err != nil {
		return nil, err
	}
	return postures, nil
}

// ------------------------------------------------------------------
// Agent-facing operations (cross-tenant; on parent Service)
// ------------------------------------------------------------------

// EnrollDevice consumes an enrollment token and provisions a new device
// row, returning the device-API-key the agent should persist.
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

	tokenHash := hashDeviceSecret(req.EnrollmentSecret)
	apiKey, err := generateDeviceSecret(DeviceAPIKeyRawLength)
	if err != nil {
		return nil, err
	}
	apiKeyHash := hashDeviceSecret(apiKey)

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

		scope := coredata.NewScope(token.OrganizationID.TenantID())

		existing := &coredata.Device{}
		err := existing.LoadByHardwareUUID(ctx, conn, scope, token.OrganizationID, req.HardwareUUID)
		switch {
		case err == nil:
			// Re-enroll: rotate API key, refresh metadata, clear
			// revocation. This makes installer re-runs idempotent.
			existing.APIKeyHash = apiKeyHash
			existing.Hostname = req.Hostname
			existing.SerialNumber = req.SerialNumber
			existing.Platform = req.Platform
			existing.OSVersion = req.OSVersion
			existing.AgentVersion = req.AgentVersion
			existing.LastSeenAt = now
			existing.UpdatedAt = now
			existing.RevokedAt = nil
			if err := existing.Reenroll(ctx, conn, scope); err != nil {
				return fmt.Errorf("cannot re-enroll device: %w", err)
			}
			device = existing
		case errors.Is(err, coredata.ErrResourceNotFound):
			device = &coredata.Device{
				ID:             gid.New(token.OrganizationID.TenantID(), coredata.DeviceEntityType),
				TenantID:       token.OrganizationID.TenantID(),
				OrganizationID: token.OrganizationID,
				HardwareUUID:   req.HardwareUUID,
				SerialNumber:   req.SerialNumber,
				Hostname:       req.Hostname,
				Platform:       req.Platform,
				OSVersion:      req.OSVersion,
				AgentVersion:   req.AgentVersion,
				APIKeyHash:     apiKeyHash,
				EnrolledAt:     now,
				LastSeenAt:     now,
				CreatedAt:      now,
				UpdatedAt:      now,
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

// AuthenticateDevice resolves a Bearer device API key to its device row.
// Returns ErrResourceNotFound or ErrDeviceRevoked when authentication
// fails.
func (s *Service) AuthenticateDevice(
	ctx context.Context,
	apiKey string,
) (*coredata.Device, error) {
	if apiKey == "" {
		return nil, coredata.ErrResourceNotFound
	}
	hash := hashDeviceSecret(apiKey)

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

// RecordHeartbeat updates the device's last-seen / version columns.
func (s *Service) RecordHeartbeat(
	ctx context.Context,
	deviceID gid.GID,
	hostname string,
	osVersion string,
	agentVersion string,
) error {
	scope := coredata.NewScope(deviceID.TenantID())
	return s.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
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
		return device.UpdateHeartbeat(ctx, conn, scope)
	})
}

// RecordPostures appends posture results for a device.
func (s *Service) RecordPostures(
	ctx context.Context,
	deviceID gid.GID,
	results []RecordPostureResult,
) error {
	if len(results) == 0 {
		return nil
	}
	scope := coredata.NewScope(deviceID.TenantID())
	now := time.Now()

	return s.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		device := &coredata.Device{}
		if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
			return fmt.Errorf("cannot load device: %w", err)
		}
		if device.RevokedAt != nil {
			return ErrDeviceRevoked
		}

		for _, r := range results {
			if !r.Status.IsValid() {
				return fmt.Errorf("invalid posture status: %q", r.Status)
			}
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
	})
}

// UnenrollDevice marks the device as revoked. The agent calls this from
// its uninstaller, before the local key is wiped.
func (s *Service) UnenrollDevice(
	ctx context.Context,
	deviceID gid.GID,
) error {
	scope := coredata.NewScope(deviceID.TenantID())
	return s.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		device := &coredata.Device{}
		if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
			return fmt.Errorf("cannot load device: %w", err)
		}
		return device.Revoke(ctx, conn, scope)
	})
}
