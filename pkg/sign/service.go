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

package sign

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/webhook"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	pg          *pg.Client
	fileManager *filemanager.Service
	logger      *log.Logger
}

func NewService(
	pgClient *pg.Client,
	fileManager *filemanager.Service,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:          pgClient,
		fileManager: fileManager,
		logger:      logger,
	}
}

// ─── Signing Config CRUD ────────────────────────────────────────────────────

type CreateSigningConfigRequest struct {
	OrganizationID     gid.GID
	DocumentID         gid.GID
	Kind               coredata.DocumentSigningConfigKind
	ReacceptancePolicy coredata.DocumentSigningReacceptancePolicy
	Mode               coredata.DocumentSigningConfigMode
	AllowedOrigins     []string
}

func (s *Service) CreateSigningConfig(
	ctx context.Context,
	scope coredata.Scoper,
	req CreateSigningConfigRequest,
) (*coredata.DocumentSigningConfig, error) {
	publishableKey, err := newPublishableKey(req.Mode)
	if err != nil {
		return nil, fmt.Errorf("cannot generate publishable key: %w", err)
	}

	now := time.Now()
	cfg := &coredata.DocumentSigningConfig{
		ID:                 gid.New(scope.GetTenantID(), coredata.DocumentSigningConfigEntityType),
		OrganizationID:     req.OrganizationID,
		DocumentID:         req.DocumentID,
		PublishableKey:     publishableKey,
		State:              coredata.DocumentSigningConfigStateDraft,
		Mode:               req.Mode,
		Kind:               req.Kind,
		ReacceptancePolicy: req.ReacceptancePolicy,
		AllowedOrigins:     req.AllowedOrigins,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return cfg.Insert(ctx, tx, scope)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot create signing config: %w", err)
	}

	return cfg, nil
}

type UpdateSigningConfigRequest struct {
	ID                 gid.GID
	Kind               *coredata.DocumentSigningConfigKind
	ReacceptancePolicy *coredata.DocumentSigningReacceptancePolicy
	AllowedOrigins     *[]string
}

func (s *Service) UpdateSigningConfig(
	ctx context.Context,
	scope coredata.Scoper,
	req UpdateSigningConfigRequest,
) (*coredata.DocumentSigningConfig, error) {
	var cfg coredata.DocumentSigningConfig

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := cfg.LoadByID(ctx, tx, scope, req.ID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrConfigNotFound
			}

			return fmt.Errorf("cannot load signing config: %w", err)
		}

		if req.Kind != nil {
			cfg.Kind = *req.Kind
		}

		if req.ReacceptancePolicy != nil {
			cfg.ReacceptancePolicy = *req.ReacceptancePolicy
		}

		if req.AllowedOrigins != nil {
			cfg.AllowedOrigins = *req.AllowedOrigins
		}

		cfg.UpdatedAt = time.Now()

		return cfg.Update(ctx, tx, scope)
	})
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) ActivateSigningConfig(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.DocumentSigningConfig, error) {
	var cfg coredata.DocumentSigningConfig

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := cfg.LoadByID(ctx, tx, scope, id); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrConfigNotFound
			}

			return fmt.Errorf("cannot load signing config: %w", err)
		}

		cfg.State = coredata.DocumentSigningConfigStateActive
		cfg.UpdatedAt = time.Now()

		return cfg.Update(ctx, tx, scope)
	})
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) ArchiveSigningConfig(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.DocumentSigningConfig, error) {
	var cfg coredata.DocumentSigningConfig

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := cfg.LoadByID(ctx, tx, scope, id); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrConfigNotFound
			}

			return fmt.Errorf("cannot load signing config: %w", err)
		}

		cfg.State = coredata.DocumentSigningConfigStateArchived
		cfg.UpdatedAt = time.Now()

		return cfg.Update(ctx, tx, scope)
	})
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) GetSigningConfig(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.DocumentSigningConfig, error) {
	var cfg coredata.DocumentSigningConfig

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := cfg.LoadByID(ctx, conn, scope, id); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrConfigNotFound
			}

			return fmt.Errorf("cannot load signing config: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (s *Service) ListSigningConfigs(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.DocumentSigningConfigOrderField],
) (coredata.DocumentSigningConfigs, error) {
	var configs coredata.DocumentSigningConfigs

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return configs.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list signing configs: %w", err)
	}

	return configs, nil
}

// ─── Organization Signing Secrets ───────────────────────────────────────────

func (s *Service) RotateOrganizationSigningSecret(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	mode coredata.DocumentSigningConfigMode,
) (string, *coredata.OrganizationSigningSecret, error) {
	plaintext, err := newSecretKeyPlaintext(mode)
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate secret key: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("cannot hash secret key: %w", err)
	}

	now := time.Now()
	secret := &coredata.OrganizationSigningSecret{
		ID:             gid.New(scope.GetTenantID(), coredata.OrganizationSigningSecretEntityType),
		OrganizationID: organizationID,
		SecretKeyHash:  string(hash),
		Mode:           mode,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		_, err := secret.Upsert(ctx, tx, scope)

		return err
	})
	if err != nil {
		return "", nil, fmt.Errorf("cannot rotate signing secret: %w", err)
	}

	return plaintext, secret, nil
}

// ─── Status Computation ─────────────────────────────────────────────────────

type AcceptanceStatus string

const (
	AcceptanceStatusRequired             AcceptanceStatus = "REQUIRED"
	AcceptanceStatusAccepted             AcceptanceStatus = "ACCEPTED"
	AcceptanceStatusReacceptanceRequired AcceptanceStatus = "REACCEPTANCE_REQUIRED"
	AcceptanceStatusArchived             AcceptanceStatus = "ARCHIVED"
)

type SubjectStatus struct {
	Status              AcceptanceStatus
	AcceptedVersionID   *gid.GID
	AcceptedAt          *time.Time
	CurrentVersionID    *gid.GID
	CurrentVersionMajor int
	CurrentVersionMinor int
}

func (s *Service) ComputeSubjectStatus(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	cfg *coredata.DocumentSigningConfig,
	subject *coredata.ClickwrapSubject,
) (*SubjectStatus, error) {
	if cfg.State == coredata.DocumentSigningConfigStateArchived {
		return &SubjectStatus{Status: AcceptanceStatusArchived}, nil
	}

	var currentVersion coredata.DocumentVersion
	if err := currentVersion.LoadCurrentPresentable(ctx, conn, scope, cfg.DocumentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return &SubjectStatus{Status: AcceptanceStatusRequired}, nil
		}

		return nil, fmt.Errorf("cannot load current presentable version: %w", err)
	}

	status := &SubjectStatus{
		CurrentVersionID:    &currentVersion.ID,
		CurrentVersionMajor: currentVersion.Major,
		CurrentVersionMinor: currentVersion.Minor,
	}

	var latestAcceptance coredata.ClickwrapAcceptance

	err := latestAcceptance.LoadLatestBySubjectAndConfig(
		ctx, conn, scope,
		cfg.OrganizationID, cfg.ID, subject.ID,
	)
	if err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			status.Status = AcceptanceStatusRequired
			return status, nil
		}

		return nil, fmt.Errorf("cannot load latest acceptance: %w", err)
	}

	status.AcceptedVersionID = &latestAcceptance.DocumentVersionID
	status.AcceptedAt = &latestAcceptance.AcceptedAt

	if latestAcceptance.DocumentVersionID == currentVersion.ID {
		status.Status = AcceptanceStatusAccepted
		return status, nil
	}

	switch cfg.ReacceptancePolicy {
	case coredata.DocumentSigningReacceptancePolicyNever:
		status.Status = AcceptanceStatusAccepted
	case coredata.DocumentSigningReacceptancePolicyOnAny:
		status.Status = AcceptanceStatusReacceptanceRequired
	case coredata.DocumentSigningReacceptancePolicyOnMajor:
		var acceptedVersion coredata.DocumentVersion
		if err := acceptedVersion.LoadByID(ctx, conn, scope, latestAcceptance.DocumentVersionID); err != nil {
			return nil, fmt.Errorf("cannot load accepted version: %w", err)
		}

		if acceptedVersion.Major < currentVersion.Major {
			status.Status = AcceptanceStatusReacceptanceRequired
		} else {
			status.Status = AcceptanceStatusAccepted
		}
	default:
		status.Status = AcceptanceStatusReacceptanceRequired
	}

	return status, nil
}

// ─── Acceptances ────────────────────────────────────────────────────────────

type RecordAcceptanceRequest struct {
	OrganizationID    gid.GID
	SigningConfigID   gid.GID
	ExternalUserID    string
	Email             string
	FullName          *string
	AccountID         *string
	DocumentVersionID gid.GID
	ConsentText       string
	IPAddress         string
	UserAgent         string
	SdkVersion        *string
	CountryCode       *string
	IdempotencyKey    *string
	Mode              coredata.DocumentSigningConfigMode
}

type RecordAcceptanceResult struct {
	Acceptance *coredata.ClickwrapAcceptance
	IsNew      bool
}

func (s *Service) RecordAcceptance(
	ctx context.Context,
	scope coredata.Scoper,
	req RecordAcceptanceRequest,
) (*RecordAcceptanceResult, error) {
	if err := validateEmail(req.Email); err != nil {
		return nil, ErrInvalidEmail
	}

	if err := validateIPAddress(req.IPAddress); err != nil {
		return nil, ErrInvalidIPAddress
	}

	var result RecordAcceptanceResult

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		var cfg coredata.DocumentSigningConfig
		if err := cfg.LoadByID(ctx, tx, scope, req.SigningConfigID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrConfigNotFound
			}

			return fmt.Errorf("cannot load signing config: %w", err)
		}

		if cfg.State != coredata.DocumentSigningConfigStateActive {
			return ErrConfigNotActive
		}

		if cfg.Mode == coredata.DocumentSigningConfigModeLive &&
			req.Mode == coredata.DocumentSigningConfigModeTest {
			return ErrTestKeyOnLiveConfig
		}

		var docVersion coredata.DocumentVersion
		if err := docVersion.LoadByID(ctx, tx, scope, req.DocumentVersionID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrVersionNotPresentable
			}

			return fmt.Errorf("cannot load document version: %w", err)
		}

		if !isVersionPresentable(&docVersion) {
			return ErrVersionNotPresentable
		}

		if docVersion.DocumentID != cfg.DocumentID {
			return ErrVersionMismatch
		}

		if docVersion.ContentSha256 == nil {
			return ErrVersionNotPresentable
		}

		consentText := buildConsentText(&cfg, &docVersion)
		if req.ConsentText != consentText {
			return ErrConsentTextMismatch
		}

		now := time.Now()

		subject := &coredata.ClickwrapSubject{
			ID:                gid.New(scope.GetTenantID(), coredata.ClickwrapSubjectEntityType),
			OrganizationID:    req.OrganizationID,
			ExternalUserID:    req.ExternalUserID,
			Email:             req.Email,
			FullName:          req.FullName,
			ExternalAccountID: req.AccountID,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if _, err := subject.Upsert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot upsert subject: %w", err)
		}

		acceptance := &coredata.ClickwrapAcceptance{
			ID:                gid.New(scope.GetTenantID(), coredata.ClickwrapAcceptanceEntityType),
			OrganizationID:    req.OrganizationID,
			SigningConfigID:   req.SigningConfigID,
			DocumentVersionID: req.DocumentVersionID,
			SubjectID:         subject.ID,
			AcceptedEmail:     req.Email,
			AcceptedFullName:  req.FullName,
			AcceptedAccountID: req.AccountID,
			ConsentText:       req.ConsentText,
			ContentSha256:     *docVersion.ContentSha256,
			IPAddress:         req.IPAddress,
			UserAgent:         req.UserAgent,
			SdkVersion:        req.SdkVersion,
			CountryCode:       req.CountryCode,
			Mode:              req.Mode,
			IdempotencyKey:    req.IdempotencyKey,
			SealVersion:       2,
			SealAttemptCount:  0,
			SealMaxAttempts:   10,
			AcceptedAt:        now,
			CreatedAt:         now,
			UpdatedAt:         now,
		}

		if err := acceptance.Insert(ctx, tx, scope); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				// ON CONFLICT DO NOTHING: idempotent match — load the existing row
				if req.IdempotencyKey != nil {
					var existing coredata.ClickwrapAcceptance
					if err2 := existing.LoadByIdempotencyKey(
						ctx, tx, scope,
						req.OrganizationID, req.SigningConfigID, *req.IdempotencyKey,
					); err2 == nil {
						result = RecordAcceptanceResult{Acceptance: &existing, IsNew: false}
						return nil
					}
				}
			}

			return fmt.Errorf("cannot insert acceptance: %w", err)
		}

		result = RecordAcceptanceResult{Acceptance: acceptance, IsNew: true}

		if req.Mode == coredata.DocumentSigningConfigModeLive {
			if err := s.fireAcceptanceCreatedWebhook(ctx, tx, scope, acceptance, subject); err != nil {
				s.logger.WarnCtx(ctx, "cannot fire acceptance created webhook", log.Error(err))
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Service) RevokeAcceptance(
	ctx context.Context,
	scope coredata.Scoper,
	acceptanceID gid.GID,
	reason *string,
) (*coredata.ClickwrapAcceptance, error) {
	var acceptance coredata.ClickwrapAcceptance

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := acceptance.LoadByID(ctx, tx, scope, acceptanceID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrAcceptanceNotFound
			}

			return fmt.Errorf("cannot load acceptance: %w", err)
		}

		if acceptance.RevokedAt != nil {
			return ErrAlreadyRevoked
		}

		now := time.Now()
		acceptance.RevokedAt = &now
		acceptance.RevocationReason = reason
		acceptance.UpdatedAt = now

		if err := acceptance.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update acceptance: %w", err)
		}

		if acceptance.Mode == coredata.DocumentSigningConfigModeLive {
			if err := s.fireAcceptanceRevokedWebhook(ctx, tx, scope, &acceptance); err != nil {
				s.logger.WarnCtx(ctx, "cannot fire acceptance revoked webhook", log.Error(err))
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &acceptance, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("empty email")
	}

	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func validateIPAddress(ip string) error {
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address: %s", ip)
	}

	return nil
}

func isVersionPresentable(dv *coredata.DocumentVersion) bool {
	return dv.Status == "PUBLISHED" &&
		dv.FileID != nil &&
		dv.ContentSha256 != nil &&
		dv.PresentableAt != nil
}

// buildConsentText constructs the canonical consent text the customer must
// display next to the checkbox. The spec requires the data plane to compare
// the submitted text against this value case-sensitively.
func buildConsentText(cfg *coredata.DocumentSigningConfig, dv *coredata.DocumentVersion) string {
	kindLabel := "Terms of Service"

	switch cfg.Kind {
	case coredata.DocumentSigningConfigKindTermsOfService:
		kindLabel = "Terms of Service"
	case coredata.DocumentSigningConfigKindAcceptableUsePolicy:
		kindLabel = "Acceptable Use Policy"
	case coredata.DocumentSigningConfigKindOther:
		kindLabel = "Agreement"
	}

	return fmt.Sprintf(
		"By checking this box, you agree to our %s (v%d.%d).",
		kindLabel, dv.Major, dv.Minor,
	)
}

// GetSubject returns a clickwrap subject by ID.
func (s *Service) GetSubject(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.ClickwrapSubject, error) {
	var subj coredata.ClickwrapSubject

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := subj.LoadByID(ctx, conn, scope, id); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return ErrSubjectNotFound
			}

			return fmt.Errorf("cannot load subject: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &subj, nil
}

// ListAcceptancesForConfig returns paginated acceptances for a signing config.
func (s *Service) ListAcceptancesForConfig(
	ctx context.Context,
	scope coredata.Scoper,
	configID gid.GID,
	cursor *page.Cursor[coredata.ClickwrapAcceptanceOrderField],
) (*page.Page[*coredata.ClickwrapAcceptance, coredata.ClickwrapAcceptanceOrderField], error) {
	var acceptances coredata.ClickwrapAcceptances

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return acceptances.LoadBySigningConfigID(ctx, conn, scope, configID, cursor)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list acceptances: %w", err)
	}

	return page.NewPage(acceptances, cursor), nil
}

// ListAcceptancesForSubject returns paginated acceptances for a subject.
func (s *Service) ListAcceptancesForSubject(
	ctx context.Context,
	scope coredata.Scoper,
	subjectID gid.GID,
	cursor *page.Cursor[coredata.ClickwrapAcceptanceOrderField],
) (*page.Page[*coredata.ClickwrapAcceptance, coredata.ClickwrapAcceptanceOrderField], error) {
	var acceptances coredata.ClickwrapAcceptances

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return acceptances.LoadBySubjectID(ctx, conn, scope, subjectID, cursor)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot list subject acceptances: %w", err)
	}

	return page.NewPage(acceptances, cursor), nil
}

// IntegrationHealth contains recent rejected requests and last acceptance time.
type IntegrationHealth struct {
	LastRejectedRequests coredata.DataPlaneRequestLogs
	LastAcceptanceAt     *time.Time
}

// GetIntegrationHealth returns rejected requests and last acceptance for a config.
func (s *Service) GetIntegrationHealth(
	ctx context.Context,
	configID gid.GID,
) (*IntegrationHealth, error) {
	scope := coredata.NewScopeFromObjectID(configID)

	var health IntegrationHealth

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		var logs coredata.DataPlaneRequestLogs
		if err := logs.LoadBySigningConfigID(ctx, conn, scope, configID, 10); err != nil {
			return fmt.Errorf("cannot load rejected requests: %w", err)
		}

		health.LastRejectedRequests = logs

		var latestAcceptance coredata.ClickwrapAcceptance

		cursor := page.NewCursor[coredata.ClickwrapAcceptanceOrderField](
			1, nil, page.Head,
			page.OrderBy[coredata.ClickwrapAcceptanceOrderField]{
				Field:     coredata.ClickwrapAcceptanceOrderFieldAcceptedAt,
				Direction: page.OrderDirectionDesc,
			},
		)

		var acceptances coredata.ClickwrapAcceptances
		if err := acceptances.LoadBySigningConfigID(ctx, conn, scope, configID, cursor); err != nil {
			return fmt.Errorf("cannot load last acceptance: %w", err)
		}

		if len(acceptances) > 0 {
			latestAcceptance = *acceptances[0]
			health.LastAcceptanceAt = &latestAcceptance.AcceptedAt
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &health, nil
}

// GetCurrentPresentableVersion returns the current PRESENTABLE version for a
// signing config's document, or ErrDocumentNotPresentable if none exists.
func (s *Service) GetCurrentPresentableVersion(
	ctx context.Context,
	conn pg.Querier,
	scope coredata.Scoper,
	cfg *coredata.DocumentSigningConfig,
) (*coredata.DocumentVersion, error) {
	var dv coredata.DocumentVersion

	if err := dv.LoadCurrentPresentable(ctx, conn, scope, cfg.DocumentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrDocumentNotPresentable
		}

		return nil, fmt.Errorf("cannot load current presentable version: %w", err)
	}

	return &dv, nil
}

func (s *Service) GeneratePDFURL(
	ctx context.Context,
	fileID gid.GID,
	expiresIn time.Duration,
) (string, error) {
	var (
		scope = coredata.NewScopeFromObjectID(fileID)
		file  coredata.File
	)

	err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return file.LoadByID(ctx, conn, scope, fileID)
	})
	if err != nil {
		return "", fmt.Errorf("cannot load file: %w", err)
	}

	url, err := s.fileManager.GeneratePresignedURL(ctx, &file, expiresIn)
	if err != nil {
		return "", fmt.Errorf("cannot generate presigned URL: %w", err)
	}

	return url, nil
}

// ─── Auth helpers ────────────────────────────────────────────────────────────

type ResolvedKey struct {
	OrganizationID gid.GID
	Mode           coredata.DocumentSigningConfigMode
	IsSecret       bool
}

func (s *Service) ResolveSecretKey(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
	key string,
) (*ResolvedKey, error) {
	if !isSecretKey(key) {
		return nil, ErrInvalidKey
	}

	mode, ok := keyMode(key)
	if !ok {
		return nil, ErrInvalidKey
	}

	scope := coredata.NewScopeFromObjectID(organizationID)

	var secret coredata.OrganizationSigningSecret
	if err := secret.LoadByOrganizationID(ctx, conn, scope, organizationID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrInvalidKey
		}

		return nil, fmt.Errorf("cannot load signing secret: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(secret.SecretKeyHash), []byte(key)); err != nil {
		return nil, ErrInvalidKey
	}

	return &ResolvedKey{
		OrganizationID: organizationID,
		Mode:           mode,
		IsSecret:       true,
	}, nil
}

// ─── Webhooks ─────────────────────────────────────────────────────────────

type acceptanceCreatedPayload struct {
	AcceptanceID    string    `json:"acceptance_id"`
	SigningConfigID string    `json:"signing_config_id"`
	VersionID       string    `json:"document_version_id"`
	SubjectID       string    `json:"subject_id"`
	ExternalUserID  string    `json:"external_user_id"`
	Email           string    `json:"email"`
	AcceptedAt      time.Time `json:"accepted_at"`
	Mode            string    `json:"mode"`
}

type acceptanceRevokedPayload struct {
	AcceptanceID    string    `json:"acceptance_id"`
	SigningConfigID string    `json:"signing_config_id"`
	SubjectID       string    `json:"subject_id"`
	ExternalUserID  string    `json:"external_user_id"`
	RevokedAt       time.Time `json:"revoked_at"`
	Reason          *string   `json:"revocation_reason,omitempty"`
}

type reacceptanceRequiredPayload struct {
	SigningConfigID    string `json:"signing_config_id"`
	DocumentID         string `json:"document_id"`
	NewVersionID       string `json:"new_version_id"`
	NewVersionMajor    int    `json:"new_version_major"`
	NewVersionMinor    int    `json:"new_version_minor"`
	ReacceptancePolicy string `json:"reacceptance_policy"`
}

func (s *Service) fireAcceptanceCreatedWebhook(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	a *coredata.ClickwrapAcceptance,
	subject *coredata.ClickwrapSubject,
) error {
	return webhook.InsertData(
		ctx,
		tx,
		scope,
		a.OrganizationID,
		coredata.WebhookEventTypeDocumentSigningAcceptanceCreated,
		acceptanceCreatedPayload{
			AcceptanceID:    a.ID.String(),
			SigningConfigID: a.SigningConfigID.String(),
			VersionID:       a.DocumentVersionID.String(),
			SubjectID:       a.SubjectID.String(),
			ExternalUserID:  subject.ExternalUserID,
			Email:           a.AcceptedEmail,
			AcceptedAt:      a.AcceptedAt,
			Mode:            a.Mode.String(),
		},
	)
}

func (s *Service) fireAcceptanceRevokedWebhook(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	a *coredata.ClickwrapAcceptance,
) error {
	subject := coredata.ClickwrapSubject{}
	if err := subject.LoadByID(ctx, tx, scope, a.SubjectID); err != nil {
		return fmt.Errorf("cannot load subject for webhook: %w", err)
	}

	var revokedAt time.Time
	if a.RevokedAt != nil {
		revokedAt = *a.RevokedAt
	}

	return webhook.InsertData(
		ctx,
		tx,
		scope,
		a.OrganizationID,
		coredata.WebhookEventTypeDocumentSigningAcceptanceRevoked,
		acceptanceRevokedPayload{
			AcceptanceID:    a.ID.String(),
			SigningConfigID: a.SigningConfigID.String(),
			SubjectID:       a.SubjectID.String(),
			ExternalUserID:  subject.ExternalUserID,
			RevokedAt:       revokedAt,
			Reason:          a.RevocationReason,
		},
	)
}

func (s *Service) FireReacceptanceRequiredWebhook(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	cfg *coredata.DocumentSigningConfig,
	newVersion *coredata.DocumentVersion,
) error {
	return webhook.InsertData(
		ctx,
		tx,
		scope,
		cfg.OrganizationID,
		coredata.WebhookEventTypeDocumentReacceptanceRequired,
		reacceptanceRequiredPayload{
			SigningConfigID:    cfg.ID.String(),
			DocumentID:         cfg.DocumentID.String(),
			NewVersionID:       newVersion.ID.String(),
			NewVersionMajor:    newVersion.Major,
			NewVersionMinor:    newVersion.Minor,
			ReacceptancePolicy: cfg.ReacceptancePolicy.String(),
		},
	)
}
