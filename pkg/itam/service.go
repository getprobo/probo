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
	// ErrEnrollmentTokenInvalid is returned when an /enroll request is
	// made with a token that does not exist, is revoked, expired, or has
	// hit its max-uses cap.
	ErrEnrollmentTokenInvalid = errors.New("enrollment token is invalid")

	// ErrDeviceRevoked is returned when an agent authenticates with a
	// device API key for a revoked device.
	ErrDeviceRevoked = errors.New("device is revoked")
)

const (
	// EnrollmentTokenRawLength is the number of random bytes used for an
	// enrollment token secret. base64url-encoded that's 43 chars.
	EnrollmentTokenRawLength = 32

	// APIKeyRawLength is the number of random bytes for a device API
	// key secret. base64url-encoded that's 64 chars.
	APIKeyRawLength = 48
)

type (
	// Service is the IT Asset Management service. Tenant-scoped admin
	// operations take a scope parameter from the caller; agent-facing
	// operations (enroll, authenticate, heartbeat, postures, unenroll)
	// resolve their own scope internally because the agent does not know
	// which tenant it belongs to until enrolment resolves it.
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

	// CreateEnrollmentTokenResult is returned to the caller once, at
	// creation time. The plain token is never persisted; only its hash
	// is stored.
	CreateEnrollmentTokenResult struct {
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

func NewService(pgClient *pg.Client, iamSvc *iam.Service, logger *log.Logger) *Service {
	iamSvc.Authorizer.RegisterPolicySet(ITAMPolicySet())

	return &Service{
		pg:     pgClient,
		logger: logger,
	}
}

// hashSecret hashes a token/api-key value the same way for storage and
// lookup. SHA-256 is sufficient: the secret is high-entropy random.
func hashSecret(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// generateSecret returns a base64url-encoded random secret of the given
// raw byte length.
func generateSecret(rawLen int) (string, error) {
	buf := make([]byte, rawLen)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("cannot generate random secret: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// ------------------------------------------------------------------
// Admin-facing operations (tenant-scoped via caller-provided scope)
// ------------------------------------------------------------------

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

func (s *Service) ListEnrollmentTokens(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (coredata.DeviceEnrollmentTokens, error) {
	var tokens coredata.DeviceEnrollmentTokens
	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tokens.LoadByOrganizationID(ctx, conn, scope, organizationID); err != nil {
				return fmt.Errorf("cannot load device enrollment tokens: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return tokens, nil
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
				return err
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

// RecordHeartbeat updates the device's last-seen / version columns.
func (s *Service) RecordHeartbeat(
	ctx context.Context,
	scope coredata.Scoper,
	deviceID gid.GID,
	hostname string,
	osVersion string,
	agentVersion string,
) error {
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
	scope coredata.Scoper,
	deviceID gid.GID,
	results []RecordPostureResult,
) error {
	if len(results) == 0 {
		return nil
	}
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
	scope coredata.Scoper,
	deviceID gid.GID,
) error {
	return s.pg.WithTx(ctx, func(ctx context.Context, conn pg.Tx) error {
		device := &coredata.Device{}
		if err := device.LoadByID(ctx, conn, scope, deviceID); err != nil {
			return fmt.Errorf("cannot load device: %w", err)
		}
		return device.Revoke(ctx, conn, scope)
	})
}
