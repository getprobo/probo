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

// Package agreement manages clickwrap agreement sites, agreements, subjects,
// and acceptances. It is the M1 data-plane service for the agreements REST API.
package agreement

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/validator"
	"go.probo.inc/probo/pkg/webhook"
	webhooktypes "go.probo.inc/probo/pkg/webhook/types"
)

var (
	// ErrSiteNotFound is returned when no site matches the provided secret key.
	ErrSiteNotFound = errors.New("agreement: site not found")
	// ErrAgreementNotFound is returned when an agreement ID is not known.
	ErrAgreementNotFound = errors.New("agreement: agreement not found")
	// ErrVersionNotPresentable is returned when a version has no PDF yet.
	ErrVersionNotPresentable = errors.New("agreement: version has no PDF; not yet presentable")
	// ErrContentHashMismatch is returned when the caller-supplied hash does not
	// match the stored content_sha256 for the version.
	ErrContentHashMismatch = errors.New("agreement: content hash mismatch")
	// ErrSubjectEmailConflict is returned when a subject already exists with a
	// different email for the same external_user_id.
	ErrSubjectEmailConflict = errors.New("agreement: subject already exists with a different email")
)

type Service struct {
	pg     *pg.Client
	esign  *esign.Service
	logger *log.Logger
}

func NewService(
	pgClient *pg.Client,
	esignSvc *esign.Service,
	logger *log.Logger,
) *Service {
	return &Service{
		pg:     pgClient,
		esign:  esignSvc,
		logger: logger,
	}
}

// AuthenticateSite looks up a site by its secret key and returns it. The key is
// never stored; only its SHA-256 hash is compared.
func (s *Service) AuthenticateSite(ctx context.Context, secretKey string) (*coredata.AgreementSite, error) {
	hashedKey := []byte(hash.SHA256Hex([]byte(secretKey)))

	var site coredata.AgreementSite
	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			return site.LoadByHashedSecretKey(ctx, conn, hashedKey)
		},
	); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrSiteNotFound
		}

		return nil, fmt.Errorf("cannot authenticate site: %w", err)
	}

	return &site, nil
}

// RecordAcceptanceRequest holds the validated input for one acceptance.
type RecordAcceptanceRequest struct {
	AgreementID       gid.GID
	ExternalUserID    string
	AccountID         *string
	Email             string
	IPAddress         *string
	UserAgent         *string
	SDKVersion        *string
	ConsentText       string
	DocumentVersionID gid.GID
	// ContentSHA256 is the hex SHA-256 of the PDF bytes, as returned by the
	// status API. If empty, Probo uses the stored value from the version record.
	ContentSHA256 string
}

func (r RecordAcceptanceRequest) Validate() error {
	v := validator.New()

	v.Check(r.ExternalUserID, "externalUserId", validator.NotEmpty())
	v.Check(r.Email, "email", validator.NotEmpty())
	v.Check(r.ConsentText, "consentText", validator.NotEmpty())

	return v.Error()
}

// RecordAcceptancesRequest is the atomic batch that records one or more
// acceptances in a single transaction.
type RecordAcceptancesRequest struct {
	Site        *coredata.AgreementSite
	Acceptances []RecordAcceptanceRequest
}

// RecordedAcceptance is the result of one successful acceptance recording.
type RecordedAcceptance struct {
	Acceptance *coredata.AgreementAcceptance
	Subject    *coredata.AgreementSubject
}

// RecordAcceptances records all acceptances atomically. Either all succeed or
// none are committed.
func (s *Service) RecordAcceptances(
	ctx context.Context,
	req *RecordAcceptancesRequest,
) ([]*RecordedAcceptance, error) {
	if len(req.Acceptances) == 0 {
		return nil, fmt.Errorf("at least one acceptance is required")
	}

	for i, a := range req.Acceptances {
		if err := a.Validate(); err != nil {
			return nil, fmt.Errorf("acceptance[%d]: %w", i, err)
		}
	}

	var results []*RecordedAcceptance

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			scope := coredata.NewScopeFromObjectID(req.Site.OrganizationID)

			for _, ar := range req.Acceptances {
				result, err := s.recordOne(ctx, tx, scope, req.Site, ar)
				if err != nil {
					return err
				}

				results = append(results, result)
			}

			for _, result := range results {
				if err := webhook.InsertData(
					ctx,
					tx,
					scope,
					req.Site.OrganizationID,
					coredata.WebhookEventTypeAgreementAccepted,
					webhooktypes.NewAgreementAcceptance(result.Acceptance, result.Subject),
				); err != nil {
					return fmt.Errorf("cannot enqueue agreement:accepted webhook: %w", err)
				}
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return results, nil
}

func (s *Service) recordOne(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	site *coredata.AgreementSite,
	req RecordAcceptanceRequest,
) (*RecordedAcceptance, error) {
	var agreement coredata.Agreement
	if err := agreement.LoadByID(ctx, tx, scope, req.AgreementID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, ErrAgreementNotFound
		}

		return nil, fmt.Errorf("cannot load agreement: %w", err)
	}

	var version coredata.DocumentVersion
	if err := version.LoadByID(ctx, tx, scope, req.DocumentVersionID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, fmt.Errorf("document version not found")
		}

		return nil, fmt.Errorf("cannot load document version: %w", err)
	}

	if version.FileID == nil || version.ContentSHA256 == nil {
		return nil, ErrVersionNotPresentable
	}

	storedHash := *version.ContentSHA256

	if req.ContentSHA256 != "" && !strings.EqualFold(req.ContentSHA256, storedHash) {
		return nil, ErrContentHashMismatch
	}

	email, err := mail.ParseAddr(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email: %w", err)
	}

	subject, err := s.upsertSubject(ctx, tx, scope, site, req, email)
	if err != nil {
		return nil, err
	}

	docType := coredata.ElectronicSignatureDocumentTypeFromDocumentType(version.DocumentType)

	sigReq := &esign.CreateAndAcceptSignatureRequest{
		OrganizationID: site.OrganizationID,
		DocumentType:   docType,
		FileID:         *version.FileID,
		SignerEmail:    email,
		SignerIPAddr:   stringOrEmpty(req.IPAddress),
		SignerUA:       stringOrEmpty(req.UserAgent),
		ConsentText:    req.ConsentText,
		Flow:           coredata.ElectronicSignatureFlowClickwrap,
	}

	sig, err := s.esign.CreateAndAcceptSignature(ctx, tx, sigReq)
	if err != nil {
		return nil, fmt.Errorf("cannot create clickwrap signature: %w", err)
	}

	now := time.Now()
	acceptance := &coredata.AgreementAcceptance{
		ID:                    gid.New(scope.GetTenantID(), coredata.AgreementAcceptanceEntityType),
		OrganizationID:        site.OrganizationID,
		SiteID:                site.ID,
		AgreementID:           agreement.ID,
		SubjectID:             subject.ID,
		DocumentVersionID:     version.ID,
		ContentSHA256:         storedHash,
		ElectronicSignatureID: sig.ID,
		IPAddress:             req.IPAddress,
		UserAgent:             req.UserAgent,
		SDKVersion:            req.SDKVersion,
		AcceptedAt:            now,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := acceptance.Insert(ctx, tx, scope); err != nil {
		return nil, fmt.Errorf("cannot insert agreement acceptance: %w", err)
	}

	return &RecordedAcceptance{Acceptance: acceptance, Subject: subject}, nil
}

func (s *Service) upsertSubject(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	site *coredata.AgreementSite,
	req RecordAcceptanceRequest,
	email mail.Addr,
) (*coredata.AgreementSubject, error) {
	var existing coredata.AgreementSubject

	err := existing.LoadBySiteAndExternalUserID(ctx, tx, site.ID, req.ExternalUserID)
	if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, fmt.Errorf("cannot load subject: %w", err)
	}

	if errors.Is(err, coredata.ErrResourceNotFound) {
		now := time.Now()
		subject := &coredata.AgreementSubject{
			ID:             gid.New(scope.GetTenantID(), coredata.AgreementSubjectEntityType),
			OrganizationID: site.OrganizationID,
			SiteID:         site.ID,
			ExternalUserID: req.ExternalUserID,
			AccountID:      req.AccountID,
			Email:          email.String(),
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		if err := subject.Insert(ctx, tx, scope); err != nil {
			return nil, fmt.Errorf("cannot insert subject: %w", err)
		}

		return subject, nil
	}

	// Update email if it changed (the customer's backend is authoritative).
	if !strings.EqualFold(existing.Email, email.String()) {
		existing.Email = email.String()
		existing.AccountID = req.AccountID
		existing.UpdatedAt = time.Now()

		if err := existing.Update(ctx, tx, scope); err != nil {
			return nil, fmt.Errorf("cannot update subject email: %w", err)
		}
	}

	return &existing, nil
}

// AgreementStatus represents a subject's acceptance status for one agreement.
type AgreementStatus struct {
	AgreementID        gid.GID
	CurrentVersionID   *gid.GID
	HasAccepted        bool
	AcceptedVersionID  *gid.GID
	RequiresAcceptance bool
}

// GetSubjectStatus returns the acceptance status of a subject for each of the
// given agreement IDs.
func (s *Service) GetSubjectStatus(
	ctx context.Context,
	site *coredata.AgreementSite,
	externalUserID string,
	agreementIDs []gid.GID,
) ([]*AgreementStatus, error) {
	var statuses []*AgreementStatus

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			scope := coredata.NewScopeFromObjectID(site.OrganizationID)

			var subject coredata.AgreementSubject
			err := subject.LoadBySiteAndExternalUserID(ctx, conn, site.ID, externalUserID)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load subject: %w", err)
			}

			subjectFound := !errors.Is(err, coredata.ErrResourceNotFound)

			for _, agreementID := range agreementIDs {
				var agreement coredata.Agreement
				if err := agreement.LoadByID(ctx, conn, scope, agreementID); err != nil {
					return fmt.Errorf("cannot load agreement %s: %w", agreementID, err)
				}

				status := &AgreementStatus{
					AgreementID:      agreement.ID,
					CurrentVersionID: agreement.CurrentVersionID,
				}

				if subjectFound && agreement.CurrentVersionID != nil {
					var latest coredata.AgreementAcceptance
					latestErr := latest.LoadLatestBySubjectAndAgreement(ctx, conn, subject.ID, agreement.ID)
					if latestErr != nil && !errors.Is(latestErr, coredata.ErrResourceNotFound) {
						return fmt.Errorf("cannot load latest acceptance: %w", latestErr)
					}

					if latestErr == nil {
						status.HasAccepted = true
						status.AcceptedVersionID = &latest.DocumentVersionID
						status.RequiresAcceptance = latest.DocumentVersionID != *agreement.CurrentVersionID
					} else {
						status.RequiresAcceptance = true
					}
				} else if agreement.CurrentVersionID != nil {
					status.RequiresAcceptance = true
				}

				statuses = append(statuses, status)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return statuses, nil
}

// CreateSite creates a new agreement site and returns the site plus its
// plaintext secret key (returned only once).
func (s *Service) CreateSite(
	ctx context.Context,
	organizationID gid.GID,
	name string,
	mode coredata.AgreementSiteMode,
) (*coredata.AgreementSite, string, error) {
	scope := coredata.NewScopeFromObjectID(organizationID)

	now := time.Now()
	site := &coredata.AgreementSite{
		ID:             gid.New(scope.GetTenantID(), coredata.AgreementSiteEntityType),
		OrganizationID: organizationID,
		Name:           name,
		Mode:           mode,
		AllowedOrigins: []string{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	secretKey, err := site.GenerateSecretKey()
	if err != nil {
		return nil, "", fmt.Errorf("cannot generate site key: %w", err)
	}

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			return site.Insert(ctx, tx, scope)
		},
	); err != nil {
		return nil, "", fmt.Errorf("cannot create agreement site: %w", err)
	}

	return site, secretKey, nil
}

func stringOrEmpty(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}
